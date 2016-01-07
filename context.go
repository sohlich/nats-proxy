package natsproxy

import "net/http"

type Context struct {
	request    *http.Request
	writer     http.ResponseWriter
	index      int
	abortIndex int
}

func newContext(rw http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		req,
		rw,
		0,
		1<<31 - 1,
	}
}
