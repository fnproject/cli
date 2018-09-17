package test

import (
	"testing"

	"github.com/fnproject/cli/testharness"
)

func TestContextCrud(t *testing.T) {
	t.Parallel()
	h := testharness.Create(t)
	defer h.Cleanup()

	h.Fn("list", "context").AssertSuccess().AssertStdoutContains("default")
	h.Fn("create", "context", "--api-url", "alskjalskdjasd/v2", "mycontext").AssertFailed()
	h.Fn("create", "context", "--api-url", "http://alskjalskdjasd", "mycontext").AssertSuccess()
	h.Fn("use", "context", "mycontext").AssertSuccess()
	h.Fn("update", "context", "api-url", "alskjalskdjaff/v2").AssertFailed()
	h.Fn("update", "context", "api-url", "http://alskjalskdjaf").AssertSuccess()
	h.Fn("delete", "context", "mycontext").AssertFailed()
	h.Fn("unset", "context", "mycontext").AssertSuccess()
	h.Fn("delete", "context", "mycontext").AssertSuccess()
}
