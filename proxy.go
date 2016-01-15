package natsproxy

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/nats-io/nats"
)

type NatsProxy struct {
	conn *nats.Conn
}

func NewNatsProxy(conn *nats.Conn) *NatsProxy {
	return &NatsProxy{
		conn,
	}
}

func (np *NatsProxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
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
	log.Printf("Sending to %s:%s", req.Method, strings.Replace(req.URL.Path, "/", ".", -1))
	msg, respErr := np.conn.Request(fmt.Sprintf("%s:%s", req.Method, strings.Replace(req.URL.Path, "/", ".", -1)),
		bytes,
		10*time.Second)
	if respErr != nil {
		http.Error(rw, "No response", http.StatusInternalServerError)
		return
	}
	response := NewResponse()
	if err := json.Unmarshal(msg.Data, response); err != nil {
		http.Error(rw, "Cannot deserialize response", http.StatusInternalServerError)
		return
	}
	log.Printf("Gettings response to %+v", response)
	copyHeader(response.Header, rw.Header())
	rw.Write(response.Body)
}

func copyHeader(src, dst http.Header) {
	for k, v := range src {
		for _, val := range v {
			dst.Add(k, val)
		}
	}
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
