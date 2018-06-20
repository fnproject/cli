package run

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

type CloudEvent struct {
	CloudEventsVersion string                 `json:"cloudEventsVersion"`
	EventID            string                 `json:"eventID"`
	Source             string                 `json:"source"`
	EventType          string                 `json:"eventType"`
	EventTypeVersion   string                 `json:"eventTypeVersion"`
	EventTime          time.Time              `json:"eventTime"` // TODO: ensure rfc3339 format
	SchemaURL          string                 `json:"schemaURL"`
	ContentType        string                 `json:"contentType"`
	Extensions         map[string]interface{} `json:"extensions"`
	Data               interface{}            `json:"data"` // from docs: the payload is encoded into a media format which is specified by the contentType attribute (e.g. application/json)
}

func createCloudEventInput(callID, contentType, deadline string, method string, requestURL string, stdin io.Reader) (string, error) {
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
	in := &CloudEvent{
		EventID:     callID,
		ContentType: contentType,
		Extensions: map[string]interface{}{
			"protocol": CallRequestHTTP{
				Method:     method,
				RequestURL: requestURL,
				Type:       "http",
				Headers: map[string][]string{
					"Content-Type": {contentType},
				},
			},
			"deadline": deadline,
		},
	}
	if len(input) == 0 {
		// nada
		// todo: should we leave as null, pass in empty string, omitempty or some default for the content type, eg: {} for json?
	} else if contentType == "application/json" {
		d := map[string]interface{}{}
		err = json.Unmarshal(input, &d)
		if err != nil {
			return "", fmt.Errorf("Error unmarshalling json input: %v", err)
		}
		in.Data = d
	} else {
		in.Data = string(input)
	}
	err = enc.Encode(in)
	if err != nil {
		return "", fmt.Errorf("Error encoding json: %v", err)
	}
	body := b.String()
	return body, nil
}

// this extracts the body from the response
func stdoutCloudEvent(stdout io.Writer) io.Writer {
	// pulls out the body and returns it to the command line
	// TODO: might want a flag to skip this and output the full json response?
	var err error
	pr, pw := io.Pipe()

	go func() {
		dec := json.NewDecoder(pr)
		for {
			jsout := &CloudEvent{}
			err = dec.Decode(jsout)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error decoding", err)
				return
			}
			if jsout.ContentType == "application/json" {
				d, err := json.Marshal(jsout.Data)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error marshalling function response 'data' to json. %v\n", err)
					return
				}
				stdout.Write(d)
			} else if jsout.ContentType == "text/plain" {
				stdout.Write([]byte(jsout.Data.(string)))
			} else {
				fmt.Fprintf(os.Stderr, "Error: Unknown content type: %v\n", jsout.ContentType)
				return
			}

		}
	}()
	return pw
}
