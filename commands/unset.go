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
	"github.com/fnproject/cli/common"
	"github.com/urfave/cli"
)

// UnsetCommand returns unset cli.command
func UnsetCommand() cli.Command {
	return cli.Command{
		Name:         "unset",
		Aliases:      []string{"un"},
		Usage:        "\tUnset elements of a created object",
		Category:     "MANAGEMENT COMMANDS",
		Hidden:       false,
		ArgsUsage:    "<subcommand>",
		Description:  "This command unsets elements ('configurations') for a created object ('app', 'function' or 'context').",
		Subcommands:  GetCommands(UnsetCmds),
		BashComplete: common.DefaultBashComplete,
	}
}
