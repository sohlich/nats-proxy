package natsproxy

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/nats-io/nats"
)

func TestProxy(t *testing.T) {

	clientConn, _ := nats.Connect(nats.DefaultURL)
	natsClient, _ := NewNatsClient(clientConn)
	natsClient.Subscribe("POST", "/test/:event/:session", Handler)
	defer clientConn.Close()

	proxyConn, _ := nats.Connect(nats.DefaultURL)
	proxyHandler, _ := NewNatsProxy(proxyConn)
	http.Handle("/", proxyHandler)
	defer proxyConn.Close()

	log.Println("initializing proxy")
	go http.ListenAndServe(":3000", nil)
	time.Sleep(1 * time.Second)

	log.Println("Posting request")
	reader := bytes.NewReader([]byte("testData"))
	resp, err := http.Post("http://127.0.0.1:3000/test/12324/123?name=testname", "multipart/form-data", reader)
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

func Handler(c *Context) {
	log.Println("Getting request")
	log.Println(c.Request.URL)
	c.Request.Form.Get("email")

	respStruct := struct {
		User string
	}{
		"Radek",
	}

	c.JSON(200, respStruct)
	c.Response.Header.Add("X-AUTH", "12345")
}
