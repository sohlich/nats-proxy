package natsproxy

import (
	"errors"

	"github.com/gogo/protobuf/proto"
)

func (r *Response) GetHeader() Variables {
	if r != nil {
		return Variables(r.Header)
	}
	return nil
}

// NewResponse creates blank
// initialized Response object.
func NewResponse() *Response {
	return &Response{
		StatusCode: int32(200),
		Header:     make(map[string]*Values, 0),
		Body:       make([]byte, 0),
		DoUpgrade:  false,
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

func (res *Response) reset() {
	res.Header = make(map[string]*Values)
	res.Body = res.Body[0:0]
	res.DoUpgrade = false
	res.StatusCode = int32(0)
}
