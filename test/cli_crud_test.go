package test

import (
	"testing"
	"github.com/fnproject/cli/test/cliharness"
	"fmt"
)

// TODO: These are both  Super minimal
func TestFnAppUpdateCycle(t *testing.T) {
	h := cliharness.Create(t)
	defer h.Cleanup()

	appName := h.NewAppName()

	// can't create an app twice
	h.Fn("apps", "create", appName).AssertSuccess()
	h.Fn("apps", "create", appName).AssertFailed()
	h.Fn("apps", "list", appName).AssertSuccess().AssertStdoutContains(appName)
	h.Fn("apps", "inspect", appName).AssertSuccess().AssertStdoutContains(fmt.Sprintf(`"name": "%s"`, appName))
	h.Fn("apps", "config", "set", appName, "fooConfig", "barval").AssertSuccess()
	h.Fn("apps", "config", "get", appName, "fooConfig").AssertSuccess().AssertStdoutContains("barval")
	h.Fn("apps", "config", "list", appName).AssertSuccess().AssertStdoutContains("fooConfig=barval")
	h.Fn("apps", "config", "unset", appName, "fooConfig").AssertSuccess()
	h.Fn("apps", "config", "get", appName, "fooConfig").AssertFailed()
	h.Fn("apps", "config", "list", appName).AssertSuccess().AssertStdoutEmpty()
}

func TestSimpleFnRouteUpdateCycle(t *testing.T) {
	h := cliharness.Create(t)
	defer h.Cleanup()
	appName1 := h.NewAppName()
	h.Fn("routes", "create", appName1, "myroute", "--image", "foo/duffimage:0.0.1").AssertSuccess()
	h.Fn("routes", "create", appName1, "myroute", "--image", "foo/duffimage:0.0.1").AssertFailed()
	h.Fn("routes", "inspect", appName1, "myroute").AssertSuccess().AssertStdoutContains(`"path": "/myroute"`)
	h.Fn("routes", "update", "-i", "bar/duffbeer:0.1.2", appName1, "myroute").AssertSuccess()
	h.Fn("routes", "config", "set", appName1, "myroute", "confA", "valB").AssertSuccess()
	h.Fn("routes", "config", "get", appName1, "myroute", "confA").AssertSuccess().AssertStdoutContains("valB")
	h.Fn("routes", "config", "list", appName1, "myroute").AssertSuccess().AssertStdoutContains("confA=valB")
	h.Fn("routes", "config", "unset", appName1, "myroute", "confA").AssertSuccess()
	h.Fn("routes", "config", "get", appName1, "myroute", "confA").AssertFailed()
}
