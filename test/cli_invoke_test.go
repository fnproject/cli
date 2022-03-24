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
	"github.com/fnproject/cli/testharness"
	"strings"
	"testing"
)

func TestFnInvokeInvalidImage(t *testing.T) {
	t.Parallel()

	h := testharness.Create(t)
	defer h.Cleanup()

	appName1 := h.NewAppName()
	funcName1 := h.NewFuncName(appName1)
	h.Fn("create", "app", appName1).AssertSuccess()
	h.Fn("create", "function", appName1, funcName1, "foo/duffimage:0.0.1").AssertSuccess()
	h.Fn("invoke", appName1, funcName1).AssertStdoutContains("Failed to pull image")
}

func TestFnInvokeValidImage(t *testing.T) {
	t.Parallel()

	h := testharness.Create(t)
	defer h.Cleanup()

	h.MkDir("fn")
	h.Cd("fn")
	withGoFunction(h)
	h.WithFile("Dockerfile", dockerFile, 0644)
	h.Docker("build", "-t", "fnproject/hello:latest", ".").AssertSuccess()

	h.Cd("")
	appName1 := h.NewAppName()
	funcName1 := h.NewFuncName(appName1)
	h.Fn("create", "app", appName1).AssertSuccess()
	h.Fn("create", "function", appName1, funcName1, "fnproject/hello:latest").AssertSuccess()
	h.Fn("invoke", appName1, funcName1).AssertSuccess()
}

func TestFnInvokeViaDirectUrl(t *testing.T) {
	t.Parallel()

	h := testharness.Create(t)
	defer h.Cleanup()

	h.MkDir("fn")
	h.Cd("fn")
	withGoFunction(h)
	h.WithFile("Dockerfile", dockerFile, 0644)
	h.Docker("build", "-t", "fnproject/hello:latest", ".").AssertSuccess()

	appName1 := h.NewAppName()
	funcName1 := h.NewFuncName(appName1)
	h.Fn("create", "app", appName1).AssertSuccess()
	h.Fn("create", "function", appName1, funcName1, "fnproject/hello:latest").AssertSuccess()

	res := h.Fn("inspect", "function", appName1, funcName1, "--endpoint").AssertSuccess()

	url := strings.TrimSpace(res.Stdout)

	h.Fn("invoke", "--endpoint", url).AssertSuccess()
}
