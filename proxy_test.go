package natsproxy

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/nats-io/nats"
)

func TestProxy(t *testing.T) {

	var reqEvent string

	// Initialize NATS client
	//
	clientConn, _ := nats.Connect(nats_url)
	natsClient, _ := NewNatsClient(clientConn)
	natsClient.Subscribe("POST", "/test/:event/:session", func(c *Context) {
		reqEvent = c.PathVariable("event")

		if reqEvent != "12324" {
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
	http.Handle("/", proxyHandler)
	defer proxyConn.Close()

	log.Println("initializing proxy")
	go http.ListenAndServe(":3000", nil)
	time.Sleep(1 * time.Second)

	log.Println("Posting request")
	resp, err := http.PostForm("http://127.0.0.1:3000/test/12324/123?name=testname",
		url.Values{
			"post": []string{"postval"},
		})
	if err != nil {
		log.Println(err)
		t.Error("Cannot do post")
		return
	}

	out, _ := ioutil.ReadAll(resp.Body)
	respStruct := &struct {
		User string
	}{}

	json.Unmarshal(out, respStruct)
	log.Println(respStruct)
	if respStruct.User != "Radek" {
		t.Error("Response assertion failed")
	}

}
