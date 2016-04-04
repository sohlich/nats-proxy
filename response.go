package natsproxy

import (
	"errors"

	"github.com/gogo/protobuf/proto"
)

// NewResponse creates blank
// initialized Response object.
func NewResponse() *Response {
	return &Response{
		StatusCode: int32(200),
		Header:     make(map[string]*Values, 0),
		Body:       make([]byte, 0),
	}
}

// DecodeResponse decodes the
// marshalled Response struct
// back to struct.
func DecodeResponse(responseData []byte) (*Response, error) {
	if responseData == nil || len(responseData) == 0 {
		return nil, errors.New("natsproxy: No response content found")
	}
	r := &Response{}
	if err := proto.Unmarshal(responseData, r); err != nil {
		return nil, err
	}
	return r, nil
}
