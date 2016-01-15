package natsproxy

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/nats-io/nats"
)

const (
	GET    = "GET"
	POST   = "POST"
	PUT    = "PUT"
	DELETE = "DELETE"
)

type NatsHandler func(c *Context)

type NatsHandlers []NatsHandler

type Connector interface {
	Subscribe(url string, handler NatsHandler)
	UnSubscribe(url string, handler NatsHandler)
}

type NatsClient struct {
	conn    *nats.Conn
	filters NatsHandlers
}

func NewNatsClient(conn *nats.Conn) *NatsClient {
	return &NatsClient{
		conn,
		make([]NatsHandler, 0),
	}
}

func (np *NatsClient) Use(middleware NatsHandler) {
	np.filters = append(np.filters, middleware)
}

func (nc *NatsClient) GET(url string, handler NatsHandler) {
	nc.Subscribe(GET, url, handler)
}

func (nc *NatsClient) POST(url string, handler NatsHandler) {
	nc.Subscribe(POST, url, handler)
}

func (nc *NatsClient) PUT(url string, handler NatsHandler) {
	nc.Subscribe(PUT, url, handler)
}

func (nc *NatsClient) DELETE(url string, handler NatsHandler) {
	nc.Subscribe(DELETE, url, handler)
}

func (nc *NatsClient) Subscribe(method, url string, handler NatsHandler) {
	subscribeUrl := strings.Replace(url, "/", ".", -1)
	subscribeUrl = fmt.Sprintf("%s:%s", method, subscribeUrl)
	log.Printf("Subscribing to %s", subscribeUrl)
	nc.conn.Subscribe(subscribeUrl, func(m *nats.Msg) {
		log.Println("Received subscription")
		request := &Request{}
		if err := request.UnmarshallFrom(m.Data); err != nil {
			log.Println(err)
			return
		}
		response := NewResponse()
		c := newContext(response, request)

		// Iterate through filters
		for _, filter := range nc.filters {
			filter(c)
			c.index++
		}

		// If request
		// not proceed to handler
		if !c.IsAborted() {
			handler(c)
		}
		bytes, err := json.Marshal(c.Response)
		if err != nil {
			log.Println(err)
			return
		}
		nc.conn.Publish(m.Reply, bytes)
	})
}
