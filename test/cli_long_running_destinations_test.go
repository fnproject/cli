package test

import (
	"testing"

	"github.com/fnproject/cli/testharness"
)

func TestInitDetachedDestinationsWriteYaml(t *testing.T) {
	t.Parallel()

	h := testharness.Create(t)
	defer h.Cleanup()

	appName := h.NewAppName()
	funcName := h.NewFuncName(appName)
	dirName := funcName + "_dir"
	h.Fn("init", "--runtime", "go", "--name", funcName,
		"--on-success", "stream:ocid1.stream.oc1..abc",
		"--on-failure", "notifications:ocid1.onstopic.oc1..abc",
		dirName).AssertSuccess()

	h.Cd(dirName)
	yamlFile := h.GetYamlFile("func.yaml")
	if yamlFile.Deploy == nil || yamlFile.Deploy.OCI == nil || yamlFile.Deploy.OCI.DetachedMode == nil {
		t.Fatal("expected detached mode settings in func.yaml")
	}
	if yamlFile.Deploy.OCI.DetachedMode.OnSuccess == nil || yamlFile.Deploy.OCI.DetachedMode.OnSuccess.Type != "stream" {
		t.Fatalf("expected on_success stream destination, got %#v", yamlFile.Deploy.OCI.DetachedMode.OnSuccess)
	}
	if yamlFile.Deploy.OCI.DetachedMode.OnFailure == nil || yamlFile.Deploy.OCI.DetachedMode.OnFailure.Type != "notifications" {
		t.Fatalf("expected on_failure notifications destination, got %#v", yamlFile.Deploy.OCI.DetachedMode.OnFailure)
	}
}