package test

import (
	"fmt"
	"github.com/fnproject/cli/testharness"
	"os"
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
	h.Fn("invoke", appName1, funcName1).AssertFailed()
}

func TestFnInvokeValidImage(t *testing.T) {
	t.Parallel()

	h := testharness.Create(t)
	defer h.Cleanup()
	appName1 := h.NewAppName()
	funcName1 := h.NewFuncName(appName1)
	h.Fn("create", "app", appName1).AssertSuccess()
	h.Fn("create", "function", appName1, funcName1, "fnproject/hello:latest").AssertSuccess()
	h.Fn("invoke", appName1, funcName1).AssertSuccess()
}

// test fn list id value matches fn inspect id value for app, function & trigger
func TestListIDValue(t *testing.T) {
	t.Parallel()

	h := testharness.Create(t)
	defer h.Cleanup()

	appName := h.NewAppName()
	funcName := h.NewFuncName(appName)
	triggerName := h.NewTriggerName(appName, funcName)
	h.Fn("create", "app", appName).AssertSuccess()
	h.Fn("create", "function", appName, funcName, "foo/duffimage:0.0.1").AssertSuccess()
	h.Fn("create", "trigger", appName, funcName, triggerName, "--type", "http", "--source", "/mytrigger").AssertSuccess()

	res := h.Fn("list", "apps", appName).AssertSuccess()
	// get app id, compare app id
	//appID :=
	t.Errorf("test..........")
	fmt.Fprintf(os.Stdout, "RESULT FROM LIST APPS \n")
	fmt.Fprintf(os.Stdout, res.String())

	//h.Fn("inspect", "app", appName).AssertSuccess().AssertStdoutContains(fmt.Sprintf(`"id": "%s"`, appID))

}
