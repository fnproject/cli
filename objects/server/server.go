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

package server

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/fnproject/cli/common"
	"github.com/urfave/cli"
)

func update(c *cli.Context) error {
	args := []string{"pull",
		common.FunctionsDockerImage,
	}
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
	select {
	case <-sigC:
		log.Println("Interrupt caught, exiting")
		err = cmd.Process.Kill()
		if err != nil {
			log.Println("Error: could not kill process")
		}
	case err := <-done:
		if err != nil {
			log.Println("Processed finished with error:", err)
		} else {
			log.Println("Process finished gracefully")
		}
	}
	return nil
}
