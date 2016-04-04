package natsproxy

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"mime"

	"net/url"
	"regexp"
	"strings"
)

var removeQueryRxp = regexp.MustCompile("[?]{1}.*")

// Context wraps the
// processed request/response
type Context struct {
	Request     *Request
	Response    *Response
	RequestForm url.Values
	index       int
	abortIndex  int
	params      map[string]int
}

// IsAborted returns true
// if the request in context
// were aborted by previous
// middleware
func (c *Context) IsAborted() bool {
	return c.index >= c.abortIndex
}

// Abort abortsthe
// request that it won's be
// processed further
func (c *Context) Abort() {
	c.abortIndex = c.index
}

// AbortWithJSON aborts the request
// and sets the HTTP status code to 500.
func (c *Context) AbortWithJSON(obj interface{}) {
	c.Abort()
	c.JSON(500, obj)
}

// BindJSON unmarshall the
// request body to given
// struct
func (c *Context) BindJSON(obj interface{}) error {
	if err := json.Unmarshal(c.Request.Body, obj); err != nil {
		return err
	}
	return nil
}

// JSON writes the serialized
// json to response
func (c *Context) JSON(statusCode int, obj interface{}) {
	c.Response.StatusCode = int32(statusCode)
	bytes, err := json.Marshal(obj)
	if err != nil {
		c.writeError(err)
	}
	c.Response.Body = bytes
}

// PathVariable returns
// the path variable
// based on its name (:xxx) defined
// in subscription URL
func (c *Context) PathVariable(name string) string {
	RawURL, err := url.Parse(c.Request.URL)
	if err != nil {
		return ""
	}
	URL := removeQueryRxp.ReplaceAllString(RawURL.Path, "")
	pathParams := strings.Split(URL, "/")

	index, ok := c.params[name]
	if !ok {
		return ""
	}
	if len(pathParams) <= index {
		return ""
	}
	return pathParams[index]
}

// FormVariable returns the
// variable from request form if
// available or empty string if not present.
func (c *Context) FormVariable(name string) string {
	arr, ok := c.Request.Form[name]
	if ok && len(arr.Arr) > 0 {
		return arr.Arr[0]
	}
	return ""
}

// HeaderVariable returns the header variable
// if avalable or empty string if header not present.
func (c *Context) HeaderVariable(name string) string {
	return getVal(name, c.Request.Header)
}

func getVal(name string, vals map[string]*Values) string {
	if val, ok := vals[name]; ok {
		if len(val.Arr) > 0 {
			return val.Arr[0]
		}
	}
	return ""
}

// ParseForm parses the request
// to values in RequestForm of the
// Context. The parsed form also includes
// the parameters from query and from body.
// Same as the http.Request, the post params
// are prior the query.
func (c *Context) ParseForm() error {
	var err error
	r := c.Request

	RawURL, err := url.Parse(r.URL)
	if err != nil {
		return err
	}

	// Parse the url query first
	queryForm := RawURL.Query()

	// Parse the post form
	var postFrom url.Values
	method := r.Method
	if method == "POST" || method == "PUT" || method == "PATCH" {
		postFrom, err = parseForm(r, c.HeaderVariable("Content-Type"))
	}

	// Merge form values
	// if post not empty
	var form url.Values
	if postFrom != nil {
		form = mergeValues(queryForm, postFrom)
	} else {
		form = queryForm
	}
	c.Request.Form = copyMap(form)
	return err
}

func parseForm(r *Request, header string) (url.Values, error) {
	var err error
	if r.Body == nil {
		return nil, errors.New("nats-proxy: missing request body")
	}
	ct := header

	// RFC 2616, section 7.2.1 - empty type
	// SHOULD be treated as application/octet-stream
	if ct == "" {
		ct = "application/octet-stream"
	}
	ct, _, err = mime.ParseMediaType(ct)
	if ct == "application/x-www-form-urlencoded" {
		var reader io.Reader = bytes.NewReader(r.Body)
		// 10 MB size limit
		maxFormSize := int64(10 << 20)
		reader = io.LimitReader(reader, maxFormSize+1)
		b, e := ioutil.ReadAll(reader)
		if e != nil {
			if err == nil {
				err = e
			}
			return nil, err
		}
		if int64(len(b)) > maxFormSize {
			return nil, errors.New("nats-proxy: POST too large")
		}
		vs, e := url.ParseQuery(string(b))
		if err == nil {
			err = e
		}
		return vs, nil
	}

	return nil, errors.New("nats-proxy: parseFrom multipart/form-data and others not supported")
}

// mergeValues the values
// with post values priority.
// So if the query param contains
// same param as the post
// form in body the param and value from
// request body will be included in parsed form.
func mergeValues(query, post url.Values) url.Values {
	merged := make(url.Values, 0)
	for key, val := range query {
		if len(val) > 0 {
			merged.Set(key, val[0])
		}
	}
	for key, val := range post {
		if len(val) > 0 {
			merged.Set(key, val[0])
		}
	}
	return merged
}

func (c *Context) writeError(err error) {
	c.Response.StatusCode = int32(500)
	c.Response.Body = []byte(err.Error())
}

func newContext(url string, res *Response, req *Request) *Context {
	m := buildParamMap(url)
	return &Context{
		req,
		res,
		nil,
		0,
		1<<31 - 1,
		m,
	}
}

func buildParamMap(url string) map[string]int {
	m := make(map[string]int)
	prmArr := strings.Split(url, "/")
	for i, prm := range prmArr {
		if len(prm) > 0 && prm[:1] == ":" {
			m[prm[1:]] = i
		}
	}
	return m
}
