package test

import (
	"testing"

	"github.com/fnproject/cli/testharness"
)

func TestSettingFuncName(t *testing.T) {
	t.Run("`fn init --name` should set the name in func.yaml", func(t *testing.T) {
		t.Parallel()
		h := testharness.Create(t)
		defer h.Cleanup()

		funcName := h.NewFuncName()
		dirName := funcName + "_dir"
		h.Fn("init", "--runtime", "java", "--name", funcName, dirName).AssertSuccess()

		h.Cd(dirName)

		yamlFile := h.GetYamlFile("func.yaml")
		if yamlFile.Name != funcName {
			t.Fatalf("Name was not set to %s in func.yaml", funcName)
		}
	})
}
