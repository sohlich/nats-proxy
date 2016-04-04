package natsproxy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestPathVariable(t *testing.T) {
	url := "/test/1234/tst?name=testuser"
	req := &Request{
		URL: url,
	}
	resp := &Response{}
	ctx := newContext("/test/:token/:session", resp, req)

	tkn := ctx.PathVariable("token")
	if tkn != "1234" {
		t.Error("Defined path variable assertion failed")
	}

	session := ctx.PathVariable("session")
	if session != "tst" {
		t.Error("Defined path variable returned empty string")
	}

	unknwn := ctx.PathVariable("novalue")
	if unknwn != "" {
		t.Error("Non existing path variable returned non empty string")
	}

	unknwn = ctx.PathVariable("")
	if unknwn != "" {
		t.Error("Non existing path variable returned non empty string")
	}

	url = ""
	req.URL = url
	tkn = ctx.PathVariable("token")
	if tkn != "" {
		t.Error("No variable in URL.Path returned non emtpy token")
	}

}

func TestParseForm(t *testing.T) {
	url := "http://127.0.0.1:3000/test/12324/123?name=queryname"
	reader := strings.NewReader("z=post&both=y&prio=2&empty=&name=postname")
	req, _ := http.NewRequest("POST", "http://127.0.0.1:3000/test/12324/123?name=queryname", reader)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	req.Header.Set("X-AUTH", "xauthpayload")
	testRequest, _ := NewRequestFromHTTP(req)
	c := newContext(url, NewResponse(), testRequest)
	c.ParseForm()

	if c.FormVariable("name") != "postname" {
		fmt.Println("Got " + c.FormVariable("name"))
		t.Error("Form variable assertion failed")
	}
}

func TestParseFormNilBody(t *testing.T) {
	url := "http://127.0.0.1:3000/test/12324/123?name=queryname"
	// reader := strings.NewReader("z=post&both=y&prio=2&empty=&name=postname")
	req, _ := http.NewRequest("POST", "http://127.0.0.1:3000/test/12324/123?name=queryname", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	req.Header.Set("X-AUTH", "xauthpayload")
	testRequest, _ := NewRequestFromHTTP(req)
	c := newContext(url, NewResponse(), testRequest)
	if err := c.ParseForm(); err == nil {
		t.FailNow()
	}
}

func TestAbortContext(t *testing.T) {
	req := &Request{
		URL: "/test/1234/tst",
	}
	resp := &Response{}
	ctx := newContext("/test/:token/:session", resp, req)
	ctx.Abort()
	if ctx.IsAborted() != true {
		t.FailNow()
	}

}

func TestAbortJSONContext(t *testing.T) {

	req := &Request{
		URL: "/test/1234/tst",
	}
	resp := &Response{}
	ctx := newContext("/test/:token/:session", resp, req)
	ctx.AbortWithJSON("test")
	if ctx.IsAborted() != true {
		t.FailNow()
	}
	if exp, _ := json.Marshal("test"); string(exp) != string(resp.Body) {
		t.FailNow()
	}

}

type testStruct struct {
	Data string
}

func TestBindJson(t *testing.T) {
	dataStruct := testStruct{
		"Test",
	}

	data, _ := json.Marshal(dataStruct)
	req := &Request{
		URL:  "/test/1234/tst",
		Body: data,
	}
	resp := &Response{}
	ctx := newContext("/test/:token/:session", resp, req)

	verifStruct := &testStruct{}
	ctx.BindJSON(verifStruct)

	if verifStruct.Data != dataStruct.Data {
		t.FailNow()
	}
}
