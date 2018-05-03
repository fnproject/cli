package test

import (
	"testing"
	"github.com/fnproject/cli/test/cliharness"
	"log"
	"strings"
)

func TestDockerRuntimeInit(t *testing.T) {
	t.Parallel()
	tctx := cliharness.Create(t)
	defer tctx.Cleanup()
	tctx.CopyFiles(map[string]string{
		"testfuncs/docker/Dockerfile": "Dockerfile",
		"testfuncs/docker/func.go":    "func.go",
	})

	tctx.Fn("init").AssertSuccess()
	tctx.Fn("build").AssertSuccess()
	tctx.Fn("run").AssertSuccess()

}

func TestDockerRuntimeBuildFailsWithNoDockerfile(t *testing.T) {
	tctx := cliharness.Create(t)
	defer tctx.Cleanup()

	tctx.CopyFiles(map[string]string{
		"testfuncs/docker/func.yaml": "func.yaml",
		"testfuncs/docker/func.go":   "func.go",
	})

	res := tctx.Fn("build")

	if res.Success {
		log.Fatalf("Build should have failed")
	}
	if !strings.Contains(res.Stderr, "Dockerfile does not exist") {
		log.Fatalf("Expected error message not found in result: %v", res)
	}
}

