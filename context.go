package natsproxy

type Context struct {
	request    *Request
	response   *Response
	index      int
	abortIndex int
}

func (c *Context) IsAborted() bool {
	return c.index >= c.abortIndex
}

func (c *Context) Abort() {
	c.abortIndex = c.index
}

func newContext(res *Response, req *Request) *Context {
	return &Context{
		req,
		res,
		0,
		1<<31 - 1,
	}
}
