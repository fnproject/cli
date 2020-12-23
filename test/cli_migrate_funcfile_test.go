package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/fnxproject/cli/commands"
	"github.com/fnxproject/cli/testharness"
)

func TestMigrateFuncYaml(t *testing.T) {

	for _, rt := range Runtimes {
		t.Run(fmt.Sprintf("%s migrating V1 func file with runtime", rt.runtime), func(t *testing.T) {
			h := testharness.Create(t)
			defer h.Cleanup()

			appName := h.NewAppName()
			funcName := h.NewFuncName(appName)
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

func TestMigrateFuncYamlV20180708(t *testing.T) {

	for _, rt := range Runtimes {
		t.Run(fmt.Sprintf("%s migrating runtime", rt.runtime), func(t *testing.T) {
			h := testharness.Create(t)
			defer h.Cleanup()

			appName := h.NewAppName()
			funcName := h.NewFuncName(appName)
			h.Fn("init", "--runtime", rt.runtime, funcName).AssertSuccess()
			h.Cd(funcName)
			h.Fn("migrate").AssertFailed().AssertStderrContains(commands.MigrateFailureMessage)
			h.Cd("")
		})
	}
}
