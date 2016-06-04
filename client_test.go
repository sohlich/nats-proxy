package natsproxy

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/nats-io/nats"
)

var nats_url = getTestNatsUrl()

func getTestNatsUrl() string {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "127.0.0.1:4222"
	}
	return fmt.Sprintf("nats://%s", natsURL)
}

func TestGetSubscribe(t *testing.T) {
	clientConn, _ := nats.Connect(nats_url)
	natsClient, _ := NewNatsClient(clientConn)
	defer clientConn.Close()
	natsClient.GET("/test", func(c *Context) {
		c.JSON(200, "OK")
	})

	testClient, _ := nats.Connect(nats_url)
	defer testClient.Close()
	r := &Request{}
	data, _ := proto.Marshal(r)

	if _, err := testClient.Request("GET:.test", data, 10*time.Second); err != nil {
		t.Error("Did not get response")
	}
}

func TestPostSubscribe(t *testing.T) {

	clientConn, _ := nats.Connect(nats_url)
	natsClient, _ := NewNatsClient(clientConn)
	defer clientConn.Close()
	natsClient.POST("/test", func(c *Context) {
		c.JSON(200, "OK")
	})

	testClient, _ := nats.Connect(nats_url)
	defer testClient.Close()
	r := &Request{}
	data, _ := proto.Marshal(r)

	if _, err := testClient.Request("POST:.test", data, 10*time.Second); err != nil {
		t.Error("Did not get response")
	}
}

func TestPutSubscribe(t *testing.T) {
	clientConn, _ := nats.Connect(nats_url)
	natsClient, _ := NewNatsClient(clientConn)
	defer clientConn.Close()
	natsClient.PUT("/test", func(c *Context) {
		c.JSON(200, "OK")
	})

	testClient, _ := nats.Connect(nats_url)
	defer testClient.Close()
	r := &Request{}
	data, _ := proto.Marshal(r)

	if _, err := testClient.Request("PUT:.test", data, 10*time.Second); err != nil {
		t.Error("Did not get response")
	}
}

func TestDeleteSubscribe(t *testing.T) {
	clientConn, _ := nats.Connect(nats_url)
	natsClient, _ := NewNatsClient(clientConn)
	defer clientConn.Close()
	natsClient.DELETE("/test", func(c *Context) {
		c.JSON(200, "OK")
	})

	testClient, _ := nats.Connect(nats_url)
	defer testClient.Close()
	r := &Request{}
	data, _ := proto.Marshal(r)

	if _, err := testClient.Request("DELETE:.test", data, 10*time.Second); err != nil {
		t.Error("Did not get response")
	}
}
