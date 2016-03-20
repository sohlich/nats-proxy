package natsproxy

import "testing"

func TestSetGetHeader(t *testing.T) {

	request := &Request{
		Header: &Values{
			Items: make([]*Value, 0),
		},
	}

	request.Header.Set("test", "value")

	result := request.Header.Get("test")

	if result != "value" {
		t.Error("Value assertion failed")
	}

}
