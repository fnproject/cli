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
		Description: "This command gets a list of all created objects of the same type ('apps', 'calls', 'configurations', 'contexts', 'functions', 'routes', 'triggers').",
		Subcommands: GetCommands(ListCmds),
	}
}
