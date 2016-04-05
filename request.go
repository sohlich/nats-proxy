package natsproxy

import (
	"bytes"
	"errors"
	"net/http"

	"github.com/gogo/protobuf/proto"
)

func (r *Request) GetHeader() Variables {
	if r != nil {
		return Variables(r.Header)
	}
	return nil
}

func (r *Request) GetForm() Variables {
	if r != nil {
		return Variables(r.Form)
	}
	return nil
}

// UnmarshallFrom unmarshal the request from
// bytes, that usually come from proxy.
func (r *Request) UnmarshallFrom(requestData []byte) error {
	if err := proto.Unmarshal(requestData, r); err != nil {
		return err
	}
	return nil
}

// NewRequestFromHTTP creates
// the Request struct from
// regular *http.Request by
// serialization of main parts of it.
func NewRequestFromHTTP(req *http.Request) (*Request, error) {
	if req == nil {
		return nil, errors.New("natsproxy: Request cannot be nil")
	}
	var buf bytes.Buffer
	if req.Body != nil {
		if _, err := buf.ReadFrom(req.Body); err != nil {
			return nil, err
		}
		if err := req.Body.Close(); err != nil {
			return nil, err
		}
	}

	headerMap := copyMap(map[string][]string(req.Header))
	request := Request{
		URL:        req.URL.String(),
		Method:     req.Method,
		Header:     headerMap,
		RemoteAddr: req.RemoteAddr,
		Body:       buf.Bytes(),
	}
	return &request, nil
}
