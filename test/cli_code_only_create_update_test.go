package test

import (
	"testing"

	"github.com/fnproject/cli/testharness"
)

func requireTestServer(t *testing.T, h *testharness.CLIHarness) {
	t.Helper()
	if res := h.Fn("list", "apps"); !res.Success {
		t.Skipf("skipping because test server is not reachable: %s", res.Stderr)
	}
}

func TestCodeOnlyCreateValidation(t *testing.T) {
	t.Run("code-only create should reject image and code-only flags together", func(t *testing.T) {
		t.Parallel()
		h := testharness.Create(t)
		defer h.Cleanup()
		requireTestServer(t, h)

		appName := h.NewAppName()
		h.Fn("create", "app", appName).AssertSuccess()
		h.Fn(
			"create", "function", appName, "mixed-mode", "some/image:1.0.0",
			"--code-only",
			"--source-type", "direct",
			"--source-file", "/tmp/archive.zip",
			"--runtime-config-type", "function-update",
			"--runtime-name", "python311.ol9",
			"--handler", "hello_world.handler",
		).AssertFailed().AssertStderrContains("Specify either an image or --code-only options, not both")
	})

	t.Run("code-only create should require source-file for direct source", func(t *testing.T) {
		t.Parallel()
		h := testharness.Create(t)
		defer h.Cleanup()
		requireTestServer(t, h)

		appName := h.NewAppName()
		h.Fn("create", "app", appName).AssertSuccess()
		h.Fn(
			"create", "function", appName, "missing-source",
			"--code-only",
			"--source-type", "direct",
			"--runtime-config-type", "function-update",
			"--runtime-name", "python311.ol9",
			"--handler", "hello_world.handler",
		).AssertFailed().AssertStderrContains("--source-file is required when --source-type=direct")
	})

	t.Run("code-only create should require bucket namespace and object name for object-storage source", func(t *testing.T) {
		t.Parallel()
		h := testharness.Create(t)
		defer h.Cleanup()
		requireTestServer(t, h)

		appName := h.NewAppName()
		h.Fn("create", "app", appName).AssertSuccess()
		h.Fn(
			"create", "function", appName, "missing-object-fields",
			"--code-only",
			"--source-type", "object-storage",
			"--runtime-config-type", "function-update",
			"--runtime-name", "python311.ol9",
			"--handler", "hello_world.handler",
		).AssertFailed().AssertStderrContains("--bucket-name, --namespace, and --object-name are required when --source-type=object-storage")
	})

	t.Run("code-only create should require runtime-version-id in manual mode", func(t *testing.T) {
		t.Parallel()
		h := testharness.Create(t)
		defer h.Cleanup()
		requireTestServer(t, h)

		appName := h.NewAppName()
		h.Fn("create", "app", appName).AssertSuccess()
		h.Fn(
			"create", "function", appName, "missing-version",
			"--code-only",
			"--source-type", "direct",
			"--source-file", "/tmp/archive.zip",
			"--runtime-config-type", "manual",
			"--runtime-name", "python311.ol9",
			"--handler", "hello_world.handler",
		).AssertFailed().AssertStderrContains("--runtime-version-id is required when --runtime-config-type=manual")
	})

	t.Run("code-only create should reject runtime-version-id in function-update mode", func(t *testing.T) {
		t.Parallel()
		h := testharness.Create(t)
		defer h.Cleanup()
		requireTestServer(t, h)

		appName := h.NewAppName()
		h.Fn("create", "app", appName).AssertSuccess()
		h.Fn(
			"create", "function", appName, "bad-function-update",
			"--code-only",
			"--source-type", "direct",
			"--source-file", "/tmp/archive.zip",
			"--runtime-config-type", "function-update",
			"--runtime-name", "python311.ol9",
			"--runtime-version-id", "ocid1.functionsruntimeversion.oc1..example",
			"--handler", "hello_world.handler",
		).AssertFailed().AssertStderrContains("--runtime-version-id is only valid for manual runtime configuration")
	})

	t.Run("code-only create should reject invalid python handler format", func(t *testing.T) {
		t.Parallel()
		h := testharness.Create(t)
		defer h.Cleanup()
		requireTestServer(t, h)

		appName := h.NewAppName()
		h.Fn("create", "app", appName).AssertSuccess()
		h.Fn(
			"create", "function", appName, "bad-handler",
			"--code-only",
			"--source-type", "direct",
			"--source-file", "/tmp/archive.zip",
			"--runtime-config-type", "function-update",
			"--runtime-name", "python311.ol9",
			"--handler", "hello_world:handler",
		).AssertFailed().AssertStderrContains("handler for runtime python311.ol9 must be in the format <fileName>.<function>")
	})

	t.Run("code-only create should reject invalid java handler format", func(t *testing.T) {
		t.Parallel()
		h := testharness.Create(t)
		defer h.Cleanup()
		requireTestServer(t, h)

		appName := h.NewAppName()
		h.Fn("create", "app", appName).AssertSuccess()
		h.WithEnv("PATH", "/usr/bin:/bin")
		h.Fn(
			"create", "function", appName, "bad-java-handler",
			"--code-only",
			"--source-type", "direct",
			"--source-file", "/tmp/archive.zip",
			"--runtime-config-type", "function-update",
			"--runtime-name", "java21.ol10",
			"--handler", "hello.handler",
		).AssertFailed().AssertStderrContains("handler for runtime java21.ol10 must be in the format <class>::<method>")
	})
}

func TestCodeOnlyUpdateValidation(t *testing.T) {
	t.Run("code-only update should reject image and code-only flags together", func(t *testing.T) {
		t.Parallel()
		h := testharness.Create(t)
		defer h.Cleanup()
		requireTestServer(t, h)

		appName := h.NewAppName()
		funcName := h.NewFuncName(appName)
		h.Fn("create", "app", appName).AssertSuccess()
		h.Fn("create", "function", appName, funcName, "foo/someimage:0.0.1").AssertSuccess()

		h.Fn(
			"update", "function", appName, funcName, "some/image:1.0.0",
			"--code-only",
			"--source-type", "direct",
			"--source-file", "/tmp/archive.zip",
			"--runtime-config-type", "function-update",
			"--runtime-name", "python311.ol9",
			"--handler", "hello_world.handler",
		).AssertFailed().AssertStderrContains("Specify either an image update or code-only update flags, not both")
	})

	t.Run("code-only update should fail when no code-only update fields are provided", func(t *testing.T) {
		t.Parallel()
		h := testharness.Create(t)
		defer h.Cleanup()
		requireTestServer(t, h)

		appName := h.NewAppName()
		funcName := h.NewFuncName(appName)
		h.Fn("create", "app", appName).AssertSuccess()
		h.Fn("create", "function", appName, funcName, "foo/someimage:0.0.1").AssertSuccess()

		h.Fn("update", "function", appName, funcName, "--code-only").AssertFailed().AssertStderrContains("no code-only update fields were provided")
	})
}