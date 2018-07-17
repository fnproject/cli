package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/fnproject/cli/testharness"
)

func TestMigrateFuncYaml(t *testing.T) {
	t.Parallel()

	for _, rt := range Runtimes {
		t.Run(fmt.Sprintf("%s migrating runtime", rt.runtime), func(t *testing.T) {
			t.Parallel()
			h := testharness.Create(t)
			defer h.Cleanup()

			funcName := h.NewFuncName()
			h.Fn("init", "--runtime", "java", funcName).AssertSuccess()

			h.Cd(funcName)
			h.Fn("build").AssertSuccess()

			h.FnWithInput("", "run").AssertSuccess()

			h.Fn("migrate").AssertSuccess().AssertStdoutContains("Successfully migrated func.yaml and created a back up func.yaml.bak")

			funcYaml := h.GetFile("func.yaml")
			if !strings.Contains(funcYaml, "schema_version") {
				t.Fatalf("Exepected schema_version in %s", funcYaml)
			}

			yamlFile := h.GetYamlFile("func.yaml")
			if !strings.HasPrefix(yamlFile.Triggers[0].Source, "/") {
				t.Fatalf("Exepected source to have a leading '/' in %s", yamlFile.Triggers[0].Source)
			}
			if yamlFile.Triggers[0].Type != "http" {
				t.Fatalf("Exepected type to be 'http' in %s", yamlFile.Triggers[0].Type)
			}

			h.Fn("build").AssertSuccess()

			h.FnWithInput("", "run").AssertSuccess()
		})
	}
}
