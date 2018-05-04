package test

import (
	"testing"

	"github.com/fnproject/cli/testharness"
)

func TestContextCrud(t *testing.T) {
	t.Parallel()
	h := testharness.Create(t)
	defer h.Cleanup()

	h.Fn("context", "list").AssertSuccess().AssertStdoutContains("default")
	h.Fn("context", "create", "--api-url", "alskjalskdjasd/v1", "mycontext").AssertFailed()
	h.Fn("context", "create", "--api-url", "http://alskjalskdjasd/v1", "mycontext").AssertSuccess()
	h.Fn("context", "use", "mycontext").AssertSuccess()
	h.Fn("context", "update", "api-url", "alskjalskdjaff/v1").AssertFailed()
	h.Fn("context", "update", "api-url", "http://alskjalskdjaff/v1").AssertSuccess()
	h.Fn("context", "delete", "mycontext").AssertFailed()
	h.Fn("context", "unset", "mycontext").AssertSuccess()
	h.Fn("context", "delete", "mycontext").AssertSuccess()
}
