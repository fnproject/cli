package provider

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/fnproject/fn_go/client/version"
	"github.com/fnproject/fn_go/clientv2"
)

// ProviderFunc constructs a provider
type ProviderFunc func(config ConfigSource, source PassPhraseSource) (Provider, error)

type FnResourceType string
const(
	ApplicationResourceType FnResourceType = "app"
	FunctionResourceType    FnResourceType = "function"
	TriggerResourceType     FnResourceType = "trigger"
)

//Providers describes a set of providers
type Providers struct {
	Providers map[string]ProviderFunc
}

func (t FnResourceType) String() string {
	return string(t)
}

// Register adds a named provider to a configuration
func (c *Providers) Register(name string, pf ProviderFunc) {
	if len(c.Providers) == 0 {
		c.Providers = make(map[string]ProviderFunc)
	}
	c.Providers[name] = pf
}

// Provider creates API clients for Fn calls adding any required middleware
type Provider interface {
	// APIURL returns the current API URL base to use with this provider
	APIURL() *url.URL
	// WrapCallTransport adds any request signing or auth to an existing round tripper for calls
	WrapCallTransport(http.RoundTripper) http.RoundTripper
	APIClientv2() *clientv2.Fn
	VersionClient() *version.Client
	// Returns a list of resource types that are not supported by this provider
	UnavailableResources() []FnResourceType
}

// CanonicalFnAPIUrl canonicalises an *FN_API_URL  to a default value
func CanonicalFnAPIUrl(urlStr string) (*url.URL, error) {
	if !strings.Contains(urlStr, "://") {
		urlStr = fmt.Sprint("http://", urlStr)
	}

	parseUrl, err := url.Parse(urlStr)

	if err != nil {
		return nil, fmt.Errorf("unparsable FN API Url: %s. Error: %s", urlStr, err)
	}

	if parseUrl.Port() == "" {
		if parseUrl.Scheme == "http" {
			parseUrl.Host = fmt.Sprint(parseUrl.Host, ":80")
		}
		if parseUrl.Scheme == "https" {
			parseUrl.Host = fmt.Sprint(parseUrl.Host, ":443")
		}
	}

	return parseUrl, nil
}

//ProviderFromConfig returns the provider corresponding to a given identifier populated with configuration from source - if a passphrase is required then it is request from phraseSource
func (c *Providers) ProviderFromConfig(id string, source ConfigSource, phraseSource PassPhraseSource) (Provider, error) {
	p, ok := c.Providers[id]
	if !ok {
		return nil, fmt.Errorf("No provider with id  '%s' is registered", id)
	}
	return p(source, phraseSource)
}
