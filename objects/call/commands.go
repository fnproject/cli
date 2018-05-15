package call

import (
	"github.com/fnproject/cli/client"
	"github.com/urfave/cli"
)

func Get() cli.Command {
	c := callsCmd{}
	return cli.Command{
		Name:      "calls",
		ShortName: "call",
		Usage:     "get function call info per app",
		Aliases:   []string{"cl"},
		Before: func(cxt *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			c.client = provider.APIClient()
			return nil
		},
		ArgsUsage: "<app> <call-id>",
		Action:    c.get,
	}
}

func List() cli.Command {
	c := callsCmd{}
	return cli.Command{
		Name:      "calls",
		ShortName: "call",
		Usage:     "list all calls for the specific app. Route is optional",
		Aliases:   []string{"cl"},
		Before: func(cxt *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			c.client = provider.APIClient()
			return nil
		},
		ArgsUsage: "<app>",
		Action:    c.list,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "path",
				Usage: "function's path",
			},
			cli.StringFlag{
				Name:  "cursor",
				Usage: "pagination cursor",
			},
			cli.StringFlag{
				Name:  "from-time",
				Usage: "'start' timestamp",
			},
			cli.StringFlag{
				Name:  "to-time",
				Usage: "'stop' timestamp",
			},
			cli.Int64Flag{
				Name:  "n",
				Usage: "number of calls to return",
				Value: int64(100),
			},
		},
	}
}
