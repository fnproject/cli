package client

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	"github.com/fnproject/fn_go/provider"
	"github.com/go-openapi/runtime/logger"
)

func EnvAsHeader(req *http.Request, selectedEnv []string) {
	detectedEnv := os.Environ()
	if len(selectedEnv) > 0 {
		detectedEnv = selectedEnv
	}

	for _, e := range detectedEnv {
		kv := strings.Split(e, "=")
		name := kv[0]
		req.Header.Set(name, os.Getenv(name))
	}
}

// InvokeRequest are the parameters provided to Invoke
type InvokeRequest struct {
	URL         string
	Content     io.Reader
	Env         []string
	ContentType string
	// TODO headers should be their real type?
}

// Invoke calls the fn invoke API
func Invoke(provider provider.Provider, ireq InvokeRequest) (*http.Response, error) {
	invokeURL := ireq.URL
	content := ireq.Content
	env := ireq.Env
	contentType := ireq.ContentType
	method := "POST"

	req, err := http.NewRequest(method, invokeURL, content)
	if err != nil {
		return nil, fmt.Errorf("Error creating request to service: %s", err)
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	} else {
		req.Header.Set("Content-Type", "text/plain")
	}

	if len(env) > 0 {
		EnvAsHeader(req, env)
	}

	transport := provider.WrapCallTransport(http.DefaultTransport)
	httpClient := http.Client{Transport: transport}

	if logger.DebugEnabled() {
		b, err := httputil.DumpRequestOut(req, content != nil)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error dumping req", err)
		}
		os.Stderr.Write(b)
		fmt.Fprintln(os.Stderr)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error invoking function: %s", err)
	}

	if logger.DebugEnabled() {
		b, err := httputil.DumpResponse(resp, true)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error dumping resp", err)
		}
		os.Stderr.Write(b)
		fmt.Fprintln(os.Stderr)
	}

	return resp, nil
}
