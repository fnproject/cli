package run

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

// CallRequestHTTP for the protocol that was used by the end user to call this function. We only have HTTP right now.
type CallRequestHTTP struct {
	Type       string      `json:"type"`
	Method     string      `json:"method"`
	RequestURL string      `json:"request_url"`
	Headers    http.Header `json:"headers"`
}

// CallResponseHTTP for the protocol that was used by the end user to call this function. We only have HTTP right now.
type CallResponseHTTP struct {
	StatusCode int         `json:"status_code,omitempty"`
	Headers    http.Header `json:"headers,omitempty"`
}

type jsonIn struct {
	CallID      string          `json:"call_id"`
	ContentType string          `json:"content_type"`
	Deadline    string          `json:"deadline"`
	Body        string          `json:"body"`
	Protocol    CallRequestHTTP `json:"protocol"`
}

// jsonOut the expected response from the function container
type jsonOut struct {
	Body        string            `json:"body"`
	ContentType string            `json:"content_type"`
	Protocol    *CallResponseHTTP `json:"protocol,omitempty"`
}

func createJSONInput(callID, contentType, deadline string, method string, requestURL string, stdin io.Reader) (string, error) {
	var err error
	input := []byte("")
	if stdin != nil {
		input, err = ioutil.ReadAll(stdin)
		if err != nil {
			return "", fmt.Errorf("Error reading from stdin: %v", err)
		}
	}

	var b strings.Builder
	enc := json.NewEncoder(&b)
	jin := &jsonIn{
		CallID:      callID,
		ContentType: contentType,
		Deadline:    deadline,
		Body:        string(input),
		Protocol: CallRequestHTTP{
			Method:     method,
			RequestURL: requestURL,
			Type:       "http",
			Headers: map[string][]string{
				"Content-Type": {contentType},
			},
		},
	}
	err = enc.Encode(jin)
	if err != nil {
		return "", fmt.Errorf("Error encoding json: %v", err)
	}
	body := b.String()
	return body, nil
}

// this extracts the body from the response
func stdoutJSON(stdout io.Writer) io.Writer {
	// pulls out the body and returns it to the command line
	// TODO: might want a flag to skip this and output the full json response?
	var err error
	pr, pw := io.Pipe()

	go func() {
		dec := json.NewDecoder(pr)
		for {
			jsout := &jsonOut{}
			err = dec.Decode(jsout)
			if err != nil {
				fmt.Println("Error decoding", err)
				return
			}
			stdout.Write([]byte(jsout.Body))
		}
	}()
	return pw
}
