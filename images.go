package main

import (
	"github.com/fnproject/cli/common"
	"github.com/fnproject/cli/objects/route"
	"github.com/fnproject/cli/run"
	"github.com/urfave/cli"
)

func images() cli.Command {
	return cli.Command{
		Name:  "images",
		Usage: "manage function images",
		Subcommands: []cli.Command{
			build(),
			deploy(),
			common.Bump(),
			route.Call(),
			push(),
			run.Run(),
			testfn(),
		},
	}
}
