package natsproxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
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
		if c.Request.Method != "POST" {
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

		if v, ok := c.Request.Header["X-Auth"]; ok {
			if len(v.Arr) == 0 || v.Arr[0] != "xauthpayload" {
				t.Error("Header assertion failed")
			}
		}

		c.ParseForm()
		formVal := c.FormVariable("both")

		if formVal != "y" {
			t.Error("Form assertion failed")
		}

		// Generate response
		c.JSON(200, respStruct)
		c.Response.Header = map[string]*Values{"X-Auth": &Values{[]string{"12345"}}}
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

// Creates a new file upload http request with optional extra params
func newfileUploadRequest(uri string, params map[string]string, paramName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, filepath.Base(path))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	return http.NewRequest("POST", uri, body)
}
