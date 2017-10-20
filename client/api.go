package client

import (
	"os"
	"net/url"

	fnclient "github.com/fnproject/fn_go/client"
	openapi "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

const (
	envFnToken = "FN_TOKEN"
)

func Host() string {
	h, err := HostURL()
	if err != nil {
		panic(err)
	}
	return h.Host
}

func HostURL() (*url.URL, error) {
	apiURL := os.Getenv("API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080"
	}

	return url.Parse(apiURL)
}

func GetTransportAndRegistry() (*openapi.Runtime, strfmt.Registry) {
	transport := openapi.New(Host(), "/v1", []string{"http"})
	if os.Getenv(envFnToken) != "" {
		transport.DefaultAuthentication = openapi.BearerToken(os.Getenv(envFnToken))
	}
	return transport, strfmt.Default
}

func APIClient() *fnclient.Fn {
	transport := openapi.New(Host(), "/v1", []string{"http"})
	if os.Getenv(envFnToken) != "" {
		transport.DefaultAuthentication = openapi.BearerToken(os.Getenv(envFnToken))
	}

	// create the API client, with the transport
	client := fnclient.New(GetTransportAndRegistry())

	return client
}
