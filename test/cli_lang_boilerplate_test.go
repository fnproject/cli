package test

import (
	"testing"
	"fmt"
	"github.com/fnproject/cli/test/cliharness"
	"strings"
)

var runtimes = []struct {
	runtime        string
	generatesTests bool
	callInput      string
}{
	{"go", true, ""},
	{"java", false, ""},
	{"java8", false, ""},
	{"java9", false, ""},
	{"kotlin", false, `{"name": "John"}`}, //  no arg fn run is broken https://github.com/fnproject/cli/issues/262
	{"node", false, ""},
	{"ruby", true, ""},
	{"rust", false, ""},
	//	{"python", true, `{"name": "John"}\n`}, //  fn run goes into infinite loop  , https://github.com/fnproject/fdk-python/issues/36
}

func TestFnInitWithBoilerplateBuildsRuns(t *testing.T) {
	for _, runtimeI := range runtimes {
		rt := runtimeI
		t.Run(fmt.Sprintf("%s runtime", rt.runtime), func(t *testing.T) {
			t.Parallel()
			h := cliharness.Create(t)
			defer h.Cleanup()

			funcName := h.NewFuncName()

			h.Fn("init", "--runtime", rt.runtime, funcName).AssertSuccess()

			h.Cd(funcName)
			h.Fn("build").AssertSuccess()

			h.FnWithInput(rt.callInput, "run").AssertSuccess()

			if rt.generatesTests {
				h.Fn("test").AssertSuccess()
			}

			appName := h.NewAppName()
			h.Fn("deploy", "--local", "--app", appName).AssertSuccess()

			h.FnWithInput(rt.callInput, "call", appName, funcName)
		})
	}

}

// This should move above but fn run does not work with python
func TestPythonCall(t *testing.T) {
	h := cliharness.Create(t)
	defer h.Cleanup()

	funcName := h.NewFuncName()

	h.MkDir(funcName)
	h.Cd(funcName)
	h.Fn("init", "--name", funcName, "--runtime", "python3.6").AssertSuccess()
	appName := h.NewAppName()
	h.Fn("deploy", "--local", "--app", appName).AssertSuccess()
	h.Fn("call", appName, funcName).AssertSuccess()
	h.FnWithInput(`{"name": "John"}`, "call", appName, funcName).AssertSuccess()

}

func TestAppYamlDeploy(t *testing.T) {
	h := cliharness.Create(t)
	defer h.Cleanup()

	appName := h.NewAppName()
	fnName := h.NewFuncName()
	h.WithFile("app.yaml", fmt.Sprintf(`name: %s`, appName))
	h.MkDir(fnName)
	h.Cd(fnName)
	h.Fn("init", "--runtime", "go").AssertSuccess()
	h.Cd("")
	h.Fn("deploy", "--all", "--local").AssertSuccess()
	h.Fn("call", appName, fnName).AssertSuccess()
	h.Fn("deploy", "--all", "--local").AssertSuccess()
	h.Fn("call", appName, fnName).AssertSuccess()

}


func TestBump(t *testing.T) {
	h := cliharness.Create(t)
	defer h.Cleanup()

	expectFuncYamlVersion := func(v string) {
		funcYaml := h.GetFile("func.yaml")
		if !strings.Contains(funcYaml, fmt.Sprintf("version: %s", v)) {
			t.Fatalf("exepected version to be %s but got %s", v, funcYaml)
		}

	}

	appName := h.NewAppName()
	fnName := h.NewFuncName()
	h.MkDir(fnName)
	h.Cd(fnName)
	h.Fn("init", "--runtime", "go").AssertSuccess()
	expectFuncYamlVersion("0.0.1")

	h.Fn("bump").AssertSuccess()
	expectFuncYamlVersion("0.0.2")

	h.Fn("bump", "--major").AssertSuccess()
	expectFuncYamlVersion("1.0.0")

	h.Fn("bump").AssertSuccess()
	expectFuncYamlVersion("1.0.1")

	h.Fn("bump", "--minor").AssertSuccess()
	expectFuncYamlVersion("1.1.0")

	h.Fn("deploy", "--local", "--app", appName).AssertSuccess()
	expectFuncYamlVersion("1.1.1")

	h.Fn("routes", "i", appName, fnName).AssertSuccess().AssertStdoutContains(fmt.Sprintf(`"image": "%s:1.1.1",`, fnName))

	h.Fn("deploy", "--local", "--no-bump", "--app", appName).AssertSuccess()
	expectFuncYamlVersion("1.1.1")

	h.Fn("routes", "i", appName, fnName).AssertSuccess().AssertStdoutContains(fmt.Sprintf(`"image": "%s:1.1.1",`, fnName))

}
