package test

import (
	"testing"
	"fmt"
	"github.com/fnproject/cli/testharness"
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
	t.Parallel()

	for _, runtimeI := range runtimes {
		rt := runtimeI
		t.Run(fmt.Sprintf("%s runtime", rt.runtime), func(t *testing.T) {
			t.Parallel()
			h := testharness.Create(t)
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
	t.Parallel()

	h := testharness.Create(t)
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
