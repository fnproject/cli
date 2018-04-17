package client

import (
	"fmt"
	"net/url"
	"os"

	"github.com/fnproject/cli/config"
	fnclient "github.com/fnproject/fn_go/client"
	openapi "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/spf13/viper"
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
	apiURL := viper.GetString(config.EnvFnAPIURL)
	return url.Parse(apiURL)
}

func GetTransportAndRegistry() (*openapi.Runtime, strfmt.Registry) {
	transport := openapi.New(Host(), "/v1", []string{"http"})

	switch viper.GetString(config.ContextProvider) {
	}
	if token := viper.GetString(config.EnvFnToken); token != "" {
		transport.DefaultAuthentication = openapi.BearerToken(token)
	}
	return transport, strfmt.Default
}

func APIClient() *fnclient.Fn {
	transport := openapi.New(Host(), "/v1", []string{"http"})
	if token := viper.GetString(config.EnvFnToken); token != "" {
		transport.DefaultAuthentication = openapi.BearerToken(token)
	}

	// create the API client, with the transport
	client := fnclient.New(GetTransportAndRegistry())

	return client
}
