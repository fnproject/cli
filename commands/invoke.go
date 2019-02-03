package commands

import (
	"fmt"
	"os"

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
const FnInvokeEndpointAnnotation = "fnproject.io/fn/invokeEndpoint"

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
		Name:  "method",
		Usage: "Http method for function",
	},
	cli.StringFlag{
		Name:  "content-type",
		Usage: "The payload Content-Type for the function invocation.",
	},
	cli.BoolFlag{
		Name:  "display-call-id",
		Usage: "whether display call ID or not",
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

	err := client.Invoke(cl.provider, invokeURL, content, os.Stdout, c.String("method"), c.StringSlice("e"), contentType, c.Bool("display-call-id"))
	// we don't want to show the help message if invoke fails, just copy error to
	// stderr but don't return the error from Invoke. also exit early here
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	return nil
}
