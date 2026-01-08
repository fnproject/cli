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
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/fnproject/cli/common"
	"github.com/fnproject/cli/config"
	"github.com/urfave/cli"
)

// StartCommand returns start server cli.command
func StartCommand() cli.Command {
	return cli.Command{
		Name:        "start",
		Usage:       "Start a local Fn server",
		Category:    "SERVER COMMANDS",
		Description: "This command starts a local Fn server by downloading its docker image.",
		Action:      start,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "log-level",
				Usage: "--log-level debug to enable debugging",
			},
			cli.BoolFlag{
				Name:  "detach, d",
				Usage: "Run container in background.",
			},
			cli.StringFlag{
				Name:  "env-file",
				Usage: "Path to Fn server configuration file.",
			},
			cli.StringFlag{
				Name:  "version",
				Usage: "Specify a specific fnproject/fnserver version to run, ex: '1.2.3'.",
				Value: "latest",
			},
			cli.IntFlag{
				Name:  "port, p",
				Value: 8080,
				Usage: "Specify port number to bind to on the host.",
			},
			cli.StringFlag{
				Name: "iofs-dir",
				Usage: "Directory or container-engine volume where Fn Server creates the UNIX socket for " +
					"cross-container communication. Defaults to named volume \"fnserversocket\"",
			},
		},
	}
}

func start(c *cli.Context) error {
	var fnDir string
	home := config.GetHomeDir()

	if c.String("data-dir") != "" {
		fnDir = c.String("data-dir")
	} else {
		fnDir = filepath.Join(home, ".fn")
	}

	/*
			User could override "iofs" directory. The iofs path is used by Fn Server to create unix socket file
			for the communication between Fn Server and Fn Containers.

			By default, the iofs directory is a container engine volume that will be created by container engine
		    automatically. We switched from plain directory on host machine to container engine volume because:

			1. For user who is using Podman or Rancher desktop, there is an issue that the socket file could not be created
		    under HOME directory (see: https://github.com/fnproject/fn/issues/1577#issuecomment-1297736260).

			2. On Windows docker desktop, if we use OS directory as iofs path, we also observed that the fsnotify did not
			notify fnserver container when the socket file was created. However, named volume worked well.

			As a result, we decided to use named volume as default as it works in all cases.
	*/
	iofsDir := "fnserversocket"
	if c.String("iofs-dir") != "" {
		iofsDir = c.String("iofs-dir")
	}

	args := []string{"run", "--rm", "-i",
		"--name", "fnserver",
		"-v", fmt.Sprintf("%s:/iofs:z", iofsDir),
		"-e", fmt.Sprintf("FN_IOFS_DOCKER_PATH=%s", iofsDir),
		"-e", "FN_IOFS_PATH=/iofs",
		"-v", fmt.Sprintf("%s/data:/app/data", fnDir),
		"-v", "/var/run/docker.sock:/var/run/docker.sock",
		"--privileged",
		"-p", fmt.Sprintf("%d:8080", c.Int("port")),
		"--entrypoint", "./fnserver",
	}
	if c.String("log-level") != "" {
		args = append(args, "-e", fmt.Sprintf("FN_LOG_LEVEL=%v", c.String("log-level")))
	}
	if c.String("env-file") != "" {
		args = append(args, "--env-file", c.String("env-file"))
	}
	if c.Bool("detach") {
		args = append(args, "-d")
	}

	image := fmt.Sprintf("%s:%s", common.FunctionsDockerImage, c.String("version"))
	args = append(args, image)
	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		log.Fatalln("Starting command failed:", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	// catch ctrl-c and kill
	sigC := make(chan os.Signal, 2)
	signal.Notify(sigC, os.Interrupt, syscall.SIGTERM)

	log.Println("¡¡¡ 'fn start' should NOT be used for PRODUCTION !!! see https://github.com/fnproject/fn-helm/")

	for {
		select {
		case <-sigC:
			log.Println("Interrupt caught, exiting")
			err = cmd.Process.Signal(syscall.SIGTERM)
			if err != nil {
				log.Println("Error: could not kill process:", err)
				return err
			}
		case err := <-done:
			if err != nil {
				log.Println("Error: processed finished with error", err)
			}
		}
		return err
	}
}
