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

package langs

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

// not a map because some helpers can handle multiple keys
var helpers = []LangHelper{}
var fallBackOlderVersions = map[string]LangHelper{}

func init() {
	registerHelper(&DotnetLangHelper{Version: "8.0"})
	registerHelper(&DotnetLangHelper{Version: "6.0"})
	registerHelper(&DotnetLangHelper{Version: "3.1"})
	registerHelper(&GoLangHelper{Version: "1.20"})
	registerHelper(&GoLangHelper{Version: "1.19"})
	// order matter, 'java' will pick up the first JavaLangHelper
	registerHelper(&JavaLangHelper{version: "17"})
	registerHelper(&JavaLangHelper{version: "11"})
	registerHelper(&JavaLangHelper{version: "8"})
	registerHelper(&NodeLangHelper{Version: "20"})
	registerHelper(&NodeLangHelper{Version: "18"})
	// order matter, 'python' will pick up the first PythonLangHelper
	registerHelper(&PythonLangHelper{Version: "3.11"})
	registerHelper(&PythonLangHelper{Version: "3.9"})
	registerHelper(&PythonLangHelper{Version: "3.8"})
	registerHelper(&PythonLangHelper{Version: "3.8.5"})

	//New runtime support for Ruby 3.1
	// order matter, 'ruby' will pick up the first RubyLangHelper
	registerHelper(&RubyLangHelper{Version: "3.1"})

	registerHelper(&KotlinLangHelper{})

	// for older versions support backwards compatibility
	fallBackOlderVersions["ruby"] = &RubyLangHelper{Version: "3.1"}
	fallBackOlderVersions["node"] = &NodeLangHelper{Version: "16"}
	fallBackOlderVersions["go"] = &GoLangHelper{Version: "1.11"}
}

func registerHelper(h LangHelper) {
	helpers = append(helpers, h)
}

func Helpers() []LangHelper {
	return helpers
}

var (
	ErrBoilerplateExists = errors.New("Function boilerplate already exists")
)

// GetLangHelper returns a LangHelper for the passed in language
func GetLangHelper(lang string) LangHelper {
	for _, h := range helpers {
		if h.Handles(lang) {
			return h
		}
	}
	return nil
}

func GetFallbackLangHelper(lang string) LangHelper {
	return fallBackOlderVersions[lang]
}

func IsFallbackSupported(lang string) bool {
	_, found := fallBackOlderVersions[lang]
	return found
}

// LangHelper is the interface that language helpers must implement.
type LangHelper interface {
	// Handles return whether it can handle the passed in lang string or not
	Handles(string) bool
	// LangStrings returns list of supported language strings user can use for runtime
	LangStrings() []string
	// Extension is the file extension this helper supports. Eg: .java, .go, .js
	Extensions() []string
	// Runtime that will be used for the build (includes version)
	Runtime() string
	// BuildFromImage is the base image to build off, typically fnproject/LANG:dev
	BuildFromImage() (string, error)
	// RunFromImage is the base image to use for deployment (usually smaller than the build images)
	RunFromImage() (string, error)
	// If set to false, it will use a single Docker build step, rather than multi-stage
	IsMultiStage() bool
	// Dockerfile build lines for building dependencies or anything else language specific
	DockerfileBuildCmds() []string
	// DockerfileCopyCmds will run in second/final stage of multi-stage build to copy artifacts form the build stage
	DockerfileCopyCmds() []string
	// Entrypoint sets the Docker Entrypoint. One of Entrypoint or Cmd is required.
	Entrypoint() (string, error)
	// Cmd sets the Docker command. One of Entrypoint or Cmd is required.
	Cmd() (string, error)
	// CustomMemory allows a helper to specify a base memory amount, return "" to leave unspecified and let the runtime decide.
	CustomMemory() uint64
	HasPreBuild() bool
	PreBuild() error
	AfterBuild() error
	// HasBoilerplate indicates whether a language has support for generating function boilerplate.
	HasBoilerplate() bool
	// GenerateBoilerplate generates basic function boilerplate. Returns ErrBoilerplateExists if the function file
	// already exists.
	GenerateBoilerplate(string) error
	// FixImagesOnInit determines if images should be fixed on initialization - BuildFromImage and RunFromImage will be written to func.yaml
	FixImagesOnInit() bool
	// GetLatestFDKVersion checks the package repository and returns the latest version of FDK version if available.
	GetLatestFDKVersion() (string, error)
}

func defaultHandles(h LangHelper, lang string) bool {
	for _, s := range h.LangStrings() {
		if lang == s {
			return true
		}
	}
	return false
}

// BaseHelper is empty implementation of LangHelper for embedding in implementations.
type BaseHelper struct {
}

func (h *BaseHelper) IsMultiStage() bool                   { return true }
func (h *BaseHelper) DockerfileBuildCmds() []string        { return []string{} }
func (h *BaseHelper) DockerfileCopyCmds() []string         { return []string{} }
func (h *BaseHelper) Entrypoint() (string, error)          { return "", nil }
func (h *BaseHelper) Cmd() (string, error)                 { return "", nil }
func (h *BaseHelper) HasPreBuild() bool                    { return false }
func (h *BaseHelper) PreBuild() error                      { return nil }
func (h *BaseHelper) AfterBuild() error                    { return nil }
func (h *BaseHelper) HasBoilerplate() bool                 { return false }
func (h *BaseHelper) GenerateBoilerplate(string) error     { return nil }
func (h *BaseHelper) CustomMemory() uint64                 { return 0 }
func (h *BaseHelper) FixImagesOnInit() bool                { return false }
func (h *BaseHelper) GetLatestFDKVersion() (string, error) { return "", nil }

// exists checks if a file exists
func exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func mkdirAndWriteFile(path, dir, filename, content string) error {
	fullPath := filepath.Join(path, dir)
	if err := os.MkdirAll(fullPath, os.FileMode(0755)); err != nil {
		return err
	}

	fullFilePath := filepath.Join(fullPath, filename)
	return ioutil.WriteFile(fullFilePath, []byte(content), os.FileMode(0644))
}

type githubTagResponse struct {
	Name string `json:"name"`
}

func getLatestFDKVersionFromGithub(repoKey string) (string, error) {
	// Github API has limit on number of calls
	resp, err := http.Get(fmt.Sprintf("https://api.github.com/repos/%s/tags", repoKey))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	responseBody := []githubTagResponse{}
	if err = json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return "", err
	}
	if len(responseBody) == 0 {
		return "", errors.New("Could not read latest version of FDK from tags")
	}
	return responseBody[0].Name, nil
}
