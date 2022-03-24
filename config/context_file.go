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

package config

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

// ContextFile defines the internal structure of a default context
type ContextFile struct {
	ContextProvider string `yaml:"provider" json:"provider"`
	EnvFnAPIURL     string `yaml:"api-url" json:"apiUrl"`
	EnvFnRegistry   string `yaml:"registry" json:"registry"`
}

// NewContextFile creates a new instance of the context YAML file
func NewContextFile(filePath string) (*ContextFile, error) {
	c := &ContextFile{}
	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		return c, err
	}
	if err = yaml.Unmarshal(contents, c); err != nil {
		return c, err
	}
	return c, nil
}
