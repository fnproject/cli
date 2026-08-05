package test

import (
	"fmt"
	"testing"

	"github.com/fnproject/cli/common"
	"github.com/fnproject/cli/testharness"
)

func TestCodeOnlyPushValidation(t *testing.T) {
	t.Run("code-only push should fail when Object Storage target is not configured in current context", func(t *testing.T) {
		t.Parallel()
		h := testharness.Create(t)
		defer h.Cleanup()

		appName := h.NewAppName()
		funcName := h.NewFuncName(appName)
		dirName := funcName + "_dir"
		h.Fn("init", "--code-only", "--runtime-name", "python311.ol9", "--runtime-config-type", "function-update", "--name", funcName, dirName).AssertSuccess()

		h.Cd(dirName)
		h.Fn("build").AssertSuccess()
		h.Fn("push").AssertFailed().AssertStderrContains("code-only Object Storage target is not configured in the current context")
	})

	t.Run("code-only push should fail when built archive is missing even if context has bucket and namespace", func(t *testing.T) {
		t.Parallel()
		h := testharness.Create(t)
		defer h.Cleanup()

		contextName := h.NewContextName()
		h.Fn("create", "context", "--api-url", "http://localhost:8080", contextName).AssertSuccess()
		h.Fn("use", "context", contextName).AssertSuccess()
		h.Fn("update", "context", "object_storage_bucket_name", "code-only-test-files").AssertSuccess()
		h.Fn("update", "context", "namespace", "oraclefunctionsdevelopm").AssertSuccess()

		h.MkDir("hello")
		h.Cd("hello")
		h.WithFile("func.yaml", fmt.Sprintf(`schema_version: %d
name: hello
version: 0.0.1
code_only: true
runtime_config:
  type: function-update
  runtime_name: python311.ol9
handler: hello_world.handler
`, common.LatestYamlVersion), 0644)

		h.Fn("push").AssertFailed().AssertStderrContains("built archive not found at hello.0.0.1.zip")
	})
}