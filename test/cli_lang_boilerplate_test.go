package test

import (
	"fmt"
	"testing"

	"github.com/fnproject/cli/testharness"
)

var Runtimes = []struct {
	runtime   string
	callInput string
}{
	{"go", ""},
	{"java", ""},
	{"java8", ""},
	{"java11", ""},
	{"kotlin", ""},
	{"node", ""},
	{"ruby", ""},
	{"python", ""},
	{"python3.6", ""},
	{"python3.7", ""},
	{"python3.8", ""},
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
			h.Fn("create", "app", appName).AssertSuccess()
			funcName := h.NewFuncName(appName)

			h.Fn("init", "--runtime", rt.runtime, funcName).AssertSuccess()

			h.Cd(funcName)
			h.Fn("build").AssertSuccess()

			h.Fn("--registry", "test", "deploy", "--local", "--app", appName).AssertSuccess()

			h.FnWithInput(rt.callInput, "invoke", appName, funcName).AssertSuccess()
		})
	}

}
