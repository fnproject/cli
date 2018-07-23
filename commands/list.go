package commands

import (
	"github.com/urfave/cli"
)

// ListCommand returns list cli.command
func ListCommand() cli.Command {
	return cli.Command{
		Name:        "list",
		Aliases:     []string{"ls"},
		Usage:       "\tReturn a list of created objects",
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<subcommand>",
		Description: "This command returns a list of created objects ('app', 'call', 'context', 'function', 'route' or 'trigger') or configurations.",
		Subcommands: GetCommands(ListCmds),
	}
}
