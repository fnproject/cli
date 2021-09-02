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
	"strings"

	"github.com/fnproject/cli/common"
	"github.com/urfave/cli"
)

// ConfigCommand returns config cli.command dependant on command parameter
func ConfigCommand(command string) cli.Command {
	var cmds []cli.Command
	switch command {
	case "list":
		cmds = GetCommands(ConfigListCmds)
	case "get":
		cmds = GetCommands(ConfigGetCmds)
	case "configure":
		cmds = GetCommands(ConfigCmds)
	case "unset", "delete":
		cmds = GetCommands(ConfigUnsetCmds)
	}

	usage := strings.Title(command) + " configurations for apps and functions"

	return cli.Command{
		Name:         "config",
		ShortName:    "config",
		Usage:        usage,
		Aliases:      []string{"cf"},
		ArgsUsage:    "<subcommand>",
		Description:  "This command unsets the configuration of created objects ('app' or 'function').",
		Subcommands:  cmds,
		BashComplete: common.DefaultBashComplete,
	}
}
