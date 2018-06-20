package commands

import (
	"github.com/urfave/cli"
)

// UpdateCommand returns update cli.command
func UpdateCommand() cli.Command {
	return cli.Command{
		Name:        "update",
		Aliases:     []string{"up"},
		Usage:       "Update elements of created object",
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<subcommand>",
		Description: "This is the description",
		Subcommands: GetCommands(UpdateCmds),
	}
}
