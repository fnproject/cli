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

	"github.com/fnproject/cli/client"
	"github.com/fnproject/cli/config"
	"github.com/urfave/cli"
)

// VersionCommand
func VersionCommand() cli.Command {
	return cli.Command{
		Name:        "version",
		Usage:       "Display Fn CLI and Fn Server versions",
		Description: "This command shows the version of the Fn CLI being used and the version of the Fn Server referenced by the current context, if available.",
		Action:      versionCMD,
	}
}

func versionCMD(c *cli.Context) error {
	provider, err := client.CurrentProvider()
	if err != nil {
		return err
	}

	ver := config.GetVersion("latest")
	if ver == "" {
		ver = "Client version: " + config.Version
	}
	fmt.Println(ver)

	versionClient := provider.VersionClient()
	v, err := versionClient.GetVersion(nil)
	if err != nil {
		fmt.Println("Server version: ", "?")
		return nil
	}
	fmt.Println("Server version: ", v.Payload.Version)
	return nil
}
