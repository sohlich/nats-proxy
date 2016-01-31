package natsproxy

import (
	"encoding/json"
	"net/http"
)

// Response server as structure
// to transport http response
// throu NATS message queue
type Response struct {
	Header     http.Header
	StatusCode int
	Body       []byte
}

// NewResponse creates blank
// initialized Response object.
func NewResponse() *Response {
	return &Response{
		make(map[string][]string, 0),
		200,
		make([]byte, 0),
	}
}

// DecodeResponse decodes the
// marshalled Response struct
// back to struct.
func DecodeResponse(responseData []byte) (*Response, error) {
	r := &Response{}
	if err := json.Unmarshal(responseData, r); err != nil {
		return nil, err
	}
	return r, nil
}
