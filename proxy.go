package natsproxy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/nats-io/nats"
)

type FilterChain []http.HandlerFunc

type NatsProxy struct {
	conn    *nats.Conn
	filters FilterChain
}

func NewNatsProxy(conn *nats.Conn) *NatsProxy {
	return &NatsProxy{
		conn,
		make([]http.HandlerFunc, 0),
	}
}

func (np *NatsProxy) Use(middleware http.HandlerFunc) {
	np.filters = append(np.filters, middleware)
}

func (np *NatsProxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	for _, filter := range np.filters {
		filter(rw, req)
	}

	response := Response{}
	request, err := NewRequestFromHttp(req)
	if err != nil {
		http.Error(rw, "Cannot process request", http.StatusInternalServerError)
		return
	}
	bytes, err := json.Marshal(&request)
	if err != nil {
		http.Error(rw, "Cannot process request", http.StatusInternalServerError)
		return
	}
	np.conn.Request(fmt.Sprintf("%s:%s", req.Method, req.URL.Path),
		bytes,
		200*time.Millisecond)
	rw.Write(response.Body)
}

// func NewProxyHandler(conn *nats.Conn) (http.HandlerFunc, error) {
// 	handler := func(rw http.ResponseWriter, req *http.Request) {
// 		response := Response{}
// 		request, err := NewRequestFromHttp(req)
// 		if err != nil {
// 			http.Error(rw, "Cannot process request", http.StatusInternalServerError)
// 			return
// 		}
// 		bytes, err := json.Marshal(&request)
// 		if err != nil {
// 			http.Error(rw, "Cannot process request", http.StatusInternalServerError)
// 			return
// 		}
// 		conn.Request(req.URL.Path, bytes, 10*time.Millisecond)
// 		rw.Write(response.Body)
// 	}
// 	return handler, nil
// }
