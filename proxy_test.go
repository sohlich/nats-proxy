package natsproxy

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/nats-io/nats"
)

func TestProxy(t *testing.T) {

	var reqEvent string
	var reqSession string

	clientConn, _ := nats.Connect(nats_url)
	natsClient, _ := NewNatsClient(clientConn)
	natsClient.Subscribe("POST", "/test/:event/:session", func(c *Context) {
		reqEvent = c.PathVariable("event")
		reqSession = c.PathVariable("session")

		// Assert method
		if c.Request.GetMethod() != "POST" {
			t.Error("Method assertion failed")
		}

		if reqEvent != "12324" {
			fmt.Println(reqEvent)
			t.Error("Event path variable assertion failed")
		}

		respStruct := struct {
			User string
		}{
			"Radek",
		}

		constainsXAuth := false
		for _, v := range c.Request.GetHeader().GetItems() {
			if v.GetKey() == "X-Auth" {
				constainsXAuth = true
			}
		}

		if !constainsXAuth {
			t.Error("Header assertion failed")
		}

		c.ParseForm()
		formVal := c.FormVariable("both")

		if formVal != "y" {
			t.Error("Form assertion failed")
		}

		// Generate response
		c.JSON(200, respStruct)
		headerKey := "X-Auth"
		c.Response.Header = &Values{
			Items: []*Value{
				&Value{
					Key:   &headerKey,
					Value: []string{"12345"},
				},
			},
		}
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
	client := &http.Client{}

	reader := strings.NewReader("z=post&both=y&prio=2&empty=")
	req, err := http.NewRequest("POST", "http://127.0.0.1:3000/test/12324/123?name=testname", reader)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	req.Header.Set("X-AUTH", "xauthpayload")

	resp, err := client.Do(req)
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

	if reqEvent != "12324" {
		t.Error("Path variable doesn't match")
	}
}
