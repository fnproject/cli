package test

import (
	"github.com/fnproject/cli/testharness"
	"testing"
)

func TestContextCrud(t *testing.T) {
	t.Parallel()
	h := testharness.Create(t)
	defer h.Cleanup()

	h.Fn("context", "list").AssertSuccess().AssertStdoutContains("default")
	h.Fn("context", "create", "--api-url", "http://alskjalskdjasd/v1", "mycontext").AssertSuccess()
	h.Fn("context", "use", "mycontext").AssertSuccess()
	h.Fn("context", "delete", "mycontext").AssertFailed()
	h.Fn("context", "unset", "mycontext").AssertSuccess()
	h.Fn("context", "delete", "mycontext").AssertSuccess()

}
