package natsproxy

import (
	"fmt"
	"testing"

	"github.com/nats-io/nats"
)

func TestUrlReplace(t *testing.T) {
	path := "/home/:event/:session/:token"
	res := SubscribeURLToNats("POST", path)
	if res != "POST:.home.*.*.*" {
		fmt.Println(res)
		t.FailNow()
	}
}

func TestTestConnection(t *testing.T) {

	clientConn, _ := nats.Connect(nats_url)
	clientConn.Close()
	if err := testConnection(clientConn); err == nil {
		t.Error("closed NATS connection assertion failed ")
	}

	if err := testConnection(nil); err == nil {
		t.Error("nil NATS connection assertion failed ")
	}

}
