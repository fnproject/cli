package main

import (
	"github.com/fnproject/fn_go/client"
	"github.com/urfave/cli"
)

type imagesCmd struct {
	*client.Fn
}

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
