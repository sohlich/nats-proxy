package natsproxy

import (
	"bytes"
	"net/http"

	"github.com/gogo/protobuf/proto"
)

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
	var buf bytes.Buffer
	if req.Body != nil {
		if _, err := buf.ReadFrom(req.Body); err != nil {
			return nil, err
		}
		if err := req.Body.Close(); err != nil {
			return nil, err
		}
	}

	URL := req.URL.String()

	headerMap := HeaderMap{
		Items: make([]*HeaderItem, len(req.Header)),
	}

	// TODO simplify
	index := 0
	for k, v := range req.Header {
		headerMap.Items[index] = &HeaderItem{
			Key:   &k,
			Value: v,
		}
		index++
	}

	formMap := FormMap{
		Items: make([]*FormItem, len(req.Form)),
	}

	index = 0
	for k, v := range req.Form {
		formMap.Items[index] = &FormItem{
			Key:   &k,
			Value: v,
		}
		index++
	}

	request := Request{
		URL:        &URL,
		Method:     &req.Method,
		Header:     &headerMap,
		Form:       &formMap,
		RemoteAddr: &req.RemoteAddr,
		Body:       buf.Bytes(),
	}
	return &request, nil
}
