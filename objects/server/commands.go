package server

import (
	"github.com/urfave/cli"
)

func Update() cli.Command {
	return cli.Command{
		Name:        "server",
		Usage:       "Pulls latest functions server",
		Category:    "MANAGEMENT COMMAND",
		Description: "This is the description",
		Aliases:     []string{"sv"},
		Action:      update,
	}
}
