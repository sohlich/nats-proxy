package natsproxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/nats-io/nats"
)

// NatsProxy serves as a proxy
// between gnats and http. It automatically
// translates the HTTP requests to nats
// messages. The url and method of the HTTP request
// serves as the name of the nats channel, where
// the message is sent.
type NatsProxy struct {
	conn *nats.Conn
}

var (
	ErrNatsClientNotConnected = fmt.Errorf("Client not connected")
)

// NewNatsProxy creates an
// initialized NatsProxy
func NewNatsProxy(conn *nats.Conn) (*NatsProxy, error) {
	if conn.Status() != nats.CONNECTED {
		return nil, ErrNatsClientNotConnected
	}
	return &NatsProxy{
		conn,
	}, nil
}

func (np *NatsProxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// Transform the HTTP request to
	// NATS proxy request.
	request, err := NewRequestFromHttp(req)
	if err != nil {
		http.Error(rw, "Cannot process request", http.StatusInternalServerError)
		return
	}

	// Serialize the request.
	reqBytes, err := json.Marshal(&request)
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
