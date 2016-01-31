package natsproxy

import (
	"fmt"
	"testing"
)

func TestUrlReplace(t *testing.T) {
	path := "/home/:event/:session/:token"
	fmt.Println(path)
	res := SubscribeURLToNats("POST", path)
	fmt.Println(res)
}
