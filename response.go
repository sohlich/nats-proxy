package natsproxy

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Header     http.Header
	StatusCode int
	Body       []byte
}

func NewResponse() *Response {
	return &Response{
		make(map[string][]string, 0),
		0,
		make([]byte, 0),
	}
}

func (r *Response) Decode(responseData []byte) error {
	if err := json.Unmarshal(responseData, r); err != nil {
		return err
	}
	return nil
}
