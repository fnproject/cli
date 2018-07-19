package commands

import (
	"github.com/urfave/cli"
)

// ConfigureCommand returns configure cli.command
func ConfigureCommand() cli.Command {
	return cli.Command{
		Name:        "config",
		Aliases:     []string{"cf"},
		Usage:       "\tSet configuration for an object",
		Category:    "MANAGEMENT COMMANDS",
		ArgsUsage:   "<subcommand>",
		Description: "This command sets a configuaration key for an application, function or route.",
		Subcommands: GetCommands(ConfigCmds),
	}
}
