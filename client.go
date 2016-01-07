package natsproxy

import (
	"encoding/json"
	"log"

	"github.com/nats-io/nats"
)

type NatsHandler func(res *Response, req *Request)

type Connector interface {
	Subscribe(url string, handler NatsHandler)
	UnSubscribe(url string, handler NatsHandler)
}

type NatsClient struct {
	conn *nats.Conn
}

func (nc *NatsClient) Subscribe(url string, handler NatsHandler) {
	nc.conn.Subscribe(url, func(m *nats.Msg) {
		request := &Request{}
		if err := request.UnmarshallFrom(m.Data); err != nil {
			log.Println(err)
			return
		}
		response := &Response{}
		handler(response, request)
		bytes, err := json.Marshal(response)
		if err != nil {
			log.Println(err)
			return
		}
		nc.conn.Publish(m.Reply, bytes)
	})
}
