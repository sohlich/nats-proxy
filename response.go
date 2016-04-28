package natsproxy

import (
	"errors"
	"sync"

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
func (r *Response) ReadFrom(responseData []byte) error {
	if responseData == nil || len(responseData) == 0 {
		return errors.New("natsproxy: No response content found")
	}
	if err := proto.Unmarshal(responseData, r); err != nil {
		return err
	}
	return nil
}

func (res *Response) reset() {
	res.Header = make(map[string]*Values)
	res.Body = res.Body[0:0]
	res.DoUpgrade = false
	res.StatusCode = int32(0)
}

type ResponsePool struct {
	sync.Pool
}

func (r *ResponsePool) GetResponse() *Response {
	res, _ := r.Get().(*Response)
	return res
}

func (r *ResponsePool) PutResponse(res *Response) {
	res.reset()
	r.Put(res)
}

func NewResponsePool() ResponsePool {
	return ResponsePool{
		sync.Pool{
			New: func() interface{} { return NewResponse() },
		}}
}
