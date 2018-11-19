package provider

import (
	"testing"
)

func TestCanonicaliseUrlURL(t *testing.T) {
	var tests = []struct {
		input  string
		scheme string
		host   string
		port   string
		path   string
	}{
		{"http://localhost:8080/v1", "http", "localhost", "8080", "/"},
		{"http://localhost:8080", "http", "localhost", "8080", ""},
		{"http://localhost", "http", "localhost", "80", ""},
		{"localhost", "http", "localhost", "80", ""},
		{"localhost:8080", "http", "localhost", "8080", ""},
		{"localhost/v1", "http", "localhost", "80", "/"},
		{"localhost/", "http", "localhost", "80", "/"},
		{"https://localhost/v1", "https", "localhost", "443", "/"},
		{"https://someprovider/specificversion/withasubpath", "https", "someprovider", "443", "/specificversion/withasubpath"},
		{"https://someprovider:450/specificversion/withasubpath", "https", "someprovider", "450", "/specificversion/withasubpath"},
	}
	for _, test := range tests {
		url, err := CanonicalFnAPIUrl(test.input)

		if err != nil {
			t.Fatalf("Failed to parse input URL %s %s", test.input, err)
		}

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
