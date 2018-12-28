package log

import (
	"github.com/fnproject/cli/client"
	"github.com/fnproject/cli/objects/app"
	"github.com/fnproject/cli/objects/fn"
	"github.com/urfave/cli"
)

// Get logs command
func Get() cli.Command {
	l := logsCmd{}
	return cli.Command{
		Name:        "logs",
		Usage:       "Get logs for a call, providing call_id or last (l)",
		Aliases:     []string{"log", "lg"},
		Category:    "MANAGEMENT COMMAND",
		Description: "This command gets logs for a call to retrieve the most recent calls logs.",
		Before: func(cxt *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			l.client = provider.APIClientv2()
			return nil
		},
		ArgsUsage: "<app-name> <function-name> <call-id>",
		Action:    l.get,
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
