package natsproxy

import "encoding/json"

type Context struct {
	Request    *Request
	Response   *Response
	index      int
	abortIndex int
}

func (c *Context) IsAborted() bool {
	return c.index >= c.abortIndex
}

func (c *Context) Abort() {
	c.abortIndex = c.index
}

func (c *Context) JSON(statusCode int, obj interface{}) {
	c.Response.StatusCode = statusCode
	bytes, err := json.Marshal(obj)
	if err != nil {
		c.writeError(err)
	}
	c.Response.Body = bytes

}

func (c *Context) writeError(err error) {
	c.Response.StatusCode = 500
	c.Response.Body = []byte(err.Error())
}

func newContext(res *Response, req *Request) *Context {
	return &Context{
		req,
		res,
		0,
		1<<31 - 1,
	}
}
