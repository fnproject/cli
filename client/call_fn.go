package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/spf13/viper"

	"github.com/fnproject/cli/config"
)

const (
	FN_CALL_ID             = "Fn_call_id"
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

func CallHostURL() *url.URL {
	url := viper.GetString(config.EnvFnCallURL)
	if url == "" {
		url = viper.GetString(config.EnvFnAPIURL)
	}
	return hostURL(url)
}

func CallFN(appName string, route string, content io.Reader, output io.Writer, method string, env []string, contentType string, includeCallID bool) error {

	u := CallHostURL()
	u.Path = path.Join("r", appName, route)

	if method == "" {
		if content == nil {
			method = "GET"
		} else {
			method = "POST"
		}
	}

	// Read the request body (up to the maximum size), as this is used in the
	// authentication signature
	var req *http.Request
	if content != nil {
		b, err := ioutil.ReadAll(io.LimitReader(content, MaximumRequestBodySize))
		buffer := bytes.NewBuffer(b)
		req, err = http.NewRequest(method, u.String(), buffer)
		if err != nil {
			return fmt.Errorf("Error creating request to service: %s", err)
		}
	} else {
		var err error
		req, err = http.NewRequest(method, u.String(), nil)
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

	transport, err := getTransport()
	if err != nil {
		return err
	}
	httpClient := http.Client{Transport: transport}

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error running route: %s", err)
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
		return fmt.Errorf("error calling function: status %v", resp.StatusCode)
	}

	return nil
}

func getTransport() (http.RoundTripper, error) {
	switch viper.GetString(config.ContextProvider) {
	case "default":
		return http.DefaultTransport, nil
	case "oracle":
		return oracleTransport(http.DefaultTransport)
	default:
		return http.DefaultTransport, nil
	}
}
