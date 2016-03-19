package natsproxy

import "github.com/gogo/protobuf/proto"

// NewResponse creates blank
// initialized Response object.
func NewResponse() *Response {
	status := int32(200)
	return &Response{
		StatusCode: &status,
		Header: &HeaderMap{
			Items: make([]*HeaderItem, 0),
		},
		Body: make([]byte, 0),
	}
}

// DecodeResponse decodes the
// marshalled Response struct
// back to struct.
func DecodeResponse(responseData []byte) (*Response, error) {
	r := &Response{}
	if err := proto.Unmarshal(responseData, r); err != nil {
		return nil, err
	}
	return r, nil
}
