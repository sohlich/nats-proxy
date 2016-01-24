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
	natsClient := NewNatsClient(clientConn)
	natsClient.Subscribe("POST", "/test/:event/:session", Handler)
	// defer clientConn.Close()

	proxyConn, _ := nats.Connect(nats.DefaultURL)
	proxyHandler := NewNatsProxy(proxyConn)
	http.Handle("/", proxyHandler)
	// defer proxyConn.Close()

	log.Println("initializing proxy")
	go http.ListenAndServe(":3000", nil)
	time.Sleep(1 * time.Second)

	log.Println("Posting request")
	reader := bytes.NewReader([]byte("testData"))
	resp, err := http.Post("http://localhost:3000/test/12324/123", "multipart/form-data", reader)
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

	bytes, _ := json.Marshal(respStruct)
	c.Response.Body = bytes
	c.Response.Header.Add("X-AUTH", "12345")
}
