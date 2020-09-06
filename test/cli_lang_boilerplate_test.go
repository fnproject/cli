package test

import (
	"fmt"
	"github.com/fnproject/cli/testharness"
	"strconv"
	"testing"
	"time"
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
	{"python3.7.1", ""},
}

func TestFnInitWithBoilerplateBuildsRuns(t *testing.T) {
	t.Parallel()

	i := 0
	for _, runtimeI := range Runtimes {
		rt := runtimeI
		t.Run(fmt.Sprintf("%s runtime", rt.runtime), func(t *testing.T) {
			t.Parallel()
			h := testharness.Create(t)
			defer h.Cleanup()

			appName := h.NewAppNameWithSuffix(strconv.Itoa(i))
			i++
			withMinimalOCIApplication(h)
			h.Fn("create", "app", appName, "--annotation", h.GetSubnetAnnotation()).AssertSuccess()
			funcName := h.NewFuncName(appName)

			h.Fn("init", "--runtime", rt.runtime, funcName).AssertSuccess()

			h.Cd(funcName)
			h.Fn("build").AssertSuccess()

			timeout := 5 * time.Minute
			// Larger timeouts are required to allow this test to complete in OCI mode
			if h.IsOCITestMode() {
				timeout = 60 * time.Minute
			}

			h.FnWithTimeoutAndInput("", timeout, "--registry", h.GetFnRegistry(), "deploy", "--app", appName).AssertSuccess()
			h.FnWithTimeoutAndInput(rt.callInput, timeout, "invoke", appName, funcName).AssertSuccess()
		})
	}

}
