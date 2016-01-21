package natsproxy

import (
	"bytes"
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

	reqBytes, err := json.Marshal(&request)
	if err != nil {
		http.Error(rw, "Cannot process request", http.StatusInternalServerError)
		return
	}
	log.Printf("Sending to %s:%s", req.Method, strings.Replace(req.URL.Path, "/", ".", -1))
	msg, respErr := np.conn.Request(fmt.Sprintf("%s:%s", req.Method, strings.Replace(req.URL.Path, "/", ".", -1)),
		reqBytes,
		10*time.Second)
	if respErr != nil {
		http.Error(rw, "No response", http.StatusInternalServerError)
		return
	}
	response := NewResponse()
	err = response.Decode(msg.Data)
	if err != nil {
		http.Error(rw, "Cannot deserialize response", http.StatusInternalServerError)
		return
	}
	writeResponse(rw, response)
}

func writeResponse(rw http.ResponseWriter, response *Response) {
	// Copy headers
	// from NATS response.
	copyHeader(response.Header, rw.Header())

	// Write the response code
	rw.WriteHeader(response.StatusCode)

	// Write the bytes of response
	// to a response writer.
	// TODO benchmark
	bytes.NewBuffer(response.Body).WriteTo(rw)
}

func copyHeader(src, dst http.Header) {
	for k, v := range src {
		for _, val := range v {
			dst.Add(k, val)
		}
	}
}
