package natsproxy

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
)

// Request wraps the HTTP request
// to be processed via pub/sub system.
type Request struct {
	URL        string
	Method     string
	Header     http.Header
	Form       url.Values
	RemoteAddr string
	Body       []byte
}

// UnmarshallFrom unmarshal the request from
// bytes, that usually come from proxy.
func (r *Request) UnmarshallFrom(requestData []byte) error {
	if err := json.Unmarshal(requestData, r); err != nil {
		return err
	}
	return nil
}

// NewRequestFromHTTP creates
// the Request struct from
// regular *http.Request by
// serialization of main parts of it.
func NewRequestFromHTTP(req *http.Request) (*Request, error) {
	var buf bytes.Buffer
	if req.Body != nil {
		if err := req.ParseForm(); err != nil {
			return nil, err
		}
		if _, err := buf.ReadFrom(req.Body); err != nil {
			return nil, err
		}
		if err := req.Body.Close(); err != nil {
			return nil, err
		}
	}

	request := Request{
		URL:        req.URL.String(),
		Method:     req.Method,
		Header:     req.Header,
		Form:       req.Form,
		RemoteAddr: req.RemoteAddr,
		Body:       buf.Bytes(),
	}
	return &request, nil
}
