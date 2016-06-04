package natsproxy

import (
	"log"
	"time"

	"encoding/json"

	"github.com/gogo/protobuf/proto"
	"github.com/nats-io/nats"
)

const (
	// GET method constant
	GET = "GET"
	// POST method constant
	POST = "POST"
	// PUT method constant
	PUT = "PUT"
	// DELETE method constant
	DELETE = "DELETE"
)

// NatsHandler handles the
// tranforrmed HTTP request from
// NatsProxy. The context c wraps the
// request and response.
type NatsHandler func(c *Context)

// NatsHandlers is an
// array of NatsHandler functions.
// This type primary function is to group
// filters in NatsClient.
type NatsHandlers []NatsHandler

// Connector is the interface for
// generic pub/sub client.
type Connector interface {
	Subscribe(url string, handler NatsHandler)
	UnSubscribe(url string, handler NatsHandler)
}

// NatsClient serves as Connector
// to NATS messaging. Allows to subscribe
// for an specific url or url pattern.
type NatsClient struct {
	conn    *nats.Conn
	filters NatsHandlers
	reqPool RequestPool
	resPool ResponsePool
}

// NewNatsClient creates new NATS client
// from given connection. The connection must be
// connected or the function will
// return error ErrNatsClientNotConnected.
func NewNatsClient(conn *nats.Conn) (*NatsClient, error) {
	if err := testConnection(conn); err != nil {
		return nil, err
	}
	return &NatsClient{
		conn,
		make([]NatsHandler, 0),
		NewRequestPool(),
		NewResponsePool(),
	}, nil
}

// Use will add the middleware NatsHandler
// for a client.
func (nc *NatsClient) Use(middleware NatsHandler) {
	nc.filters = append(nc.filters, middleware)
}

// GET subscribes the client
// for an url with GET method.
func (nc *NatsClient) GET(url string, handler NatsHandler) {
	nc.Subscribe(GET, url, handler)
}

// POST subscribes the client
// for an url with POST method.
func (nc *NatsClient) POST(url string, handler NatsHandler) {
	nc.Subscribe(POST, url, handler)
}

// PUT subscribes the client
// for an url with PUT method.
func (nc *NatsClient) PUT(url string, handler NatsHandler) {
	nc.Subscribe(PUT, url, handler)
}

// DELETE subscribes the client
// for an url with DELETE method.
func (nc *NatsClient) DELETE(url string, handler NatsHandler) {
	nc.Subscribe(DELETE, url, handler)
}

// Subscribe is a generic subscribe function
// for any http method. It also
// wraps the processing of the context.
func (nc *NatsClient) Subscribe(method, url string, handler NatsHandler) {
	subscribeURL := SubscribeURLToNats(method, url)
	nc.conn.Subscribe(subscribeURL, func(m *nats.Msg) {
		request := nc.reqPool.GetRequest()
		defer nc.reqPool.Put(request)
		if err := request.UnmarshallFrom(m.Data); err != nil {
			log.Println(err)
			return
		}
		response := nc.resPool.GetResponse()
		defer nc.resPool.Put(response)
		c := newContext(url, response, request)

		// Iterate through filters
		for _, filter := range nc.filters {
			filter(c)
			c.index++
		}

		// If request is aborted do
		// not proceed to handler.
		if !c.IsAborted() {
			handler(c)
		}
		bytes, err := proto.Marshal(c.Response)
		if err != nil {
			log.Println(err)
			return
		}
		nc.conn.Publish(m.Reply, bytes)
	})
}

// HandleWebsocket subscribes the
// handler for specific websocketID.
// The method adds the specific prefix
// for client to proxy communication.
func (nc *NatsClient) HandleWebsocket(webSocketID string, handler nats.MsgHandler) {
	nc.conn.Subscribe(ws_IN+webSocketID, handler)
}

// WriteWebsocketJSON writes struct
// serialized to JSON to registered
// websocketID NATS subject.
func (nc *NatsClient) WriteWebsocketJSON(websocketID string, msg interface{}) error {
	if data, err := json.Marshal(msg); err == nil {
		return nc.WriteWebsocket(websocketID, data)
	} else {
		return err
	}
}

// WriteWebsocket writes given bytes
// to given websocket subject.
func (nc *NatsClient) WriteWebsocket(websocketID string, data []byte) error {
	return nc.conn.Publish(ws_OUT+websocketID, data)
}

func (nc *NatsClient) SendGET(url string, req *Request) (response *Response, err error) {
	return nc.Send(GET, url, req)
}

func (nc *NatsClient) SendPOST(url string, req *Request) (response *Response, err error) {
	return nc.Send(POST, url, req)
}

func (nc *NatsClient) SendDELETE(url string, req *Request) (response *Response, err error) {
	return nc.Send(DELETE, url, req)
}

func (nc *NatsClient) SendPUT(url string, req *Request) (response *Response, err error) {
	return nc.Send(PUT, url, req)
}

func (nc *NatsClient) Send(method string, url string, req *Request) (response *Response, err error) {
	subject := SubscribeURLToNats(method, url)
	response, err = nc.requestResponse(subject, req)
	return
}

func (nc *NatsClient) requestResponse(subj string, req *Request) (res *Response, err error) {
	res = &Response{}
	data, err := proto.Marshal(req)
	if err != nil {
		return
	}
	msg, err := nc.conn.Request(subj, data, time.Second)
	if err != nil {
		return
	}
	err = res.ReadFrom(msg.Data)
	return
}
