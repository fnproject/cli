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

func TestSimpleFnRouteUpdateCycle(t *testing.T) {
	t.Parallel()

	h := testharness.Create(t)
	defer h.Cleanup()
	appName1 := h.NewAppName()
	h.Fn("create", "route", appName1, "myroute", "foo/duffimage:0.0.1").AssertSuccess()
	h.Fn("create", "route", appName1, "myroute", "foo/duffimage:0.0.1").AssertFailed()
	h.Fn("inspect", "route", appName1, "myroute").AssertSuccess().AssertStdoutContains(`"path": "/myroute"`)
	h.Fn("update", "route", appName1, "myroute", "bar/duffbeer:0.1.2").AssertSuccess()
	h.Fn("config", "route", appName1, "myroute", "confA", "valB").AssertSuccess()
	h.Fn("get", "config", "route", appName1, "myroute", "confA").AssertSuccess().AssertStdoutContains("valB")
	h.Fn("list", "config", "route", appName1, "myroute").AssertSuccess().AssertStdoutContains("valB")
	h.Fn("unset", "config", "route", appName1, "myroute", "confA").AssertSuccess()
	h.Fn("get", "config", "route", appName1, "myroute", "confA").AssertFailed()
}

// func
func TestSimpleFnFunctionUpdateCycle(t *testing.T) {
	t.Parallel()

	h := testharness.Create(t)
	defer h.Cleanup()
	appName1 := h.NewAppName()
	funcName1 := h.NewFuncName()
	h.Fn("create", "function", appName1, funcName1, "foo/duffimage:0.0.1").AssertSuccess()
	h.Fn("create", "function", appName1, funcName1, "foo/duffimage:0.0.1").AssertFailed()
	h.Fn("inspect", "function", appName1, funcName1).AssertSuccess().AssertStdoutContains(`"path": "/myfunc"`)
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
	h.Fn("create", "trigger", appName1, funcName1, "mytrigger", "--type", "http", "--source", "mytrigger").AssertSuccess()
	h.Fn("create", "trigger", appName1, funcName1, "mytrigger", "--type", "http", "--source", "mytrigger").AssertFailed()
	h.Fn("inspect", "trigger", appName1, funcName1, "mytrigger").AssertSuccess().AssertStdoutContains(`"source": "/"mytrigger""`)
	h.Fn("update", "trigger", appName1, funcName1, "mytrigger", "--annnotation", `"val1='["val2"]'"`).AssertSuccess()
	h.Fn("config", "trigger", appName1, funcName1, "mytrigger", "confA", "valB").AssertSuccess()
}
func TestRemovingRouteAnnotation(t *testing.T) {
	t.Parallel()

	h := testharness.Create(t)
	defer h.Cleanup()
	appName1 := h.NewAppName()
	h.Fn("create", "route", appName1, "myroute", "foo/duffimage:0.0.1", "--annotation", "test=1").AssertSuccess()
	h.Fn("inspect", "route", appName1, "myroute").AssertSuccess().AssertStdoutContainsJSON([]string{"annotations", "test"}, 1.0)
	h.Fn("update", "route", appName1, "myroute", "foo/duffimage:0.0.1", "--annotation", `test=""`).AssertSuccess()
	h.Fn("inspect", "route", appName1, "myroute").AssertSuccess().AssertStdoutMissingJSONPath([]string{"annotations", "test"})
}

func TestInvalidAnnotationValue(t *testing.T) {
	t.Parallel()

	h := testharness.Create(t)
	defer h.Cleanup()
	appName1 := h.NewAppName()

	// The route should still be created, but without the invalid annotation
	h.Fn("create", "route", appName1, "myroute", "foo/duffimage:0.0.1", "--annotation", "test=value").AssertSuccess().AssertStderrContains("Unable to parse annotation value 'value'. Annotations values must be valid JSON strings.")
	h.Fn("inspect", "route", appName1, "myroute").AssertSuccess().AssertStdoutMissingJSONPath([]string{"annotations", "test"})
}

func TestRouteUpdateValues(t *testing.T) {
	t.Parallel()

	validCases := []struct {
		args   []string
		query  []string
		result interface{}
	}{
		{[]string{"--memory", "129"}, []string{"memory"}, 129.0},
		{[]string{"--type", "async"}, []string{"type"}, "async"},
		{[]string{"--headers", "foo=bar"}, []string{"headers", "foo", "0"}, "bar"},
		{[]string{"--format", "default"}, []string{"format"}, "default"},
		{[]string{"--timeout", "111"}, []string{"timeout"}, 111.0},
		{[]string{"--idle-timeout", "128"}, []string{"idle_timeout"}, 128.0},
		{[]string{"--annotation", "test=1"}, []string{"annotations", "test"}, 1.0},
	}

	for i, tcI := range validCases {
		tc := tcI
		t.Run(fmt.Sprintf("Valid Case %d", i), func(t *testing.T) {
			t.Parallel()
			h := testharness.Create(t)
			defer h.Cleanup()
			appName1 := h.NewAppName()
			h.Fn("create", "route", appName1, "myroute", "foo/someimage:0.0.1").AssertSuccess()

			h.Fn(append([]string{"update", "route", appName1, "myroute", "baz/fooimage:1.0.0"}, tc.args...)...).AssertSuccess()
			h.Fn("inspect", "route", appName1, "myroute").AssertSuccess().AssertStdoutContainsJSON(tc.query, tc.result)
		})
	}

	invalidCases := [][]string{
		{"--image", "fooimage:1.0.0"}, // image with no registry
		//	{"--memory", "0"},  bug?
		{"--memory", "wibble"},
		{"--type", "blancmange"},
		{"--headers", "potatocakes"},
		{"--format", "myharddisk"},
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
			h.Fn("create", "route", appName1, "myroute", "foo/someimage:0.0.1").AssertSuccess()

			h.Fn(append([]string{"update", "route", appName1, "myroute"}, tc...)...).AssertFailed()
		})
	}

}

func TestRoutesInspectWithPrefix(t *testing.T) {
	t.Parallel()

	h := testharness.Create(t)
	defer h.Cleanup()
	appName := h.NewAppName()
	h.Fn("routes", "create", appName, "myroute", "--image", "foo/someimage:0.0.1").AssertSuccess()
	h.Fn("routes", "inspect", appName, "myroute").AssertSuccess()
	h.Fn("routes", "inspect", appName, "/myroute").AssertSuccess()

}
