package test

import (
	"fmt"
	"github.com/fnproject/cli/testharness"
	"testing"
)

// TODO: These are both  Super minimal
func TestFnAppUpdateCycle(t *testing.T) {
	t.Parallel()

	h := testharness.Create(t)
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
	t.Parallel()

	h := testharness.Create(t)
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

func TestRemovingRouteAnnotation(t *testing.T) {
	t.Parallel()

	h := testharness.Create(t)
	defer h.Cleanup()
	appName1 := h.NewAppName()
	h.Fn("routes", "create", appName1, "myroute", "--image", "foo/duffimage:0.0.1", "--annotation", "test=1").AssertSuccess()
	h.Fn("routes", "inspect", appName1, "myroute").AssertSuccess().AssertStdoutContainsJSON([]string{"annotations", "test"}, 1.0)
	h.Fn("routes", "update", appName1, "myroute", "--image", "foo/duffimage:0.0.1", "--annotation", `test=""`).AssertSuccess()
	h.Fn("routes", "inspect", appName1, "myroute").AssertSuccess().AssertStdoutMissingJSONPath([]string{"annotations", "test"})
}

func TestRouteUpdateValues(t *testing.T) {
	t.Parallel()

	validCases := []struct {
		args   []string
		query  []string
		result interface{}
	}{
		{[]string{"--image", "baz/fooimage:1.0.0"}, []string{"image"}, "baz/fooimage:1.0.0"},
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
			h.Fn("routes", "create", appName1, "myroute", "--image", "foo/someimage:0.0.1").AssertSuccess()

			h.Fn(append([]string{"routes", "update", appName1, "myroute"}, tc.args...)...).AssertSuccess()
			h.Fn("routes", "inspect", appName1, "myroute").AssertSuccess().AssertStdoutContainsJSON(tc.query, tc.result)
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
			h.Fn("routes", "create", appName1, "myroute", "--image", "foo/someimage:0.0.1").AssertSuccess()

			h.Fn(append([]string{"routes", "update", appName1, "myroute"}, tc...)...).AssertFailed()
		})
	}

}
