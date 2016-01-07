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
	natsClient := NatsClient{
		clientConn,
	}
	natsClient.Subscribe("/test", Handler)

	proxyConn, _ := nats.Connect(nats.DefaultURL)
	proxyHandler := NewNatsProxy(proxyConn)
	http.Handle("/", proxyHandler)

	log.Println("initializing proxy")
	go http.ListenAndServe(":3000", nil)
	time.Sleep(10 * time.Second)

	log.Println("Posting request")
	reader := bytes.NewReader([]byte("testData"))
	resp, err := http.Post("http://localhost:3000/test", "multipart/form-data", reader)
	if err != nil {
		log.Println(err)
		t.Error("Cannot do post")
		return
	}

	out, _ := ioutil.ReadAll(resp.Body)
	log.Println(string(out))

}

func Handler(res *Response, req *Request) {
	log.Println("Getting request")
	log.Println(req.URL)
	req.Form.Get("email")

	respStruct := struct {
		User string
	}{
		"Radek",
	}

	bytes, _ := json.Marshal(respStruct)
	res.Body = bytes
}
