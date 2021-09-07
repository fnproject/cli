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
	"os/exec"

	"github.com/urfave/cli"
)

// StopCommand returns stop server cli.command
func StopCommand() cli.Command {
	return cli.Command{
		Name:        "stop",
		Usage:       "Stop a function server",
		Category:    "SERVER COMMANDS",
		Description: "This command stops a Fn server.",
		Action:      stop,
	}
}
func stop(c *cli.Context) error {
	cmd := exec.Command("docker", "stop", "fnserver")
	err := cmd.Run()
	if err != nil {
		return errors.New("Failed to stop 'fnserver'")
	}

	fmt.Println("Successfully stopped 'fnserver'")

	return err
}
