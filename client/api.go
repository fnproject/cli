package client

import (
	"fmt"
	"net/url"
	"os"

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
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}
	return h.Host
}

func HostURL() (*url.URL, error) {
	apiURL := os.Getenv("FN_API_URL")

	if apiURL == "" {
		if os.Getenv("API_URL") != "" {
			fmt.Fprint(os.Stderr, "Error: API_URL is deprecated, please use FN_API_URL.")
			os.Exit(1)
		}
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
