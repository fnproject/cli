package client

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/fnproject/fn_go/client/version"
	"github.com/fnproject/fn_go/clientv2"
	"github.com/fnproject/fn_go/provider"
)

type invokeTestProvider struct{}

func (p *invokeTestProvider) APIClientv2() *clientv2.Fn { return nil }
func (p *invokeTestProvider) APIURL() *url.URL          { return nil }
func (p *invokeTestProvider) UnavailableResources() []provider.FnResourceType {
	return nil
}
func (p *invokeTestProvider) VersionClient() *version.Client { return nil }
func (p *invokeTestProvider) WrapCallTransport(rt http.RoundTripper) http.RoundTripper {
	return rt
}

func TestInvokeSetsFnInvokeTypeHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("fn-invoke-type"); got != "detached" {
			t.Fatalf("expected fn-invoke-type detached, got %q", got)
		}
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	resp, err := Invoke(&invokeTestProvider{}, InvokeRequest{
		URL:          server.URL,
		FnInvokeType: "detached",
		Content:      strings.NewReader("{}"),
	})
	if err != nil {
		t.Fatalf("Invoke() error = %v", err)
	}
	defer resp.Body.Close()
	_, _ = io.ReadAll(resp.Body)
}