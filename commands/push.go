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
	"errors"
	"fmt"

	"github.com/fnproject/cli/common"
	"github.com/urfave/cli"
)

// PushCommand returns push cli.command
func PushCommand() cli.Command {
	cmd := pushcmd{}
	var flags []cli.Flag
	flags = append(flags, cmd.flags()...)
	return cli.Command{
		Name:        "push",
		Usage:       "\tPush function to docker registry",
		Aliases:     []string{"p"},
		Category:    "DEVELOPMENT COMMANDS",
		Description: "This command pushes the created image to the Docker registry.",
		Flags:       flags,
		Action:      cmd.push,
	}
}

type pushcmd struct {
	registry string
}

func (p *pushcmd) flags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:        "verbose, v",
			Usage:       "Verbose mode",
			Destination: &common.CommandVerbose,
		},
		cli.StringFlag{
			Name:        "registry",
			Usage:       "Set the Docker owner for images and optionally the registry. This will be prefixed to your function name for pushing to Docker registries.\n eg: `--registry username` will set only the owner prefix. `--registry registry.hub.docker.com/username` will set the registry and owner.",
			Destination: &p.registry,
		},
	}
}

// push will take the found function and check for the presence of a
// Dockerfile, and run a three step process: parse functions file,
// push the container, and finally it will update the function. Optionally,
// the function can be overriden inside the functions file.
func (p *pushcmd) push(c *cli.Context) error {
	ffV, err := common.ReadInFuncFile()
	version := common.GetFuncYamlVersion(ffV)
	if version == common.LatestYamlVersion {
		_, ff, err := common.LoadFuncFileV20180708(".")
		if err != nil {
			if _, ok := err.(*common.NotFoundError); ok {
				return errors.New("Image name is missing or no function file found")
			}
			return err
		}

		fmt.Println("pushing", ff.ImageNameV20180708())

		if err := common.PushV20180708(ff); err != nil {
			return err
		}

		fmt.Printf("Function %v pushed successfully to the registry.\n", ff.ImageNameV20180708())
		return nil
	}

	_, ff, err := common.LoadFuncfile(".")

	if err != nil {
		if _, ok := err.(*common.NotFoundError); ok {
			return errors.New("Image name is missing or no function file found")
		}
		return err
	}

	fmt.Println("pushing", ff.ImageName())

	if err := common.Push(ff); err != nil {
		return err
	}

	fmt.Printf("Function %v pushed successfully to the registry.\n", ff.ImageName())
	return nil
}
