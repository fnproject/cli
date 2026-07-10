package commands

import (
	"flag"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	cliClient "github.com/fnproject/cli/client"
	"github.com/fnproject/cli/common"
	"github.com/fnproject/cli/config"
	"github.com/fnproject/fn_go/client/version"
	"github.com/fnproject/fn_go/clientv2"
	"github.com/fnproject/fn_go/modelsv2"
	"github.com/fnproject/fn_go/provider"
	"github.com/fnproject/fn_go/provider/oracle"
	"github.com/spf13/viper"
	"github.com/urfave/cli"
)

type invokeCommandTestProvider struct {
	apiURL *url.URL
}

func (p *invokeCommandTestProvider) APIClientv2() *clientv2.Fn { return nil }
func (p *invokeCommandTestProvider) APIURL() *url.URL          { return p.apiURL }
func (p *invokeCommandTestProvider) UnavailableResources() []provider.FnResourceType {
	return nil
}
func (p *invokeCommandTestProvider) VersionClient() *version.Client { return nil }
func (p *invokeCommandTestProvider) WrapCallTransport(rt http.RoundTripper) http.RoundTripper {
	return rt
}

func TestInvokeUsesEndpointCacheAfterFirstResolution(t *testing.T) {
	restore := stubInvokeCommandDependencies(t)
	defer restore()

	cachePath := t.TempDir() + "/invoke-endpoint-cache.json"
	newInvokeEndpointCache = func() *common.InvokeEndpointCache {
		return common.NewInvokeEndpointCache(cachePath)
	}

	var appLookups int
	var fnLookups int
	getInvokeAppByName = func(_ *clientv2.Fn, appName string) (*modelsv2.App, error) {
		appLookups++
		return &modelsv2.App{ID: "app-id", Name: appName}, nil
	}
	getInvokeFnByName = func(_ *clientv2.Fn, appID, fnName string) (*modelsv2.Fn, error) {
		fnLookups++
		return &modelsv2.Fn{
			ID:   "fn-id",
			Name: fnName,
			Annotations: map[string]interface{}{
				FnInvokeEndpointAnnotation: "https://invoke.example.com/fn",
			},
		}, nil
	}

	var invokedURLs []string
	invokeFunction = func(_ provider.Provider, req cliClient.InvokeRequest) (*http.Response, error) {
		invokedURLs = append(invokedURLs, req.URL)
		return invokeResponse(), nil
	}

	cl := invokeCmd{provider: testInvokeProvider(t)}
	for i := 0; i < 2; i++ {
		if err := cl.invoke(newInvokeCLIContext(t, "app", "fn"), ""); err != nil {
			t.Fatal(err)
		}
	}

	if appLookups != 1 {
		t.Fatalf("expected one app lookup, got %d", appLookups)
	}
	if fnLookups != 1 {
		t.Fatalf("expected one function lookup, got %d", fnLookups)
	}
	if len(invokedURLs) != 2 {
		t.Fatalf("expected two invoke calls, got %d", len(invokedURLs))
	}
	for _, got := range invokedURLs {
		if got != "https://invoke.example.com/fn" {
			t.Fatalf("unexpected invoked URL %q", got)
		}
	}
}

func TestInvokeEndpointFlagBypassesEndpointCache(t *testing.T) {
	restore := stubInvokeCommandDependencies(t)
	defer restore()

	getInvokeAppByName = func(_ *clientv2.Fn, appName string) (*modelsv2.App, error) {
		t.Fatalf("unexpected app lookup for %s", appName)
		return nil, nil
	}
	getInvokeFnByName = func(_ *clientv2.Fn, appID, fnName string) (*modelsv2.Fn, error) {
		t.Fatalf("unexpected function lookup for %s/%s", appID, fnName)
		return nil, nil
	}
	newInvokeEndpointCache = func() *common.InvokeEndpointCache {
		t.Fatal("unexpected cache access")
		return nil
	}

	var invokedURL string
	invokeFunction = func(_ provider.Provider, req cliClient.InvokeRequest) (*http.Response, error) {
		invokedURL = req.URL
		return invokeResponse(), nil
	}

	cl := invokeCmd{provider: testInvokeProvider(t)}
	if err := cl.invoke(newInvokeCLIContext(t, "--endpoint", "https://explicit.example.com/invoke"), ""); err != nil {
		t.Fatal(err)
	}

	if invokedURL != "https://explicit.example.com/invoke" {
		t.Fatalf("expected explicit endpoint, got %q", invokedURL)
	}
}

func TestInvokeNoEndpointCacheBypassesExistingCacheEntry(t *testing.T) {
	restore := stubInvokeCommandDependencies(t)
	defer restore()

	cachePath := t.TempDir() + "/invoke-endpoint-cache.json"
	newInvokeEndpointCache = func() *common.InvokeEndpointCache {
		return common.NewInvokeEndpointCache(cachePath)
	}
	key := common.NewInvokeEndpointCacheKey(testInvokeProvider(t), "app", "fn")
	if err := common.NewInvokeEndpointCache(cachePath).Put(key, "https://cached.example.com/fn"); err != nil {
		t.Fatal(err)
	}

	var appLookups int
	var fnLookups int
	getInvokeAppByName = func(_ *clientv2.Fn, appName string) (*modelsv2.App, error) {
		appLookups++
		return &modelsv2.App{ID: "app-id", Name: appName}, nil
	}
	getInvokeFnByName = func(_ *clientv2.Fn, appID, fnName string) (*modelsv2.Fn, error) {
		fnLookups++
		return &modelsv2.Fn{
			ID:   "fn-id",
			Name: fnName,
			Annotations: map[string]interface{}{
				FnInvokeEndpointAnnotation: "https://fresh.example.com/fn",
			},
		}, nil
	}

	var invokedURL string
	invokeFunction = func(_ provider.Provider, req cliClient.InvokeRequest) (*http.Response, error) {
		invokedURL = req.URL
		return invokeResponse(), nil
	}

	cl := invokeCmd{provider: testInvokeProvider(t)}
	if err := cl.invoke(newInvokeCLIContext(t, "--no-endpoint-cache", "app", "fn"), ""); err != nil {
		t.Fatal(err)
	}

	if appLookups != 1 || fnLookups != 1 {
		t.Fatalf("expected cache bypass to resolve function once, appLookups=%d fnLookups=%d", appLookups, fnLookups)
	}
	if invokedURL != "https://fresh.example.com/fn" {
		t.Fatalf("expected fresh endpoint, got %q", invokedURL)
	}
}

func stubInvokeCommandDependencies(t *testing.T) func() {
	t.Helper()
	viper.Reset()
	viper.Set(config.CurrentContext, "ctx")
	viper.Set(config.ContextProvider, "oracle")
	viper.Set(oracle.CfgCompartmentID, "ocid1.compartment.oc1..example")

	oldGetAppByName := getInvokeAppByName
	oldGetFnByName := getInvokeFnByName
	oldInvokeFunction := invokeFunction
	oldNewInvokeEndpointCache := newInvokeEndpointCache

	return func() {
		getInvokeAppByName = oldGetAppByName
		getInvokeFnByName = oldGetFnByName
		invokeFunction = oldInvokeFunction
		newInvokeEndpointCache = oldNewInvokeEndpointCache
		viper.Reset()
	}
}

func newInvokeCLIContext(t *testing.T, args ...string) *cli.Context {
	t.Helper()
	cmd := InvokeCommand()
	fs := flag.NewFlagSet("invoke-test", flag.ContinueOnError)
	for _, f := range cmd.Flags {
		f.Apply(fs)
	}
	if err := fs.Parse(args); err != nil {
		t.Fatal(err)
	}
	return cli.NewContext(cli.NewApp(), fs, nil)
}

func testInvokeProvider(t *testing.T) provider.Provider {
	t.Helper()
	apiURL, err := url.Parse("https://functions.example.com")
	if err != nil {
		t.Fatal(err)
	}
	return &invokeCommandTestProvider{apiURL: apiURL}
}

func invokeResponse() *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader("ok\n")),
	}
}
