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

package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/fnproject/cli/common"
	"github.com/fnproject/cli/langs"
	"github.com/fnproject/cli/testharness"
)

const (
	rubySrcBoilerplate = `require 'fdk'
def myfunction(context:, input:)
  "ruby#{RUBY_VERSION}"
end

FDK.handle(target: :myfunction)
`

	rubyGemfileBoilerplate = `source 'https://rubygems.org' do
  gem 'fdk', '>= %s'
end
`
	funcYamlContent = `schema_version: 20180708
name: %s 
version: 0.0.1
runtime: ruby 
entrypoint: ruby func.rb`
)

/*
This test case check for backwards compatibility with older cli func.yaml file
Cases Tested:
 1. During `fn build` make sure Build_image and Run_image are stamped in func.yaml file
 2. Function container is build using proper fallback runtime and dev image, check
    by invoking container and fetching runtime version, should match with fallback version.
*/
func TestFnBuildWithOlderRuntimeWithoutVersion(t *testing.T) {
	t.Run("`fn invoke` should return the fallback ruby version", func(t *testing.T) {
		t.Parallel()
		h := testharness.Create(t)
		defer h.Cleanup()

		appName := h.NewAppName()
		funcName := h.NewFuncName(appName)
		dirName := funcName + "_dir"
		fmt.Println(appName + " " + funcName)
		h.Fn("create", "app", appName).AssertSuccess()
		h.Fn("init", "--runtime", "ruby", "--name", funcName, dirName).AssertSuccess()

		// change dir to newly created function dir
		h.Cd(dirName)

		// write custom function file which returns runtime version
		h.WithFile("func.rb", rubySrcBoilerplate, 0644)

		// inject func name in yaml file placeholder
		oldClientYamlFile := fmt.Sprintf(funcYamlContent, funcName)

		//update back yaml file
		h.WithFile("func.yaml", oldClientYamlFile, 0644)

		fallBackHandler := langs.GetFallbackLangHelper("ruby")
		fallBackVersion := fallBackHandler.LangStrings()[1]

		h.Fn("--verbose", "build").AssertSuccess()

		bi, err := fallBackHandler.BuildFromImage()
		if err != nil {
			panic(err)
		}
		bi = common.AddContainerNamespace(bi)

		ri, err := fallBackHandler.RunFromImage()
		if err != nil {
			panic(err)
		}
		ri = common.AddContainerNamespace(ri)
		fmt.Println(bi)
		fmt.Println(ri)

		updatedFuncFile := h.GetYamlFile("func.yaml")
		fmt.Println(updatedFuncFile.Build_image)
		fmt.Printf(updatedFuncFile.Run_image)
		if bi == "" || ri == "" {
			err := "Build_image or Run_image property is not set in func.yaml file"
			panic(err)
		}

		//Test whether build_image set in func.yaml is correct or not
		if bi != updatedFuncFile.Build_image {
			err := fmt.Sprintf("Expected Build image %s and Build image in func.yaml do not match%s", bi, updatedFuncFile.Build_image)
			panic(err)
		}

		//Test whether run_image set in func.yaml is correct or not
		if ri != updatedFuncFile.Run_image {
			err := fmt.Sprintf("Expected Run image %s and Run image in func.yaml do not match%s", ri, updatedFuncFile.Run_image)
			panic(err)
		}

		h.Fn("--registry", "test", "deploy", "--local", "--app", appName).AssertSuccess()
		result := h.Fn("invoke", appName, funcName).AssertSuccess()

		// // get the returned version from ruby image
		imageVersion := result.Stdout

		fmt.Println("Ruby version returned by image :" + imageVersion)
		fmt.Println("Fallback ruby version :" + fallBackVersion)
		match := strings.Contains(imageVersion, fallBackVersion)
		if !match {
			err := fmt.Sprintf("Versions do not match, `ruby` image version %s does not match with fallback version %s", imageVersion, fallBackVersion)
			panic(err)
		}
	})
}
