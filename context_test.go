package natsproxy

import (
	"encoding/json"
	"testing"
)

func TestPathVariable(t *testing.T) {
	req := &Request{
		URL: "/test/1234/tst",
	}
	resp := &Response{}
	ctx := newContext("/test/:token/:session", resp, req)

	tkn := ctx.PathVariable("token")
	if tkn != "1234" {
		t.Error("Defined path variable assertion failed")
	}

	session := ctx.PathVariable("session")
	if session != "tst" {
		t.Error("Defined path variable assertion failed")
	}

	unknwn := ctx.PathVariable("novalue")
	if unknwn != "" {
		t.Error("Non existing path variable returned non empty string")
	}

	unknwn = ctx.PathVariable("")
	if unknwn != "" {
		t.Error("Non existing path variable returned non empty string")
	}

	req.URL = ""
	tkn = ctx.PathVariable("token")
	if tkn != "" {
		t.Error("No variable in URL.Path returned non emtpy token")
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
