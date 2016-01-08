package natsproxy

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/nats-io/nats"
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

func (nc *NatsClient) Subscribe(method, url string, handler NatsHandler) {
	nc.conn.Subscribe(fmt.Sprintf("%s:%s", method, url), func(m *nats.Msg) {
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
		bytes, err := json.Marshal(c.response)
		if err != nil {
			log.Println(err)
			return
		}
		nc.conn.Publish(m.Reply, bytes)
	})
}
