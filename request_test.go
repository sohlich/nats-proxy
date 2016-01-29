package natsproxy

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
)

func TestUnmarshallFrom(t *testing.T) {
	original := &Request{
		URL:  "/api/method",
		Body: []byte{0xFF, 0xFC},
	}
	payload, _ := json.Marshal(original)

	copyObj := &Request{}
	if err := copyObj.UnmarshallFrom(payload); err != nil {
		t.Error(err)
	}

	if original.URL != copyObj.URL {
		t.Error("URL not equals")
	}

	// TODO more precise equality test
	if len(original.Body) != len(copyObj.Body) {
		t.Error("Body not equals")
	}
}

func TestNewRequestFromHttp(t *testing.T) {
	url, _ := url.Parse("http://test.com/test")
	httpReq := &http.Request{
		Method: "GET",
		URL:    url,
		Body:   ioutil.NopCloser(bytes.NewReader([]byte{0xFF, 0xFC})),
	}
	req, err := NewRequestFromHttp(httpReq)

	if err != nil {
		t.Error(err)
	}

	if req.URL != "http://test.com/test" {
		t.Error("Url not equals")
	}
	// TODO better test for Body
	if len(req.Body) != 2 {
		t.Error("Body length not equals")
	}
}
