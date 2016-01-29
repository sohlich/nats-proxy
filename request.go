package natsproxy

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
)

type Request struct {
	URL        string
	Method     string
	Header     http.Header
	Form       url.Values
	RemoteAddr string
	Body       []byte
}

func (r *Request) UnmarshallFrom(requestData []byte) error {
	if err := json.Unmarshal(requestData, r); err != nil {
		return err
	}
	return nil
}

func NewRequestFromHttp(req *http.Request) (*Request, error) {
	var buf bytes.Buffer
	if req.Body != nil {
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
