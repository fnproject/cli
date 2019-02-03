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
	CallIDHeader           = "Fn-Call-Id"
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

func Invoke(provider provider.Provider, invokeURL string, content io.Reader, output io.Writer, method string, env []string, contentType string, includeCallID bool) error {

	method = "POST"

	// Read the request body (up to the maximum size), as this is used in the
	// authentication signature
	var req *http.Request
	if content != nil {
		b, err := ioutil.ReadAll(io.LimitReader(content, MaximumRequestBodySize))
		buffer := bytes.NewBuffer(b)
		req, err = http.NewRequest(method, invokeURL, buffer)
		if err != nil {
			return fmt.Errorf("Error creating request to service: %s", err)
		}
	} else {
		var err error
		req, err = http.NewRequest(method, invokeURL, nil)
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
		return fmt.Errorf("Error invoking function: %s", err)
	}
	defer resp.Body.Close()

	if logger.DebugEnabled() {
		b, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return err
		}
		fmt.Printf(string(b) + "\n")
	}

	if cid, ok := resp.Header[CallIDHeader]; ok && includeCallID {
		fmt.Fprint(output, fmt.Sprintf("Call ID: %v\n", cid[0]))
	}

	var body io.Reader = resp.Body
	if resp.StatusCode >= 400 {
		// if we don't get json, we need to buffer the input so that we can
		// display the user's function output as it was...
		var b bytes.Buffer
		body = io.TeeReader(resp.Body, &b)

		var msg apiErr
		err = json.NewDecoder(body).Decode(&msg)
		if err == nil && msg.Message != "" {
			// this is likely from fn, so unravel this...
			return fmt.Errorf("Error invoking function. status: %v message: %v", resp.StatusCode, msg.Message)
		}

		// read anything written to buffer first, then copy out rest of body
		body = io.MultiReader(&b, resp.Body)
	}

	// at this point, it's not an fn error, so output function output as is
	// TODO we should give users the option to see a status code too, like call id?

	lcc := lastCharChecker{reader: body}
	body = &lcc
	io.Copy(output, body)

	// #1408 - flush stdout
	if lcc.last != '\n' {
		fmt.Fprintln(output)
	}
	return nil
}

// lastCharChecker wraps an io.Reader to return the last read character
type lastCharChecker struct {
	reader io.Reader
	last   byte
}

func (l *lastCharChecker) Read(b []byte) (int, error) {
	n, err := l.reader.Read(b)
	if n > 0 {
		l.last = b[n-1]
	}
	return n, err
}
