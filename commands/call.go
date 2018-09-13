package commands

import (
	"fmt"

	"github.com/fnproject/fn_go/provider"
	"github.com/urfave/cli"
)

type callCmd struct {
	provider provider.Provider
}

// CallCommand returns call cli.command
func CallCommand() cli.Command {
	cl := callCmd{}
	return cli.Command{
		Name:        "call",
		Usage:       "\tPrompts for migration.",
		Aliases:     []string{"cl"},
		ArgsUsage:   "",
		Category:    "DEVELOPMENT COMMANDS",
		Description: "This command no longer executes a function, instead it prompts users to migrate their routes to fn/triggers.",
		Action:      cl.Call,
	}
}

func (cl *callCmd) Call(c *cli.Context) {
	fmt.Printf("Routes are no longer supported, please use the migrate command to update your metadata.\n")
}
