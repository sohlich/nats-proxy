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

func (r *Response) UnmarshallFrom(requestData []byte) (*Response, error) {
	if err := json.Unmarshal(requestData, r); err != nil {
		return nil, err
	}
	return r, nil
}
