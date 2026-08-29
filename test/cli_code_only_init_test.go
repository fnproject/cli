package test

import (
	"testing"

	"github.com/fnproject/cli/common"
	"github.com/fnproject/cli/testharness"
)

func TestCodeOnlyInit(t *testing.T) {
	t.Run("`fn init --code-only --runtime-name python311.ol9` should generate code-only func.yaml and python boilerplate", func(t *testing.T) {
		t.Parallel()
		h := testharness.Create(t)
		defer h.Cleanup()

		appName := h.NewAppName()
		funcName := h.NewFuncName(appName)
		dirName := funcName + "_dir"
		h.Fn("init", "--code-only", "--runtime-name", "python311.ol9", "--runtime-config-type", "function-update", "--name", funcName, dirName).AssertSuccess()

		h.Cd(dirName)
		yamlFile := h.GetYamlFile("func.yaml")

		if yamlFile.Schema_version != common.LatestYamlVersion {
			t.Fatalf("schema_version was %d, expected %d", yamlFile.Schema_version, common.LatestYamlVersion)
		}
		if !yamlFile.Code_only {
			t.Fatal("code_only was not set in func.yaml")
		}
		if yamlFile.Runtime_config == nil {
			t.Fatal("runtime_config was not set in func.yaml")
		}
		if yamlFile.Runtime_config.Type != "function-update" {
			t.Fatalf("runtime_config.type was %q, expected function-update", yamlFile.Runtime_config.Type)
		}
		if yamlFile.Runtime_config.Runtime_name != "python311.ol9" {
			t.Fatalf("runtime_config.runtime_name was %q, expected python311.ol9", yamlFile.Runtime_config.Runtime_name)
		}
		if yamlFile.Handler != "hello_world.handler" {
			t.Fatalf("handler was %q, expected hello_world.handler", yamlFile.Handler)
		}
		if yamlFile.Build_image != "" || yamlFile.Run_image != "" {
			t.Fatal("code-only func.yaml should not contain build_image or run_image")
		}
		if yamlFile.Runtime != "" {
			t.Fatal("code-only func.yaml should not contain runtime")
		}
		if h.GetFile("hello_world.py") == "" {
			t.Fatal("expected hello_world.py boilerplate to be generated")
		}
	})

	t.Run("`fn init --code-only --runtime go` should generate code-only func.yaml and go boilerplate", func(t *testing.T) {
		t.Parallel()
		h := testharness.Create(t)
		defer h.Cleanup()

		appName := h.NewAppName()
		funcName := h.NewFuncName(appName)
		dirName := funcName + "_dir"
		h.Fn("init", "--code-only", "--runtime", "go", "--runtime-config-type", "function-update", "--name", funcName, dirName).AssertSuccess()

		h.Cd(dirName)
		yamlFile := h.GetYamlFile("func.yaml")

		if !yamlFile.Code_only {
			t.Fatal("code_only was not set in func.yaml")
		}
		if yamlFile.Runtime_config == nil {
			t.Fatal("runtime_config was not set in func.yaml")
		}
		if yamlFile.Runtime_config.Runtime_name != "go" {
			t.Fatalf("runtime_config.runtime_name was %q, expected go", yamlFile.Runtime_config.Runtime_name)
		}
		if yamlFile.Handler != "" {
			t.Fatalf("handler was %q, expected empty for go", yamlFile.Handler)
		}
		if h.GetFile("func.go") == "" {
			t.Fatal("expected func.go boilerplate to be generated")
		}
	})

	t.Run("`fn init --code-only --runtime java` should require Maven", func(t *testing.T) {
		t.Parallel()
		h := testharness.Create(t)
		defer h.Cleanup()

		h.WithEnv("PATH", "/usr/bin:/bin")
		h.Fn("init", "--code-only", "--runtime", "java", "--runtime-config-type", "function-update", "hello-java").AssertFailed().AssertStderrContains("Maven was not found in PATH")
	})
}

func TestRuntimeDiscoveryArgValidation(t *testing.T) {
	t.Run("`fn list runtime-versions` should require runtime-name", func(t *testing.T) {
		t.Parallel()
		h := testharness.Create(t)
		defer h.Cleanup()

		h.Fn("list", "runtime-versions").AssertFailed().AssertStderrContains("--runtime-name is required")
	})

	t.Run("`fn get latest-runtime-version` should require runtime-name", func(t *testing.T) {
		t.Parallel()
		h := testharness.Create(t)
		defer h.Cleanup()

		h.Fn("get", "latest-runtime-version").AssertFailed().AssertStderrContains("--runtime-name is required")
	})
}