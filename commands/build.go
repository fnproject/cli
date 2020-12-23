package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fnxproject/cli/common"
	"github.com/urfave/cli/v2"
)

// BuildCommand returns build cli.command
func BuildCommand() *cli.Command {
	cmd := buildcmd{}
	flags := append([]cli.Flag{}, cmd.flags()...)
	return &cli.Command{
		Name:        "build",
		Usage:       "\tBuild function version",
		Category:    "DEVELOPMENT COMMANDS",
		Description: "This command builds a new function.",
		ArgsUsage:   "[function-subdirectory]",
		Aliases:     []string{"bu"},
		Flags:       flags,
		Action:      cmd.build,
	}
}

type buildcmd struct {
	noCache bool
}

func (b *buildcmd) flags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:        "verbose, v",
			Usage:       "Verbose mode",
			Destination: &common.CommandVerbose,
		},
		&cli.BoolFlag{
			Name:        "no-cache",
			Usage:       "Don't use docker cache",
			Destination: &b.noCache,
		},
		&cli.StringSliceFlag{
			Name:  "build-arg",
			Usage: "Set build-time variables",
		},
		&cli.StringFlag{
			Name:  "working-dir, w",
			Usage: "Specify the working directory to build a function, must be the full path.",
		},
	}
}

// build will take the found valid function and build it
func (b *buildcmd) build(c *cli.Context) error {
	dir := common.GetDir(c)

	path := c.Args().First()
	if path != "" {
		fmt.Printf("Building function at: ./%s\n", path)
		dir = filepath.Join(dir, path)
	}

	err := os.Chdir(dir)
	if err != nil {
		return err
	}
	defer os.Chdir(dir)

	ffV, err := common.ReadInFuncFile()
	if err != nil {
		return err
	}

	switch common.GetFuncYamlVersion(ffV) {
	case common.LatestYamlVersion:
		fpath, ff, err := common.FindAndParseFuncFileV20180708(dir)
		if err != nil {
			return err
		}

		buildArgs := c.StringSlice("build-arg")
		ff, err = common.BuildFuncV20180708(common.IsVerbose(), fpath, ff, buildArgs, b.noCache)
		if err != nil {
			return err
		}

		fmt.Printf("Function %v built successfully.\n", ff.ImageNameV20180708())
		return nil

	default:
		fpath, ff, err := common.FindAndParseFuncfile(dir)
		if err != nil {
			return err
		}

		buildArgs := c.StringSlice("build-arg")
		ff, err = common.BuildFunc(common.IsVerbose(), fpath, ff, buildArgs, b.noCache)
		if err != nil {
			return err
		}

		fmt.Printf("Function %v built successfully.\n", ff.ImageName())
		return nil
	}
}
