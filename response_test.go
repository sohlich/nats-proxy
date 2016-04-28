package natsproxy

import "testing"

func TestDecodeResponseError(t *testing.T) {
	response := make([]byte, 5)
	r := NewResponse()
	err := r.ReadFrom(response)
	if err == nil {
		t.Error("Bad content assertion failed")
	}

	var nilData []byte
	err = r.ReadFrom(nilData)
	if err == nil {
		t.Error("Nil content assertion failed")
	}

}
