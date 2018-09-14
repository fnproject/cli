package commands

import (
	"github.com/urfave/cli"
)

// CreateCommand returns create cli.command
func CreateCommand() cli.Command {
	return cli.Command{
		Name:        "create",
		Aliases:     []string{"c"},
		Usage:       "\tCreate a new object",
		Description: "This command creates a new object ('app', 'context', 'function', or 'trigger').",
		Hidden:      false,
		ArgsUsage:   "<object-type>",
		Category:    "MANAGEMENT COMMANDS",
		Subcommands: GetCommands(CreateCmds),
	}
}
