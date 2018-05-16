package client

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/fnproject/cli/config"
	fnclient "github.com/fnproject/fn_go/client"
	openapi "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/spf13/viper"
)

func Host() string {
	hostURL := HostURL()
	return hostURL.Host
}

func HostURL() *url.URL {
	return hostURL(viper.GetString(config.EnvFnAPIURL))
}

func hostURL(urlStr string) *url.URL {
	if !strings.Contains(urlStr, "://") {
		urlStr = fmt.Sprint("http://", urlStr)
	}

	url, err := url.Parse(urlStr)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unparsable FN API Url: %s. Error: %s \n", urlStr, err)
		os.Exit(1)
	}

	if url.Port() == "" {
		if url.Scheme == "http" {
			url.Host = fmt.Sprint(url.Host, ":80")
		}
		if url.Scheme == "https" {
			url.Host = fmt.Sprint(url.Host, ":443")
		}
	}

	//maintain backwards compatibility with first version FN_API_URL env vars
	if url.Path == "" || url.Path == "/" {
		url.Path = "/v1"
	}

	return url
}

func defaultProvider(transport *openapi.Runtime) {
	if token := viper.GetString(config.EnvFnToken); token != "" {
		transport.DefaultAuthentication = openapi.BearerToken(token)
	}
}

func oracleProvider(transport *openapi.Runtime) (err error) {
	t, err := oracleTransport(transport.Transport)
	if err != nil {
		return
	}
	transport.Transport = t
	return
}

func oracleTransport(roundTripper http.RoundTripper) (http.RoundTripper, error) {

	keyID, pKey, err := OracleConfigFile()
	if err != nil {
		return nil, err
	}

	compartmentID := viper.GetString(oracleCompartmentID)

	if viper.GetBool(oracleDisableCerts) {
		roundTripper = InsecureRoundTripper(roundTripper)
	}

	return NewCompartmentIDRoundTripper(
		compartmentID,
		NewOCISigningRoundTripper(
			keyID,
			pKey,
			roundTripper)), nil
}

func GetTransportAndRegistry() (*openapi.Runtime, strfmt.Registry, error) {
	hostURL := HostURL()
	transport := openapi.New(hostURL.Host, hostURL.Path, []string{hostURL.Scheme})
	var err error
	switch viper.GetString(config.ContextProvider) {
	case "default":
		defaultProvider(transport)
	case "oracle":
		err = oracleProvider(transport)
	default:
		defaultProvider(transport)
	}

	return transport, strfmt.Default, err
}

func APIClient() (*fnclient.Fn, error) {
	transport, registry, err := GetTransportAndRegistry()
	if err != nil {
		return nil, err
	}
	return fnclient.New(transport, registry), nil
}
