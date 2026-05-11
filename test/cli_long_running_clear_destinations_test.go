package test

import (
	"testing"

	"github.com/fnproject/cli/testharness"
)

func TestUpdateDestinationClearFlagsConflict(t *testing.T) {
	t.Parallel()

	h := testharness.Create(t)
	defer h.Cleanup()

	res := h.Fn("update", "function", "dummy-app", "dummy-fn",
		"--on-success", "stream:ocid1.stream.oc1..abc",
		"--clear-on-success")
	if res.Success {
		t.Fatal("expected update to fail when both --on-success and --clear-on-success are provided")
	}
	res.AssertStderrContains("--on-success and --clear-on-success cannot be used together")
}