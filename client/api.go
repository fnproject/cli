package client

import (
	"fmt"
	"net/url"
	"os"

	fnclient "github.com/fnproject/fn_go/client"
	openapi "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/spf13/viper"
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
	apiURL := viper.GetString("api_url")
	return url.Parse(apiURL)
}

func GetTransportAndRegistry() (*openapi.Runtime, strfmt.Registry) {
	transport := openapi.New(Host(), "/v1", []string{"http"})
	if token := viper.GetString("token"); token != "" {
		transport.DefaultAuthentication = openapi.BearerToken(token)
	}
	return transport, strfmt.Default
}

func APIClient() *fnclient.Fn {
	transport := openapi.New(Host(), "/v1", []string{"http"})
	if token := viper.GetString("token"); token != "" {
		transport.DefaultAuthentication = openapi.BearerToken(token)
	}

	// create the API client, with the transport
	client := fnclient.New(GetTransportAndRegistry())

	return client
}
