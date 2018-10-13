package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	"github.com/fnproject/fn_go/provider"
	"github.com/go-openapi/runtime/logger"
)

const (
	FN_CALL_ID             = "Fn-Call-Id"
	MaximumRequestBodySize = 5 * 1024 * 1024 // bytes
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

type apiErr struct {
	Message string `json:"message"`
}

type callID struct {
	CallID string `json:"call_id"`
	Error  apiErr `json:"error"`
}

func Invoke(provider provider.Provider, invokeUrl string, content io.Reader, output io.Writer, method string, env []string, contentType string, includeCallID bool) error {

	method = "POST"

	// Read the request body (up to the maximum size), as this is used in the
	// authentication signature
	var req *http.Request
	if content != nil {
		b, err := ioutil.ReadAll(io.LimitReader(content, MaximumRequestBodySize))
		buffer := bytes.NewBuffer(b)
		req, err = http.NewRequest(method, invokeUrl, buffer)
		if err != nil {
			return fmt.Errorf("Error creating request to service: %s", err)
		}
	} else {
		var err error
		req, err = http.NewRequest(method, invokeUrl, nil)
		if err != nil {
			return fmt.Errorf("Error creating request to service: %s", err)
		}
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
			return err
		}
		fmt.Printf(string(b) + "\n")
	}

	resp, err := httpClient.Do(req)

	if err != nil {
		return fmt.Errorf("Error invoking fn: %s", err)
	}

	if logger.DebugEnabled() {
		b, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return err
		}
		fmt.Printf(string(b) + "\n")
	}

	// for sync calls
	if call_id, found := resp.Header[FN_CALL_ID]; found {
		if includeCallID {
			fmt.Fprint(os.Stderr, fmt.Sprintf("Call ID: %v\n", call_id[0]))
		}
		io.Copy(output, resp.Body)
	} else {
		// for async calls and error discovering
		c := &callID{}
		err = json.NewDecoder(resp.Body).Decode(c)
		if err == nil {
			// decode would not fail in both cases:
			// - call id in body
			// - error in body
			// that's why we need to check values of attributes
			if c.CallID != "" {
				fmt.Fprint(os.Stderr, fmt.Sprintf("Call ID: %v\n", c.CallID))
			} else {
				fmt.Fprint(output, fmt.Sprintf("Error: %v\n", c.Error.Message))
			}
		} else {
			return err
		}
	}

	if resp.StatusCode >= 400 {
		// TODO: parse out error message
		return fmt.Errorf("Error calling function: status %v", resp.StatusCode)
	}

	return nil
}
