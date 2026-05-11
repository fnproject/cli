package test

import (
	"testing"

	"github.com/fnproject/cli/testharness"
)

func TestInitDetachedTimeoutWritesYaml(t *testing.T) {
	t.Parallel()

	h := testharness.Create(t)
	defer h.Cleanup()

	appName := h.NewAppName()
	funcName := h.NewFuncName(appName)
	dirName := funcName + "_dir"
	h.Fn("init", "--runtime", "go", "--name", funcName, "--detached-timeout", "20m", dirName).AssertSuccess()

	h.Cd(dirName)
	yamlFile := h.GetYamlFile("func.yaml")
	if yamlFile.Deploy == nil || yamlFile.Deploy.OCI == nil || yamlFile.Deploy.OCI.DetachedMode == nil {
		t.Fatal("expected detached mode settings in func.yaml")
	}
	if yamlFile.Deploy.OCI.DetachedMode.Timeout != "20m" {
		t.Fatalf("expected detached timeout 20m, got %#v", yamlFile.Deploy.OCI.DetachedMode.Timeout)
	}
}