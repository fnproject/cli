package commands

import (
	"os"

	"github.com/fnproject/cli/client"
	"github.com/fnproject/cli/common"
	"github.com/fnproject/cli/objects/route"
	"github.com/fnproject/cli/run"
	fnclient "github.com/fnproject/fn_go/client"
	"github.com/fnproject/fn_go/provider"
	"github.com/urfave/cli"
)

type callCmd struct {
	provider provider.Provider
	client   *fnclient.Fn
}

// CallCommand returns call cli.command
func CallCommand() cli.Command {
	cl := callCmd{}
	return cli.Command{
		Name:    "call",
		Usage:   "call a remote function",
		Aliases: []string{"cl"},
		Before: func(c *cli.Context) error {
			var err error
			cl.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			cl.client = cl.provider.APIClient()
			return nil
		},
		ArgsUsage: "<app> </path>",
		Flags:     route.CallFnFlags,
		Category:  "DEVELOPMENT COMMANDS",
		Action:    cl.Call,
	}
}

func (cl *callCmd) Call(c *cli.Context) error {
	var contentType string

	appName := c.Args().Get(0)
	route := route.WithoutSlash(c.Args().Get(1))
	content := run.Stdin()
	wd := common.GetWd()

	if c.String("content-type") != "" {
		contentType = c.String("content-type")
	} else {
		_, ff, err := common.FindAndParseFuncfile(wd)
		if err == nil && ff.ContentType != "" {
			contentType = ff.ContentType
		}
	}

	return client.CallFN(cl.provider, appName, route, content, os.Stdout, c.String("method"), c.StringSlice("e"), contentType, c.Bool("display-call-id"))
}
