package commands

import (
	"fmt"
	"os"

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
		ArgsUsage:   "<app-name> <function-name>",
		Flags:       InvokeFnFlags,
		Category:    "DEVELOPMENT COMMANDS",
		Description: "This command explicitly invokes a function.",
		Action:      cl.Invoke,
	}
}

func (cl *invokeCmd) Invoke(c *cli.Context) error {
	var contentType string

	appName := c.Args().Get(0)
	fnName := c.Args().Get(1)

	app, err := app.GetAppByName(cl.client, appName)
	if err != nil {
		return err
	}
	fn, err := fn.GetFnByName(cl.client, app.ID, fnName)
	if err != nil {
		return err
	}

	content := run.Stdin()
	wd := common.GetWd()

	if c.String("content-type") != "" {
		contentType = c.String("content-type")
	} else {
		_, ff, err := common.FindAndParseFuncFileV20180708(wd)
		if err == nil && ff.Content_type != "" {
			contentType = ff.Content_type
		}
	}

	invokeURL := fn.Annotations[FnInvokeEndpointAnnotation]
	if invokeURL == nil {
		return fmt.Errorf("Fn invoke url annotation not present, %s", FnInvokeEndpointAnnotation)
	}

	return client.Invoke(cl.provider, invokeURL.(string), content, os.Stdout, c.String("method"), c.StringSlice("e"), contentType, c.Bool("display-call-id"))
}
