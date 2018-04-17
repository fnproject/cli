package main

import (
	//"errors"

	"github.com/urfave/cli"
)

var contextsPath string
var defaultContextPath string

func contextCmd() cli.Command {
	return cli.Command{
		Name:  "context",
		Usage: "manage context",
		Subcommands: []cli.Command{
			{
				Name:      "create",
				Aliases:   []string{"c"},
				Usage:     "create a new context",
				ArgsUsage: "<context> <provider>",
				Action:    create,
				Flags: []cli.Flag{
					cli.StringSliceFlag{
						Name:  "config",
						Usage: "context configuration",
					},
				},
			},
			{
				Name:    "list",
				Aliases: []string{"l"},
				Usage:   "list contexts",
				Action:  list,
			},
			{
				Name:      "set",
				Aliases:   []string{"s"},
				Usage:     "set context for future invocations",
				ArgsUsage: "<context>",
				Action:    set,
			},
		},
	}
}

func create(c *cli.Context) {

}

func set(c *cli.Context) {

}

func list(c *cli.Context) {

}
