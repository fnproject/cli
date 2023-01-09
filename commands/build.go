/*
 * Copyright (c) 2019, 2020 Oracle and/or its affiliates. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package commands

import (
	"fmt"
	"github.com/fnproject/cli/common"
	"github.com/urfave/cli"
	"os"
	"path/filepath"
)

// BuildCommand returns build cli.command
func BuildCommand() cli.Command {
	cmd := buildcmd{}
	flags := append([]cli.Flag{}, cmd.flags()...)
	return cli.Command{
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
		cli.BoolFlag{
			Name:        "verbose, v",
			Usage:       "Verbose mode",
			Destination: &common.CommandVerbose,
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
		ff, err = common.BuildFuncV20180708(common.IsVerbose(), fpath, ff, buildArgs, b.noCache, nil)
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

		//to remove
		fmt.Println(ff.Platforms)
		if len(ff.Platforms) > 1 {
			fmt.Printf("fn build is not supported for multi-arch images")
			return nil
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
