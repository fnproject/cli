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
	"testing"

	"github.com/fnproject/cli/testharness"
)

var Runtimes = []struct {
	runtime   string
	callInput string
}{
	{"go", ""},
	{"dotnet", ""},
	{"dotnet3.1", ""},
	{"dotnet6.0", ""},
	{"java", ""},
	{"java8", ""},
	{"java11", ""},
	{"kotlin", ""},
	{"node", ""},
	{"ruby", ""},
	{"python", ""},
	{"python3.6", ""},
	{"python3.7", ""},
	{"python3.8", ""},
}

func TestFnInitWithBoilerplateBuildsRuns(t *testing.T) {
	t.Parallel()

	for _, runtimeI := range Runtimes {
		rt := runtimeI
		t.Run(fmt.Sprintf("%s runtime", rt.runtime), func(t *testing.T) {
			t.Parallel()
			h := testharness.Create(t)
			defer h.Cleanup()

			appName := h.NewAppName()
			h.Fn("create", "app", appName).AssertSuccess()
			funcName := h.NewFuncName(appName)

			h.Fn("init", "--runtime", rt.runtime, funcName).AssertSuccess()

			h.Cd(funcName)
			h.Fn("build").AssertSuccess()

			h.Fn("--registry", "test", "deploy", "--local", "--app", appName).AssertSuccess()

			h.FnWithInput(rt.callInput, "invoke", appName, funcName).AssertSuccess()
		})
	}

}
