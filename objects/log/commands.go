package log

import (
	"github.com/fnproject/cli/client"
	"github.com/urfave/cli"
)

func Get() cli.Command {
	l := logsCmd{}
	return cli.Command{
		Name:      "logs",
		ShortName: "log",
		Usage:     "get logs for a call. Must provide call_id or last (l) to retrieve the most recent calls logs.",
		Aliases:   []string{"lg"},
		Before: func(cxt *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			l.client = provider.APIClient()
			return nil
		},
		ArgsUsage: "<app> <call-id>",
		Action:    l.get,
	}
}
