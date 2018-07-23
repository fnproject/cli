package commands

import (
	"github.com/urfave/cli"
)

// GetCommand returns get cli.command
func GetCommand() cli.Command {
	return cli.Command{
		Name:        "get",
		Aliases:     []string{"g"},
		Usage:       "\tGet an object to retrieve its information",
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<subcommand>",
		Description: "This commands gets a 'call', 'configuaration' or 'log' to retrieve information for an object ('app', 'route' or 'function').",
		Subcommands: GetCommands(GetCmds),
	}
}
