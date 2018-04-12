package main

import (
	"github.com/urfave/cli"
)

func images() cli.Command {
	return cli.Command{
		Name:  "images",
		Usage: "manage function images",
		Subcommands: []cli.Command{
			build(),
			deploy(),
			bump(),
			call(),
			push(),
			run(),
			testfn(),
		},
	}
}
