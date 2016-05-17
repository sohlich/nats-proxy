package natsproxy

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"sync"

	"github.com/gogo/protobuf/proto"
	"github.com/nats-io/nuid"
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

func (r *Request) IsWebSocket() bool {
	return r.WebSocketID != ""
}

func (r *Request) GetWebSocketID() string {
	return r.WebSocketID
}

func (r *Request) FromHTTP(req *http.Request) error {

	if req == nil {
		return errors.New("natsproxy: Request cannot be nil")
	}

	isWebSock := IsWebSocketRequest(req)
	wsID := ""
	if isWebSock {
		wsID = nuid.Next()
	}

	buf := bytes.NewBuffer(r.Body)
	buf.Reset()
	if req.Body != nil {
		if _, err := io.Copy(buf, req.Body); err != nil {
			return err
		}
		if err := req.Body.Close(); err != nil {
			return err
		}
	}

	headerMap := copyMap(map[string][]string(req.Header))
	r.URL = req.URL.String()
	r.Method = req.Method
	r.Header = headerMap
	r.RemoteAddr = req.RemoteAddr
	r.WebSocketID = wsID
	r.Body = buf.Bytes()
	return nil
}

func NewRequest() *Request {
	return &Request{
		Header: make(map[string]*Values),
		Form:   make(map[string]*Values),
		Body:   make([]byte, 4096),
	}
}

func (req *Request) reset() {
	req.Header = make(map[string]*Values)
	req.Form = make(map[string]*Values)
	req.Method = req.Method[0:0]
	req.Body = req.Body[0:0]
	req.RemoteAddr = req.RemoteAddr[0:0]
	req.URL = req.URL[0:0]
}

type RequestPool struct {
	sync.Pool
}

func (r *RequestPool) GetRequest() *Request {
	request, _ := r.Get().(*Request)
	return request
}

func (r *RequestPool) PutRequest(req *Request) {
	req.reset()
	r.Put(req)
}

func NewRequestPool() RequestPool {
	return RequestPool{
		sync.Pool{
			New: func() interface{} { return NewRequest() },
		}}
}
