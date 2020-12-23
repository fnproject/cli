package commands

import (
	"strings"

	"github.com/fnxproject/cli/common"
	"github.com/urfave/cli/v2"
)

// ConfigCommand returns config cli.command dependant on command parameter
func ConfigCommand(command string) *cli.Command {
	var cmds []*cli.Command
	switch command {
	case "list":
		cmds = GetCommands(ConfigListCmds)
	case "get":
		cmds = GetCommands(ConfigGetCmds)
	case "configure":
		cmds = GetCommands(ConfigCmds)
	case "unset", "delete":
		cmds = GetCommands(ConfigUnsetCmds)
	}

	usage := strings.Title(command) + " configurations for apps and functions"

	return &cli.Command{
		Name:         "config",
		Usage:        usage,
		Aliases:      []string{"cf"},
		ArgsUsage:    "<subcommand>",
		Description:  "This command unsets the configuration of created objects ('app' or 'function').",
		Subcommands:  cmds,
		BashComplete: common.DefaultBashComplete,
	}
}
