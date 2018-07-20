package log

import (
	"github.com/fnproject/cli/client"
	"github.com/urfave/cli"
)

func Get() cli.Command {
	l := logsCmd{}
	return cli.Command{
		Name:        "logs",
		Usage:       "Get logs for a call, providing call_id or last (l) to retrieve the most recent calls logs",
		Aliases:     []string{"log", "lg"},
		Category:    "MANAGEMENT COMMAND",
		Description: "This is the description",
		Before: func(cxt *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			l.client = provider.APIClient()
			return nil
		},
		ArgsUsage: "<app-name> <call-id>",
		Action:    l.get,
	}
}
