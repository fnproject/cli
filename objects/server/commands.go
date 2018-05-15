package server

import (
	"github.com/urfave/cli"
)

func Update() cli.Command {
	return cli.Command{
		Name:    "server",
		Usage:   "pulls latest functions server",
		Aliases: []string{"sv"},
		Action:  update,
	}
}
