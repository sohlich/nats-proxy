package natsproxy

import "testing"

func TestDecodeResponseError(t *testing.T) {
	response := make([]byte, 5)
	_, err := DecodeResponse(response)
	if err == nil {
		t.Error("Bad content assertion failed")
	}

	var nilData []byte
	_, err = DecodeResponse(nilData)
	if err == nil {
		t.Error("Nil content assertion failed")
	}

}
