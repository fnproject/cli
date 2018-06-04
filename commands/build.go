package commands

import (
	"fmt"
	"os"

	"github.com/fnproject/cli/common"
	"github.com/urfave/cli"
)

// BuildCommand returns build cli.command
func BuildCommand() cli.Command {
	cmd := buildcmd{}
	flags := append([]cli.Flag{}, cmd.flags()...)
	return cli.Command{
		Name:     "build",
		Usage:    "build function version",
		Category: "DEVELOPMENT COMMANDS",
		Aliases:  []string{"bu"},
		Flags:    flags,
		Action:   cmd.build,
	}
}

type buildcmd struct {
	verbose bool
	noCache bool
}

func (b *buildcmd) flags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:        "v",
			Usage:       "verbose mode",
			Destination: &b.verbose,
		},
		cli.BoolFlag{
			Name:        "no-cache",
			Usage:       "Don't use docker cache",
			Destination: &b.noCache,
		},
		cli.StringSliceFlag{
			Name:  "build-arg",
			Usage: "set build-time variables",
		},
	}
}

// build will take the found valid function and build it
func (b *buildcmd) build(c *cli.Context) error {
	path, err := os.Getwd()
	if err != nil {
		return err
	}
	fpath, ff, err := common.FindAndParseFuncfile(path)
	if err != nil {
		return err
	}

	buildArgs := c.StringSlice("build-arg")
	ff, err = common.BuildFunc(c, fpath, ff, buildArgs, b.noCache)
	if err != nil {
		return err
	}

	fmt.Printf("Function %v built successfully.\n", ff.ImageName())
	return nil
}
