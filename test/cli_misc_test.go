package test

import (
	"testing"
	"github.com/fnproject/cli/test/cliharness"
	"strings"
	"fmt"
)

func TestFnVersion(t *testing.T) {
	tctx := cliharness.Create(t)
	res := tctx.Fn("version")
	res.AssertSuccess()
}

func TestFnApiUrlDifferentFormats(t *testing.T) {
	tctx := cliharness.Create(t)
	for _, url := range []string{"http://localhost:8080", "http://localhost:8080/v1", "localhost:8080", "localhost:8080/v1"} {
		tctx.WithEnv("FN_API_URL", url)
		tctx.Fn("apps", "list").AssertSuccess()
	}
}

// Not sure what this test was intending (copied from old test.sh)
func TestSettingMillisWorks(t *testing.T) {
	h := cliharness.Create(t)
	defer h.Cleanup()
	h.WithEnv("FN_REGISTRY", "some_random_registry")

	appName := h.NewAppName()
	h.Fn("apps", "create", appName).AssertSuccess()
	res := h.Fn("apps", "list")

	if !strings.Contains(res.Stdout, fmt.Sprintf("%s\n", appName)) {
		t.Fatalf("Expecting app list to contain app name , got %v", res)
	}

	funcName := h.NewFuncName()

	h.MkDir(funcName)
	h.Cd(funcName)
	h.Fn("init", "--runtime", "go", "--name", funcName).AssertSuccess()
	h.FileAppend("func.yaml", "\ncpus: 50m\n")

	h.Fn("deploy", "--app", appName, "--local").AssertSuccess()
	h.Fn("call", appName, funcName).AssertSuccess()
	inspectRes := h.Fn("routes", "inspect", appName, funcName)
	inspectRes.AssertSuccess()
	if !strings.Contains(inspectRes.Stdout, `"cpus": "50m"`) {
		t.Errorf("Expecting fn inspect to contain CPU %v", inspectRes)
	}

	h.Fn("routes", "create", appName, "/another", "--image", "some_random_registry/"+funcName+":0.0.2").AssertSuccess()

	h.Fn("call", appName, "/another").AssertSuccess()
}

func TestAllMainCommandsExist(t *testing.T) {
	h := cliharness.Create(t)
	defer h.Cleanup()

	testCommands := []string{
		"init",
		"apps",
		"routes",
		"images",
		"lambda",
		"version",
		"build",
		"bump",
		"deploy",
		"run",
		"push",
		"logs",
		"calls",
		"call",
	}

	for _, cmd := range testCommands {
		res := h.Fn(cmd)
		if strings.Contains(res.Stderr, "command not found") {
			t.Errorf("expected command %s to exist", cmd)
		}
	}
}
