package commands

import (
	"github.com/urfave/cli"
)

// CreateCommand returns create cli.command
func CreateCommand() cli.Command {
	return cli.Command{
		Name:        "create",
		Aliases:     []string{"c"},
		Usage:       "create",
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<object>",
		Subcommands: GetCommands(CreateCmds),
	}
}
