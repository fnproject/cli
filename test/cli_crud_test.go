package test

import (
	"fmt"
	"testing"

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
	funcName1 := h.NewFuncName()
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
	funcName1 := h.NewFuncName()
	h.Fn("create", "app", appName1).AssertSuccess()
	h.Fn("create", "function", appName1, funcName1, "foo/duffimage:0.0.1").AssertSuccess()
	h.Fn("create", "trigger", appName1, funcName1, "mytrigger", "--type", "http", "--source", "/mytrigger").AssertSuccess()
	h.Fn("create", "trigger", appName1, funcName1, "mytrigger", "--type", "http", "--source", "/mytrigger").AssertFailed()
	h.Fn("inspect", "trigger", appName1, funcName1, "mytrigger").AssertSuccess().AssertStdoutContains(`"source": "/mytrigger`)
	h.Fn("update", "trigger", appName1, funcName1, "mytrigger", "--annotation", `"val1='["val2"]'"`).AssertSuccess()
	h.Fn("config", "trigger", appName1, funcName1, "mytrigger", "confA", "valB").AssertSuccess()
}
