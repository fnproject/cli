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
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

type NodeLangHelper struct {
	BaseHelper
	Version string
}

func (h *NodeLangHelper) Handles(lang string) bool {
	return defaultHandles(h, lang)
}
func (h *NodeLangHelper) Runtime() string {
	return h.LangStrings()[0]
}

func (lh *NodeLangHelper) LangStrings() []string {
	return []string{"node", fmt.Sprintf("node%s", lh.Version)}
}
func (lh *NodeLangHelper) Extensions() []string {
	// this won't be chosen by default
	return []string{".js"}
}
func (lh *NodeLangHelper) BuildFromImage() (string, error) {
	return fmt.Sprintf("fnproject/node:%s-dev", lh.Version), nil
}
func (lh *NodeLangHelper) RunFromImage() (string, error) {
	return fmt.Sprintf("fnproject/node:%s", lh.Version), nil
}

const funcJsContent = `const fdk=require('@fnproject/fdk');

fdk.handle(function(input){
  let name = 'World';
  if (input.name) {
    name = input.name;
  }
  console.log('\nInside Node Hello World function')
  return {'message': 'Hello ' + name}
})
`

const packageJsonContent = `{
	"name": "hellofn",
    "version": "1.0.0",
	"description": "example function",
	"main": "func.js",
	"author": "",
	"license": "Apache-2.0",
	"dependencies": {
		"@fnproject/fdk": ">=%s"
	}
}
`

func (h *NodeLangHelper) GenerateBoilerplate(path string) error {
	fdkVersion, err := h.GetLatestFDKVersion()
	if err != nil {
		return err
	}

	pathToPackageJsonFile := filepath.Join(path, "package.json")
	pathToFuncJs := filepath.Join(path, "func.js")

	if exists(pathToPackageJsonFile) || exists(pathToFuncJs) {
		return ErrBoilerplateExists
	}

	packageJsonBoilerplate := fmt.Sprintf(packageJsonContent, fdkVersion)
	err = ioutil.WriteFile(pathToPackageJsonFile, []byte(packageJsonBoilerplate), os.FileMode(0644))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(pathToFuncJs, []byte(funcJsContent), os.FileMode(0644))
	if err != nil {
		return err
	}

	return nil
}

func (h *NodeLangHelper) HasBoilerplate() bool { return true }

// CustomMemory - no memory override here.
func (h *NodeLangHelper) CustomMemory() uint64 { return 0 }

func (h *NodeLangHelper) Entrypoint() (string, error) {
	return "node func.js", nil
}

func (h *NodeLangHelper) DockerfileBuildCmds() []string {
	r := []string{}
	// skip npm -install if node_modules is local - allows local development
	if exists("dist") {
		// Do nothing
	} else if exists("package.json") && !exists("node_modules") {
		if exists("package-lock.json") {
			r = append(r, "ADD package-lock.json /function/")
		}

		r = append(r,
			"ADD package.json /function/",
			"RUN npm install",
		)
	}
	return r
}

func (h *NodeLangHelper) DockerfileCopyCmds() []string {
	// If they have a dist folder (likely from webpack) let's just include that
	if exists("dist") {
		r := []string{"ADD dist/main.js /function/func.js"}
		return r
	}
	// excessive but content could be anything really

	r := []string{"ADD . /function/"}
	if exists("package.json") && !exists("node_modules") {
		r = append(r, "COPY --from=build-stage /function/node_modules/ /function/node_modules/")
	}
	r = append(r, "RUN chmod -R o+r /function")

	return r
}

func (h *NodeLangHelper) GetLatestFDKVersion() (string, error) {

	const versionURL = "https://registry.npmjs.org/@fnproject/fdk"
	const versionEnv = "FN_NODE_FDK_VERSION"
	fetchError := fmt.Errorf("failed to fetch latest Node FDK version from %v. "+
		"Check your network settings or manually override the Node FDK version by setting %s", versionURL, versionEnv)

	version := os.Getenv(versionEnv)
	if version != "" {
		return version, nil
	}

	resp, err := http.Get(versionURL)
	if err != nil || resp.StatusCode != 200 {
		return "", fetchError
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fetchError
	}

	parsedResp := struct {
		DistTags struct {
			Latest string `json:"latest"`
		} `json:"dist-tags"`
	}{}
	err = json.Unmarshal(body, &parsedResp)
	if err != nil {
		return "", fetchError
	}

	return parsedResp.DistTags.Latest, nil
}

func (h *NodeLangHelper) FixImagesOnInit() bool {
	return true
}
