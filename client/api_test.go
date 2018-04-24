package client

import (
	"testing"
)

func TestHostURL(t *testing.T) {
	var tests = []struct {
		input  string
		scheme string
		host   string
		port   string
		path   string
	}{
		{"http://localhost:8080/v1", "http", "localhost", "8080", "/v1"},
		{"http://localhost:8080", "http", "localhost", "8080", "/v1"},
		{"http://localhost", "http", "localhost", "80", "/v1"},
		{"localhost", "http", "localhost", "80", "/v1"},
		{"localhost:8080", "http", "localhost", "8080", "/v1"},
		{"localhost/v1", "http", "localhost", "80", "/v1"},
		{"localhost/", "http", "localhost", "80", "/v1"},
		{"https://localhost/v1", "https", "localhost", "443", "/v1"},
		{"https://someprovider/specificversion/withasubpath", "https", "someprovider", "443", "/specificversion/withasubpath"},
		{"https://someprovider:450/specificversion/withasubpath", "https", "someprovider", "450", "/specificversion/withasubpath"},
	}
	for _, test := range tests {
		url := hostURL(test.input)

		if url.Scheme != test.scheme {
			t.Errorf("Scheme not parsed: %s, %s", test.input, url.Scheme)
		}
		if url.Hostname() != test.host {
			t.Errorf("Host not parsed: %s, %s", test.input, url.Hostname())
		}
		if url.Port() != test.port {
			t.Errorf("Port not parsed: %s, %s", test.input, url.Port())
		}
		if url.Path != test.path {
			t.Errorf("Path not parsed: %s, %s", test.input, url.Path)
		}
	}
}
