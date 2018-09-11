package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/fnproject/cli/commands"
	"github.com/fnproject/cli/testharness"
)

func TestMigrateFuncYaml(t *testing.T) {
	for _, rt := range Runtimes {
		t.Run(fmt.Sprintf("%s migrating V1 func file with runtime", rt.runtime), func(t *testing.T) {
			h := testharness.Create(t)
			defer h.Cleanup()

			funcName := h.NewFuncName()
			h.MkDir(funcName)
			h.Cd(funcName)

			h.CreateFuncfile(funcName, rt.runtime)
			h.Fn("migrate").AssertSuccess().AssertStdoutContains(commands.MigrateSuccessMessage)

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
		})
	}
}
