package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli"
)

func build() cli.Command {
	cmd := buildcmd{}
	flags := append([]cli.Flag{}, cmd.flags()...)
	return cli.Command{
		Name:   "build",
		Usage:  "build function version",
		Flags:  flags,
		Action: cmd.build,
	}
}

type buildcmd struct {
	verbose bool
	noCache bool
}

func (b *buildcmd) flags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:        "verbose",
			Usage:       "verbose mode",
			Destination: &b.verbose,
		},
		cli.BoolFlag{
			Name:        "no-cache",
			Usage:       "Don't use docker cache",
			Destination: &b.noCache,
		},
	}
}

// build will take the found valid function and build it
func (b *buildcmd) build(c *cli.Context) error {
	path, err := os.Getwd()
	if err != nil {
		return err
	}
	fpath, ff, err := findAndParseFuncfile(path)
	if err != nil {
		return err
	}
	// get name from directory if it's not defined
	if ff.Name == "" {
		ff.Name = filepath.Base(filepath.Dir(fpath)) // todo: should probably make a copy of ff before changing it
	}

	ff, err = buildfunc(c, fpath, ff, b.noCache)
	if err != nil {
		return err
	}

	fmt.Printf("Function %v built successfully.\n", ff.ImageName())
	return nil
}
