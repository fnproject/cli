package commands

import (
	"github.com/fnxproject/cli/common"
	"github.com/urfave/cli/v2"
)

// ConfigureCommand returns configure cli.command
func ConfigureCommand() *cli.Command {
	return &cli.Command{
		Name:         "config",
		Aliases:      []string{"cf"},
		Usage:        "\tSet configuration for an object",
		Category:     "MANAGEMENT COMMANDS",
		ArgsUsage:    "<subcommand>",
		Description:  "This command sets a configuration key for an 'app' or 'function'.",
		Subcommands:  GetCommands(ConfigCmds),
		BashComplete: common.DefaultBashComplete,
	}
}
