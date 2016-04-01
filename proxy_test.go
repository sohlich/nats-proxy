package natsproxy

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nats-io/nats"
)

// TestProxy integration test to
// test complete proxy and client.
func TestProxy(t *testing.T) {

	var reqEvent string

	// Initialize NATS client
	//
	clientConn, _ := nats.Connect(nats_url)
	natsClient, _ := NewNatsClient(clientConn)
	natsClient.Use(func(c *Context) {
		c.Response.Header.Set("Middleware", "True")
	})
	natsClient.Subscribe("POST", "/test/:event/:session", func(c *Context) {
		reqEvent = c.PathVariable("event")

		if reqEvent != "12324" {
			fmt.Println("ReqEvent: " + reqEvent)
			t.Error("Path variable doesn't match")
		}

		// Assert that the form
		// is also parsed for the
		// query params
		nameVal := c.Request.Form.Get("name")
		if nameVal != "testname" {
			t.Error("Form value assertion failed")
		}

		// Assets that the form params
		// are also parsed for post forms
		nameVal = c.Request.Form.Get("post")
		if nameVal != "postval" {
			fmt.Println("postval: " + nameVal)
			t.Error("Form value assertion failed")
		}

		respStruct := struct {
			User string
		}{
			"Radek",
		}

		c.JSON(200, respStruct)
		c.Response.Header.Add("X-AUTH", "12345")
	})
	defer clientConn.Close()

	proxyConn, _ := nats.Connect(nats_url)
	proxyHandler, _ := NewNatsProxy(proxyConn)
	defer proxyConn.Close()

	// go http.ListenAndServe(":3000", nil)
	// time.Sleep(1 * time.Second)

	reader := strings.NewReader("post=postval")
	req, _ := http.NewRequest("POST", "http://127.0.0.1:3000/test/12324/123?name=testname", reader)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	req.Header.Set("X-AUTH", "xauthpayload")

	rw := httptest.NewRecorder()
	proxyHandler.ServeHTTP(rw, req)

	// resp, err := http.PostForm("http://127.0.0.1:3000/test/12324/123?name=testname",
	// 	url.Values{
	// 		"post": []string{"postval"},
	// 	})
	// if err != nil {
	// 	log.Println(err)
	// 	t.Error("Cannot do post")
	// 	return
	// }

	if rw.Header().Get("Middleware") != "True" {
		t.Error("Middleware usage assertion failed")
	}

	out, _ := ioutil.ReadAll(rw.Body)
	respStruct := &struct {
		User string
	}{}

	json.Unmarshal(out, respStruct)
	log.Println(respStruct)
	if respStruct.User != "Radek" {
		t.Error("Response assertion failed")
	}

}

func TestProxyNilConnectionClient(t *testing.T) {
	//Should not be the IP of
	clientConn, _ := nats.Connect("127.0.0.0:6666")
	_, err := NewNatsProxy(clientConn)

	if err == nil {
		t.FailNow()
	}

}

func TestProxyNotConnectedClient(t *testing.T) {
	//Should not be the IP of
	clientConn, _ := nats.Connect(nats_url)
	clientConn.Close()
	_, err := NewNatsProxy(clientConn)

	if err == nil {
		t.FailNow()
	}

}

func TestProxyServeHttpError(t *testing.T) {

	proxyConn, _ := nats.Connect(nats_url)
	proxyHandler, _ := NewNatsProxy(proxyConn)
	defer proxyConn.Close()
	rw := httptest.NewRecorder()
	proxyHandler.ServeHTTP(rw, nil)

	if rw.Code != http.StatusInternalServerError {
		t.Error()
	}

	req, _ := http.NewRequest("", "", nil)
	proxyHandler.ServeHTTP(rw, req)
	if rw.Code != http.StatusInternalServerError {
		t.Error()
	}

}
