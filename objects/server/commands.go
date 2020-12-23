package server

import (
	"github.com/urfave/cli/v2"
)

func Update() *cli.Command {
	return &cli.Command{
		Name:        "server",
		Usage:       "Pulls latest functions server",
		Category:    "MANAGEMENT COMMAND",
		Description: "This command updates the latest Fn server",
		Aliases:     []string{"sv"},
		Action:      update,
	}
}
