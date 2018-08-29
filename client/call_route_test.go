package client

import (
	"net/http"
	"os"
	"testing"
)

func TestEnvAsHeader(t *testing.T) {
	const expectedValue = "v=v"
	os.Setenv("k", expectedValue)

	cases := [][]string{
		nil,
		[]string{},
		[]string{"k"},
	}
	for _, selectedEnv := range cases {
		req, _ := http.NewRequest("GET", "http://www.example.com", nil)
		EnvAsHeader(req, selectedEnv)
		if found := req.Header.Get("k"); found != expectedValue {
			t.Errorf("Not found expected header: %v", found)
		}
	}
}
