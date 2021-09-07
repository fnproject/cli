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

package context

import (
	"github.com/fnproject/cli/config"
)

// Info holds the information found in the context YAML file
type Info struct {
	Current bool   `json:"current"`
	Name    string `json:"name"`
	*config.ContextFile
}

// NewInfo creates an instance of the contextInfo
// by parsing the provided context YAML file. This is used
// for outputting the context information
func NewInfo(name string, isCurrent bool, contextFile *config.ContextFile) *Info {
	return &Info{
		Name:        name,
		Current:     isCurrent,
		ContextFile: contextFile,
	}
}
