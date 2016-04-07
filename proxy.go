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
	"github.com/satori/go.uuid"
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
	conn     *nats.Conn
	hooks    map[string]hookGroup
	wsMapper *webSocketMapper
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
	}, nil
}

func (np *NatsProxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	isWebSock := IsWebSocketRequest(req)
	wsID := ""
	if isWebSock {
		log.Println(URLToNats(req.Method, req.URL.Path))
		wsID = uuid.NewV4().String()
	}

	// Transform the HTTP request to
	// NATS proxy request.
	request, err := NewRequestFromHTTP(req)
	request.WebSocketId = wsID
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
	var response *Response
	response, err = DecodeResponse(msg.Data)
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

	if isWebSock && response.DoUpgrade {
		header := http.Header{}
		copyHeader(response.Header, header)
		if conn, err := upgrader.Upgrade(rw, req, header); err == nil {
			log.Println("Subscribing")
			np.wsMapper.fromNats[wsID] = conn
			np.wsMapper.toNats[conn] = wsID
			np.conn.Subscribe("WS_OUT"+wsID, func(m *nats.Msg) {
				log.Println("Sending data to %s\n" + wsID)
				err = conn.WriteMessage(websocket.BinaryMessage, m.Data)
				if err != nil {
					log.Println("Error writing a message", err)
				}
			})
			go func() {
				for {
					log.Println("Running go func to read WS")
					if _, p, err := conn.ReadMessage(); err == nil {
						log.Printf("Reading data from %s\n", wsID)
						np.conn.Publish("WS_IN"+wsID, p)
					} else {
						//TODO finish
						log.Println(err)
						break
					}
				}
			}()
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
