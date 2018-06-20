package call

import (
	"github.com/fnproject/cli/client"
	"github.com/urfave/cli"
)

func Get() cli.Command {
	c := callsCmd{}
	return cli.Command{
		Name:        "calls",
		ShortName:   "call",
		Usage:       "Get function call info per app",
		Aliases:     []string{"cl"},
		Category:    "MANAGEMENT COMMAND",
		Description: "This is the description",
		Before: func(cxt *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			c.client = provider.APIClient()
			return nil
		},
		ArgsUsage: "<app_name> <call-id>",
		Action:    c.get,
	}
}

func List() cli.Command {
	c := callsCmd{}
	return cli.Command{
		Name:        "calls",
		ShortName:   "call",
		Usage:       "List all calls for the specific app. Route is optional",
		Aliases:     []string{"cl"},
		Category:    "MANAGEMENT COMMAND",
		Description: "This is the description",
		Before: func(cxt *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			c.client = provider.APIClient()
			return nil
		},
		ArgsUsage: "<app_name>",
		Action:    c.list,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "path",
				Usage: "Function's path",
			},
			cli.StringFlag{
				Name:  "cursor",
				Usage: "Pagination cursor",
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
				Usage: "Number of calls to return",
				Value: int64(100),
			},
		},
	}
}
