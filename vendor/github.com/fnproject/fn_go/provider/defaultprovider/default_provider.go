package defaultprovider

import (
	openapi "github.com/go-openapi/runtime/client"

	"net/http"
	"net/url"

	"path"

	"github.com/fnproject/fn_go/client/version"
	"github.com/fnproject/fn_go/clientv2"
	"github.com/fnproject/fn_go/provider"
	"github.com/go-openapi/strfmt"
)

// Provider is the default Auth provider
type Provider struct {
	// Optional token to add as  bearer token to auth calls
	Token string
	// API url to use for FN API interactions
	FnApiUrl *url.URL
}

func (dp *Provider) APIClientv2() *clientv2.Fn {
	transport := openapi.New(dp.FnApiUrl.Host, path.Join(dp.FnApiUrl.Path, clientv2.DefaultBasePath), []string{dp.FnApiUrl.Scheme})
	if dp.Token != "" {
		transport.DefaultAuthentication = openapi.BearerToken(dp.Token)
	}

	return clientv2.New(transport, strfmt.Default)
}

//  NewFromConfig creates a default provider  that does un-authenticated calls to
func NewFromConfig(configSource provider.ConfigSource, _ provider.PassPhraseSource) (provider.Provider, error) {

	apiUrl, err := provider.CanonicalFnAPIUrl(configSource.GetString(provider.CfgFnAPIURL))

	if err != nil {
		return nil, err
	}

	return &Provider{
		Token:    configSource.GetString(provider.CfgFnToken),
		FnApiUrl: apiUrl,
	}, nil
}

func (dp *Provider) WrapCallTransport(t http.RoundTripper) http.RoundTripper {
	return t
}

func (dp *Provider) UnavailableResources() []provider.FnResourceType {
	return []provider.FnResourceType{}
}

func (dp *Provider) APIURL() *url.URL {
	return dp.FnApiUrl
}

func (dp *Provider) APIClient() *clientv2.Fn {
	join := path.Join(dp.FnApiUrl.Path, clientv2.DefaultBasePath)
	transport := openapi.New(dp.FnApiUrl.Host, join, []string{dp.FnApiUrl.Scheme})
	if dp.Token != "" {
		transport.DefaultAuthentication = openapi.BearerToken(dp.Token)
	}

	return clientv2.New(transport, strfmt.Default)
}

func (op *Provider) VersionClient() *version.Client {
	runtime := openapi.New(op.FnApiUrl.Host, op.FnApiUrl.Path, []string{op.FnApiUrl.Scheme})
	runtime.Transport = op.WrapCallTransport(runtime.Transport)
	return version.New(runtime, strfmt.Default)
}
