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

	"regexp"
	"strings"

	"github.com/fnproject/cli/testharness"
)

// TODO: These are both  Super minimal
func TestFnAppUpdateCycle(t *testing.T) {
	t.Parallel()

	h := testharness.Create(t)
	defer h.Cleanup()

	appName := h.NewAppName()

	// can't create an app twice
	h.Fn("create", "app", appName).AssertSuccess()
	h.Fn("create", "app", appName).AssertFailed()
	h.Fn("list", "apps", appName).AssertSuccess().AssertStdoutContains(appName)
	// Test looking up app by name when multiple pages worth of apps exist
	for i := 0; i < 50; i++ {
		h.Fn("create", "app", fmt.Sprintf("%s%d", appName, i)).AssertSuccess()
	}
	h.Fn("inspect", "app", appName).AssertSuccess().AssertStdoutContains(fmt.Sprintf(`"name": "%s"`, appName))
	h.Fn("config", "app", appName, "fooConfig", "barval").AssertSuccess()
	h.Fn("get", "config", "app", appName, "fooConfig").AssertSuccess().AssertStdoutContains("barval")
	h.Fn("list", "config", "app", appName).AssertSuccess().AssertStdoutContains("barval")
	h.Fn("unset", "config", "app", appName, "fooConfig").AssertSuccess()
	h.Fn("get", "config", "app", appName, "fooConfig").AssertFailed()
	h.Fn("list", "config", "app", appName).AssertSuccess().AssertStdoutEmpty()
	h.Fn("delete", "app", appName).AssertSuccess()
}

// func
func TestSimpleFnFunctionUpdateCycle(t *testing.T) {
	t.Parallel()

	h := testharness.Create(t)
	defer h.Cleanup()
	appName1 := h.NewAppName()
	funcName1 := h.NewFuncName(appName1)
	h.Fn("create", "function", appName1, funcName1, "foo/duffimage:0.0.1").AssertFailed()
	h.Fn("create", "app", appName1).AssertSuccess()
	h.Fn("create", "function", appName1, funcName1, "foo/duffimage:0.0.1").AssertSuccess()
	h.Fn("create", "function", appName1, funcName1, "foo/duffimage:0.0.1").AssertFailed()
	h.Fn("inspect", "function", appName1, funcName1).AssertSuccess().AssertStdoutContains(`"image": "foo/duffimage:0.0.1"`)
	h.Fn("update", "function", appName1, funcName1, "bar/duffbeer:0.1.2").AssertSuccess()
	h.Fn("config", "function", appName1, funcName1, "confA", "valB").AssertSuccess()
	h.Fn("get", "config", "function", appName1, funcName1, "confA").AssertSuccess().AssertStdoutContains("valB")
	h.Fn("list", "config", "function", appName1, funcName1).AssertSuccess().AssertStdoutContains("valB")
	h.Fn("unset", "config", "function", appName1, funcName1, "confA").AssertSuccess()
	h.Fn("get", "config", "function", appName1, funcName1, "confA").AssertFailed()
}

// triggers
func TestSimpleFnTriggerUpdateCycle(t *testing.T) {
	t.Parallel()

	h := testharness.Create(t)
	defer h.Cleanup()
	appName1 := h.NewAppName()
	funcName1 := h.NewFuncName(appName1)
	triggerName1 := h.NewTriggerName(appName1, funcName1)
	h.Fn("create", "trigger", appName1, funcName1, triggerName1).AssertFailed()
	h.Fn("create", "app", appName1).AssertSuccess()
	h.Fn("create", "trigger", appName1, funcName1, triggerName1).AssertFailed()
	h.Fn("create", "function", appName1, funcName1, "foo/duffimage:0.0.1").AssertSuccess()
	h.Fn("create", "trigger", appName1, funcName1, triggerName1, "--type", "http", "--source", "/mytrigger").AssertSuccess()
	h.Fn("create", "trigger", appName1, funcName1, triggerName1, "--type", "http", "--source", "/mytrigger").AssertFailed()
	h.Fn("inspect", "trigger", appName1, funcName1, triggerName1).AssertSuccess().AssertStdoutContains(`"source": "/mytrigger`)
	h.Fn("update", "trigger", appName1, funcName1, triggerName1, "--annotation", `"val1='["val2"]'"`).AssertSuccess()
}

func TestRemovingFnAnnotation(t *testing.T) {
	t.Parallel()

	h := testharness.Create(t)
	defer h.Cleanup()
	appName1 := h.NewAppName()
	funcName1 := h.NewFuncName(appName1)
	h.Fn("create", "app", appName1).AssertSuccess()
	h.Fn("create", "fn", appName1, funcName1, "foo/duffimage:0.0.1", "--annotation", "test=1").AssertSuccess()
	h.Fn("inspect", "fn", appName1, funcName1).AssertSuccess().AssertStdoutContainsJSON([]string{"annotations", "test"}, 1.0)
	h.Fn("update", "fn", appName1, funcName1, "foo/duffimage:0.0.1", "--annotation", `test=""`).AssertSuccess()
	h.Fn("inspect", "fn", appName1, funcName1).AssertSuccess().AssertStdoutMissingJSONPath([]string{"annotations", "test"})
}

func TestInvalidFnAnnotationValue(t *testing.T) {
	t.Parallel()

	h := testharness.Create(t)
	defer h.Cleanup()
	appName1 := h.NewAppName()
	funcName1 := h.NewFuncName(appName1)

	h.Fn("create", "app", appName1).AssertSuccess()
	h.Fn("create", "fn", appName1, funcName1, "foo/duffimage:0.0.1", "--annotation", "test=value").AssertSuccess().AssertStderrContains("Unable to parse annotation value 'value'. Annotations values must be valid JSON strings.")
	h.Fn("inspect", "fn", appName1, funcName1).AssertSuccess().AssertStdoutMissingJSONPath([]string{"annotations", "test"})
}

func TestFnUpdateValues(t *testing.T) {
	t.Parallel()

	validCases := []struct {
		args   []string
		query  []string
		result interface{}
	}{
		{[]string{"--memory", "129"}, []string{"memory"}, 129.0},
		{[]string{"--timeout", "111"}, []string{"timeout"}, 111.0},
		{[]string{"--idle-timeout", "128"}, []string{"idle_timeout"}, 128.0},
		{[]string{"--config", "test=val"}, []string{"config", "test"}, "val"},
		{[]string{"--annotation", "test=1"}, []string{"annotations", "test"}, 1.0},
		{[]string{"--image", "fnproject/blah-blah:0.1.0"}, []string{"image"}, "fnproject/blah-blah:0.1.0"},
	}

	for i, tcI := range validCases {
		tc := tcI
		t.Run(fmt.Sprintf("Valid Case %d", i), func(t *testing.T) {
			t.Parallel()
			h := testharness.Create(t)
			defer h.Cleanup()
			appName1 := h.NewAppName()
			funcName1 := h.NewFuncName(appName1)
			h.Fn("create", "app", appName1)
			h.Fn("create", "fn", appName1, funcName1, "foo/someimage:0.0.1").AssertSuccess()

			h.Fn(append([]string{"update", "fn", appName1, funcName1, "baz/fooimage:1.0.0"}, tc.args...)...).AssertSuccess()
			h.Fn("inspect", "fn", appName1, funcName1).AssertSuccess().AssertStdoutContainsJSON(tc.query, tc.result)
		})
	}

	invalidCases := [][]string{
		// image with no registry is valid case for local development
		// {"--image", "fooimage:1.0.0"}, // image with no registry
		//	{"--memory", "0"},  bug?
		{"--memory", "wibble"},
		{"--type", "blancmange"},
		{"--headers", "potatocakes"},
		{"--timeout", "86400"},
		{"--timeout", "sit_in_the_corner"},
		{"--idle-timeout", "86000"},
		{"--idle-timeout", "yawn"},
	}

	for i, tcI := range invalidCases {
		tc := tcI
		t.Run(fmt.Sprintf("Invalid Case %d", i), func(t *testing.T) {
			t.Parallel()
			h := testharness.Create(t)
			defer h.Cleanup()
			appName1 := h.NewAppName()
			funcName1 := h.NewFuncName(appName1)
			h.Fn("create", "app", appName1)
			h.Fn("create", "fn", appName1, funcName1, "foo/someimage:0.0.1").AssertSuccess()

			h.Fn(append([]string{"update", "fn", appName1, funcName1}, tc...)...).AssertFailed()
		})
	}

}

func TestInspectEndpoints(t *testing.T) {

	h := testharness.Create(t)
	defer h.Cleanup()
	appName1 := h.NewAppName()
	funcName1 := h.NewFuncName(appName1)
	h.Fn("create", "app", appName1)
	h.Fn("create", "fn", appName1, funcName1, "foo/someimage:0.0.1").AssertSuccess()
	h.Fn("create", "trigger", appName1, funcName1, "t1", "--type", "http", "--source", "/trig").AssertSuccess()

	res := h.Fn("inspect", "function", appName1, funcName1, "id").AssertSuccess()
	fnId := strings.Trim(strings.TrimSpace(res.Stdout), "\"")

	res = h.Fn("inspect", "function", appName1, funcName1, "--endpoint").AssertSuccess()
	invokeUrl := strings.TrimSpace(res.Stdout)

	invokePattern := regexp.MustCompile("^http://.*/invoke/" + regexp.QuoteMeta(fnId) + "$")

	if !invokePattern.MatchString(invokeUrl) {
		t.Errorf("Expected invoke URL matching %s, got %s", invokePattern, invokeUrl)
	}

	res = h.Fn("inspect", "trigger", appName1, funcName1, "t1", "--endpoint").AssertSuccess()

	triggerUrl := strings.TrimSpace(res.Stdout)
	triggerPattern := regexp.MustCompile("^http://.*/t/" + regexp.QuoteMeta(appName1) + "/trig$")

	if !triggerPattern.MatchString(triggerUrl) {
		t.Errorf("Expected trigger URL matching %s, got %s", triggerPattern, triggerUrl)
	}

}

func TestEmptyConfigs(t *testing.T) {
	h := testharness.Create(t)
	defer h.Cleanup()
	appName1 := h.NewAppName()
	funcName1 := h.NewFuncName(appName1)

	h.Fn("create", "app", appName1)
	h.Fn("create", "fn", appName1, funcName1, "foo/someimage:0.0.1").AssertSuccess()
	//begin tests
	//get config
	h.Fn("get", "config", "app", appName1, "nonexistantKey").AssertFailed()
	h.Fn("get", "config", "function", appName1, funcName1, "nonexistantKey").AssertFailed()
	//list config
	h.Fn("list", "config", "app", appName1).AssertSuccess()
	h.Fn("list", "config", "function", appName1, funcName1).AssertSuccess()
	//unset config
	h.Fn("unset", "config", "app", appName1, "nonexistantKey").AssertSuccess()
	h.Fn("unset", "config", "function", appName1, funcName1, "nonexistantKey").AssertSuccess()
}

func TestRecursiveDelete(t *testing.T) {
	t.Parallel()

	h := testharness.Create(t)
	defer h.Cleanup()
	appName1 := h.NewAppName()
	funcName1 := h.NewFuncName(appName1)
	triggerName1 := h.NewTriggerName(appName1, funcName1)

	h.Fn("create", "app", appName1).AssertSuccess()
	h.Fn("create", "function", appName1, funcName1, "foo/duffimage:0.0.1").AssertSuccess()
	h.Fn("create", "trigger", appName1, funcName1, triggerName1, "--type", "http", "--source", "/mytrigger").AssertSuccess()
	h.Fn("delete", "app", appName1, "-f", "-r").AssertSuccess().
		AssertStdoutContains(appName1).
		AssertStdoutContains(funcName1).
		AssertStdoutContains(triggerName1)
}

func TestFnAppUpdateCycleWithArch(t *testing.T) {
	t.Parallel()

	h := testharness.Create(t)
	defer h.Cleanup()

	appName := h.NewAppName()

	// can't create an app twice
	h.Fn("create", "app", appName, "--shape", "GENERIC_X86").AssertSuccess()
	h.Fn("inspect", "app", appName).AssertSuccess().AssertStdoutContains("GENERIC_X86")

	appName = h.NewAppName()
	h.Fn("create", "app", appName).AssertSuccess()
	h.Fn("inspect", "app", appName).AssertSuccess().AssertStdoutContains("GENERIC_X86")

	//Test update to multiarch, should fail as we don't allow any update operation

	shape := "GENERIC_X86_ARM"
	h.Fn("update", "app", appName, "--shape", shape).AssertFailed()

	//Test update back to arm, should fail
	h.Fn("update", "app", appName, "--shape", "GENERIC_ARM").AssertFailed()
}