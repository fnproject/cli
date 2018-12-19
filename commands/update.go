package commands

import (
	"fmt"

	"github.com/urfave/cli"
)

// UpdateCommand returns update cli.command
func UpdateCommand() cli.Command {
	return cli.Command{
		Name:        "update",
		Aliases:     []string{"up"},
		Usage:       "\tUpdate a created object",
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<subcommand>",
		Description: "This command updates an object ('app', 'context', 'function', 'server' or 'trigger').",
		Subcommands: GetCommands(UpdateCmds),
		BashComplete: func(ctx *cli.Context) {
			for _, c := range GetCommands(UpdateCmds) {
				fmt.Println(c.Name)
			}
		},
	}
}
