package commands

import (
	"fmt"

	"github.com/fnproject/cli/common"
	"github.com/urfave/cli"
)

// BuildCommand returns build cli.command
func BuildCommand() cli.Command {
	cmd := buildcmd{}
	flags := append([]cli.Flag{}, cmd.flags()...)
	return cli.Command{
		Name:        "build",
		Usage:       "\tBuild function version",
		Category:    "DEVELOPMENT COMMANDS",
		Description: "This is the description",
		Aliases:     []string{"bu"},
		Flags:       flags,
		Action:      cmd.build,
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
			Usage:       "Verbose mode",
			Destination: &b.verbose,
		},
		cli.BoolFlag{
			Name:        "no-cache",
			Usage:       "Don't use docker cache",
			Destination: &b.noCache,
		},
		cli.StringSliceFlag{
			Name:  "build-arg",
			Usage: "Set build-time variables",
		},
		cli.StringFlag{
			Name:  "working-dir, w",
			Usage: "Specify the working directory to build a function, must be the full path.",
		},
	}
}

// build will take the found valid function and build it
func (b *buildcmd) build(c *cli.Context) error {
	var err error

	dir := common.GetDir(c)

	ffV, err := common.ReadInFuncFile()
	version := common.GetFuncYamlVersion(ffV)
	if version == common.LatestYamlVersion {
		fpath, ff, err := common.FindAndParseFuncFileV20180707(dir)
		if err != nil {
			return err
		}

		buildArgs := c.StringSlice("build-arg")
		ff, err = common.BuildFuncV20180707(c, fpath, ff, buildArgs, b.noCache)
		if err != nil {
			return err
		}

		fmt.Printf("Function %v built successfully.\n", ff.ImageNameV20180707())
		return nil

	}

	fpath, ff, err := common.FindAndParseFuncfile(dir)
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
