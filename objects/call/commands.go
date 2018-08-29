package call

import (
	"github.com/fnproject/cli/client"
	"github.com/urfave/cli"
)

func Get() cli.Command {
	c := callsCmd{}
	return cli.Command{
		Name:        "call",
		Usage:       "Get function call info per app",
		Aliases:     []string{"calls", "cl"},
		Category:    "MANAGEMENT COMMAND",
		Description: "This command gets information call of a route.",
		Before: func(cxt *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			c.client = provider.APIClient()
			return nil
		},
		ArgsUsage: "<app-name> <call-id>",
		Action:    c.get,
	}
}

func List() cli.Command {
	c := callsCmd{}
	return cli.Command{
		Name:        "calls",
		Usage:       "List all calls for the specific app (route is optional)",
		Description: "This command provides a list of defined calls for a specific application.",
		Aliases:     []string{"call", "cl"},
		Category:    "MANAGEMENT COMMAND",
		Before: func(cxt *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			c.client = provider.APIClient()
			return nil
		},
		ArgsUsage: "<app-name>",
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
			cli.StringFlag{
				Name:  "output",
				Usage: "Output format (json)",
				Value: "",
			},
		},
	}
}
