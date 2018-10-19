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
	h.Fn("create", "context", "--api-url", "alskjalskdjasd/v2", "mycontext").AssertFailed().AssertStderrContains("Invalid Fn API URL: does not contain ://")
	h.Fn("create", "context", "--api-url", "http://alskjalskdjasd", "mycontext").AssertSuccess()
	h.Fn("use", "context", "mycontext").AssertSuccess()
	h.Fn("update", "context", "api-url", "alskjalskdjaff/v2").AssertFailed()
	h.Fn("update", "context", "api-url").AssertFailed().AssertStderrContains("Please specify a value")
	h.Fn("update", "context", "api-url", "http://alskjalskdjaf").AssertSuccess()
	h.Fn("update", "context", "--delete").AssertFailed().AssertStderrContains("Update context files using fn update context requires the missing argument '<key>'")
	h.Fn("update", "context", "api-u", "--delete").AssertFailed().AssertStderrContains("Context file does not contain key: api-u")
	h.Fn("update", "context", "api-url", "--delete").AssertSuccess().AssertStdoutContains("Current context deleted api-url")
	h.Fn("delete", "context", "mycontext").AssertFailed().AssertStderrContains("Cannot delete the current context: mycontext")
	h.Fn("unset", "context", "mycontext").AssertSuccess()
	h.Fn("delete", "context", "mycontext").AssertSuccess()
}
