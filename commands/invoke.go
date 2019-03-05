package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"errors"

	"github.com/fnproject/cli/client"
	"github.com/fnproject/cli/common"
	"github.com/fnproject/cli/objects/app"
	"github.com/fnproject/cli/objects/fn"
	"github.com/fnproject/fn_go/clientv2"
	"github.com/fnproject/fn_go/provider"
	"github.com/urfave/cli"
)

// FnInvokeEndpointAnnotation is the annotation that exposes the fn invoke endpoint as defined in models/fn.go
const (
	FnInvokeEndpointAnnotation = "fnproject.io/fn/invokeEndpoint"
	CallIDHeader               = "Fn-Call-Id"
)

type invokeCmd struct {
	provider provider.Provider
	client   *clientv2.Fn
}

// InvokeFnFlags used to invoke and fn
var InvokeFnFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "endpoint",
		Usage: "Specify the function invoke endpoint for this function, the app-name and func-name parameters will be ignored",
	},
	cli.StringFlag{
		Name:  "content-type",
		Usage: "The payload Content-Type for the function invocation.",
	},
	cli.BoolFlag{
		Name:  "display-call-id",
		Usage: "whether display call ID or not",
	},
	cli.StringFlag{
		Name:  "output",
		Usage: "Output format (json)",
	},
}

// InvokeCommand returns call cli.command
func InvokeCommand() cli.Command {
	cl := invokeCmd{}
	return cli.Command{
		Name:    "invoke",
		Usage:   "\tInvoke a remote function",
		Aliases: []string{"iv"},
		Before: func(c *cli.Context) error {
			var err error
			cl.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			cl.client = cl.provider.APIClientv2()
			return nil
		},
		ArgsUsage:   "[app-name] [function-name]",
		Flags:       InvokeFnFlags,
		Category:    "DEVELOPMENT COMMANDS",
		Description: `This command invokes a function. Users may send input to their function by passing input to this command via STDIN.`,
		Action:      cl.Invoke,
		BashComplete: func(c *cli.Context) {
			switch len(c.Args()) {
			case 0:
				app.BashCompleteApps(c)
			case 1:
				fn.BashCompleteFns(c)
			}
		},
	}
}

func (cl *invokeCmd) Invoke(c *cli.Context) error {
	var contentType string

	invokeURL := c.String("endpoint")

	if invokeURL == "" {

		appName := c.Args().Get(0)
		fnName := c.Args().Get(1)

		if appName == "" || fnName == "" {
			return errors.New("missing app and function name")
		}

		app, err := app.GetAppByName(cl.client, appName)
		if err != nil {
			return err
		}
		fn, err := fn.GetFnByName(cl.client, app.ID, fnName)
		if err != nil {
			return err
		}
		var ok bool
		invokeURL, ok = fn.Annotations[FnInvokeEndpointAnnotation].(string)
		if !ok {
			return fmt.Errorf("Fn invoke url annotation not present, %s", FnInvokeEndpointAnnotation)
		}
	}
	content := stdin()
	wd := common.GetWd()

	if c.String("content-type") != "" {
		contentType = c.String("content-type")
	} else {
		_, ff, err := common.FindAndParseFuncFileV20180708(wd)
		if err == nil && ff.Content_type != "" {
			contentType = ff.Content_type
		}
	}

	resp, err := client.Invoke(cl.provider,
		client.InvokeRequest{
			URL:         invokeURL,
			Content:     content,
			Env:         c.StringSlice("e"),
			ContentType: contentType,
		},
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	outputFormat := strings.ToLower(c.String("output"))
	if outputFormat == "json" {
		outputJSON(os.Stdout, resp)
	} else {
		outputNormal(os.Stdout, resp, c.Bool("display-call-id"))
	}
	// TODO we should have a 'raw' option to output the raw http request, it may be useful, idk

	return nil
}

func outputJSON(output io.Writer, resp *http.Response) {
	var b bytes.Buffer
	// TODO this is lame
	io.Copy(&b, resp.Body)

	i := struct {
		Body       string      `json:"body"`
		Headers    http.Header `json:"headers"`
		StatusCode int         `json:"status_code"`
	}{
		Body:       b.String(),
		Headers:    resp.Header,
		StatusCode: resp.StatusCode,
	}

	enc := json.NewEncoder(output)
	enc.SetIndent("", "    ")
	enc.Encode(i)
}

func outputNormal(output io.Writer, resp *http.Response, includeCallID bool) {
	if cid, ok := resp.Header[CallIDHeader]; ok && includeCallID {
		fmt.Fprint(output, fmt.Sprintf("Call ID: %v\n", cid[0]))
	}

	var body io.Reader = resp.Body
	if resp.StatusCode >= 400 {
		// if we don't get json, we need to buffer the input so that we can
		// display the user's function output as it was...
		var b bytes.Buffer
		body = io.TeeReader(resp.Body, &b)

		var msg struct {
			Message string `json:"message"`
		}
		err := json.NewDecoder(body).Decode(&msg)
		if err == nil && msg.Message != "" {
			// this is likely from fn, so unravel this...
			// TODO this should be stderr maybe? meh...
			fmt.Fprintf(output, "Error invoking function. status: %v message: %v\n", resp.StatusCode, msg.Message)
			return
		}

		// read anything written to buffer first, then copy out rest of body
		body = io.MultiReader(&b, resp.Body)
	}

	// at this point, it's not an fn error, so output function output as is

	lcc := lastCharChecker{reader: body}
	body = &lcc
	io.Copy(output, body)

	// #1408 - flush stdout
	if lcc.last != '\n' {
		fmt.Fprintln(output)
	}
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
