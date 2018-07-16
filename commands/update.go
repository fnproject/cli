package commands

import (
	"github.com/urfave/cli"
)

// UpdateCommand returns update cli.command
func UpdateCommand() cli.Command {
	return cli.Command{
		Name:        "update",
		Aliases:     []string{"up"},
		Usage:       "\tUpdate elements of created object",
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<subcommand>",
		Description: "This command updates created objects ('app', 'context', 'function', 'route', 'server', 'trigger')",
		Subcommands: GetCommands(UpdateCmds),
	}
}
