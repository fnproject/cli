package commands

import (
	"github.com/fnproject/cli/common"
	"github.com/fnproject/cli/objects/app"
	"github.com/urfave/cli"
)

func ChangeCompartmentCommand() cli.Command {
	cmds := Cmd{
		"apps": app.ChangeCompartment(),
	}
	return cli.Command{
		Name:         "change-compartment",
		Usage:        "\tMove a supported resource to another compartment",
		Category:     "MANAGEMENT COMMANDS",
		ArgsUsage:    "<subcommand>",
		Description:  "This command changes the compartment for supported OCI-backed resources.",
		Subcommands:  GetCommands(cmds),
		BashComplete: common.DefaultBashComplete,
	}
}