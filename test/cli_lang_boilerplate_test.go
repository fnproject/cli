package test

import (
	"fmt"
	"testing"

	"github.com/fnproject/cli/testharness"
)

var Runtimes = []struct {
	runtime        string
	generatesTests bool
	callInput      string
}{
	{"go", false, ""},
	{"java", false, ""},
	{"java8", false, ""},
	{"java9", false, ""},
	{"kotlin", false, `{"name": "John"}`}, //  no arg fn run is broken https://github.com/fnproject/cli/issues/262
	{"node", false, ""},
	{"ruby", true, ""},
	{"rust", false, ""},
	{"python", true, `{"name": "John"}`},
}

func TestFnInitWithBoilerplateBuildsRuns(t *testing.T) {
	t.Parallel()

	for _, runtimeI := range Runtimes {
		rt := runtimeI
		t.Run(fmt.Sprintf("%s runtime", rt.runtime), func(t *testing.T) {
			t.Parallel()
			h := testharness.Create(t)
			defer h.Cleanup()

			appName := h.NewAppName()
			funcName := h.NewFuncName(appName)

			h.Fn("init", "--runtime", rt.runtime, funcName).AssertSuccess()

			h.Cd(funcName)
			h.Fn("build").AssertSuccess()

			if rt.generatesTests {
				h.Fn("test").AssertSuccess()
			}

			h.Fn("--registry", "test", "deploy", "--local", "--app", appName).AssertSuccess()

			h.FnWithInput(rt.callInput, "invoke", appName, funcName).AssertSuccess()
		})
	}

}
