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

// copy the values into protocol buffer
// struct
func copyMap(values map[string][]string) map[string]*Values {
	headerMap := make(map[string]*Values, 0)
	for k, v := range values {
		headerMap[k] = &Values{
			v,
		}
	}
	return headerMap
}
