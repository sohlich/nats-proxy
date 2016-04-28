package natsproxy

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/gorilla/websocket"
	"github.com/nats-io/nats"
)

var (
	// ErrNatsClientNotConnected is returned
	// if the natsclient inserted
	// in NewNatsProxy is not connected.
	ErrNatsClientNotConnected = fmt.Errorf("Client not connected")
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// HookFunc is the function that is
// used to modify response just before its
// transformed to HTTP response
type HookFunc func(*Response)

type webSocketMapper struct {
	toNats   map[*websocket.Conn]string
	fromNats map[string]*websocket.Conn
}

// NatsProxy serves as a proxy
// between gnats and http. It automatically
// translates the HTTP requests to nats
// messages. The url and method of the HTTP request
// serves as the name of the nats channel, where
// the message is sent.
type NatsProxy struct {
	conn         *nats.Conn
	hooks        map[string]hookGroup
	wsMapper     *webSocketMapper
	requestPool  RequestPool
	responsePool ResponsePool
}

type hookGroup struct {
	regexp *regexp.Regexp
	hooks  []HookFunc
}

// NewNatsProxy creates an
// initialized NatsProxy
func NewNatsProxy(conn *nats.Conn) (*NatsProxy, error) {
	if err := testConnection(conn); err != nil {
		return nil, err
	}
	return &NatsProxy{
		conn,
		make(map[string]hookGroup, 0),
		&webSocketMapper{
			make(map[*websocket.Conn]string, 0),
			make(map[string]*websocket.Conn, 0),
		},
		NewRequestPool(),
		NewResponsePool(),
	}, nil
}

func (np *NatsProxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	// Transform the HTTP request to
	// NATS proxy request.
	request := np.requestPool.GetRequest()
	defer np.requestPool.Put(request)

	err := request.FromHTTP(req)
	if err != nil {
		http.Error(rw, "Cannot process request", http.StatusInternalServerError)
		return
	}

	// Serialize the request.
	reqBytes, err := proto.Marshal(request)
	if err != nil {
		http.Error(rw, "Cannot process request", http.StatusInternalServerError)
		return
	}

	// Post request to message queue
	msg, respErr := np.conn.Request(
		URLToNats(req.Method, req.URL.Path),
		reqBytes,
		10*time.Second)
	if respErr != nil {
		http.Error(rw, "No response", http.StatusInternalServerError)
		return
	}
	response := np.responsePool.GetResponse()
	err = response.ReadFrom(msg.Data)
	defer np.responsePool.Put(response)
	if err != nil {
		http.Error(rw, "Cannot deserialize response", http.StatusInternalServerError)
		return
	}

	// Apply hook if regex match
	for _, hG := range np.hooks {
		if hG.regexp.MatchString(req.URL.Path) {
			for _, hook := range hG.hooks {
				hook(response)
			}
		}
	}

	// If response contains
	// the permission to do ws upgrade, the
	// proxy side upgrades the connection and
	// provides the Web Socket proxiing on
	// provided wsID toppic (WS_IN+wsID as receiving
	// and WS_OUT+wsID as outcoming)
	if request.IsWebSocket() && response.DoUpgrade {
		header := http.Header{}
		copyHeader(response.Header, header)
		if conn, err := upgrader.Upgrade(rw, req, header); err == nil {
			np.activateWSProxySubject(conn, request.WebSocketID)
		} else {
			log.Println("natsproxy error: " + err.Error())
		}
	} else {
		writeResponse(rw, response)
	}

}

// AddHook add the hook to modify,
// process response just before
// its transformed to HTTP form.
func (np *NatsProxy) AddHook(urlRegex string, hook HookFunc) error {
	hG, ok := np.hooks[urlRegex]
	if !ok {
		regexp, err := regexp.Compile(urlRegex)
		if err != nil {
			return err
		}
		hooks := make([]HookFunc, 1)
		hooks[0] = hook
		np.hooks[urlRegex] = hookGroup{
			regexp,
			hooks,
		}
	} else {
		hG.hooks = append(hG.hooks, hook)
	}
	return nil
}

func (np *NatsProxy) activateWSProxySubject(conn *websocket.Conn, wsID string) {
	np.addToWSMapper(conn, wsID)
	np.conn.Subscribe("WS_OUT"+wsID, func(m *nats.Msg) {
		err := conn.WriteMessage(websocket.TextMessage, m.Data)
		if err != nil {
			log.Println("Error writing a message", err)
		}
	})
	go func() {
		for {
			if _, p, err := conn.ReadMessage(); err == nil {
				np.conn.Publish("WS_IN"+wsID, p)
			} else {
				np.removeFromWSMapper(conn, wsID)
				logWebsocketError(wsID, err)
				break
			}
		}
	}()
}

func (np *NatsProxy) addToWSMapper(conn *websocket.Conn, wsID string) {
	np.wsMapper.fromNats[wsID] = conn
	np.wsMapper.toNats[conn] = wsID
}

func (np *NatsProxy) removeFromWSMapper(conn *websocket.Conn, wsID string) {
	delete(np.wsMapper.fromNats, wsID)
	delete(np.wsMapper.toNats, conn)
}

func (np *NatsProxy) resetWSMapper() {
	np.wsMapper.fromNats = make(map[string]*websocket.Conn, 0)
	np.wsMapper.toNats = make(map[*websocket.Conn]string, 0)
}

func (np *NatsProxy) closeAllWebsockets() error {
	var outerError error
	closed := make([]string, 0)
	for key, val := range np.wsMapper.fromNats {
		if err := val.Close(); err != nil {
			outerError = fmt.Errorf("nats-proxy: closing websocket ID: %s caused error: %s", key, err.Error())
		}
		closed = append(closed, key)
	}
	if outerError == nil {
		np.resetWSMapper()
	} else {
		// Remove just closed
		for _, val := range closed {
			np.removeFromWSMapper(np.wsMapper.fromNats[val], val)
		}
	}

	return outerError
}

func logWebsocketError(wsID string, err error) {
	log.Printf("nats-proxy: underlying websocker ID: %s error: %s", wsID, err.Error())
}

func writeResponse(rw http.ResponseWriter, response *Response) {
	// Copy headers
	// from NATS response.
	copyHeader(response.Header, rw.Header())

	// Write the response code
	rw.WriteHeader(int(response.StatusCode))

	// Write the bytes of response
	// to a response writer.
	// TODO benchmark
	bytes.NewBuffer(response.Body).WriteTo(rw)
}

func copyHeader(src map[string]*Values, dst http.Header) {
	for key, it := range src {
		for _, val := range it.Arr {
			dst.Add(key, val)
		}
	}
}
