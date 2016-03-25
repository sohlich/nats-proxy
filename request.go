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
		// if err := req.ParseForm(); err != nil {
		// 	return nil, err
		// }
		if _, err := buf.ReadFrom(req.Body); err != nil {
			return nil, err
		}
		if err := req.Body.Close(); err != nil {
			return nil, err
		}
	}

	URL := req.URL.String()

	headerMap := copyMap(map[string][]string(req.Header))
	// formMap := copyMap(map[string][]string(req.Form))

	request := Request{
		URL:    &URL,
		Method: &req.Method,
		Header: headerMap,
		// Form:       formMap,
		RemoteAddr: &req.RemoteAddr,
		Body:       buf.Bytes(),
	}
	return &request, nil
}

// copy the values into protocol buffer
// struct
func copyMap(values map[string][]string) *Values {
	valueMap := Values{
		Items: make([]*Value, len(values)),
	}
	index := 0
	for k, v := range values {
		key := k // Needed to copy the adress of string
		valueMap.Items[index] = &Value{
			Key:   &key,
			Value: v,
		}
		index++
	}
	return &valueMap
}
