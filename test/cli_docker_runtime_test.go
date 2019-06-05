package test

import (
	"github.com/fnproject/cli/testharness"
	"testing"
)

const dockerFile = `FROM golang:latest
FROM fnproject/go:dev as build-stage
WORKDIR /function
ADD . /go/src/func/
RUN cd /go/src/func/ && go build -o func
FROM fnproject/go
WORKDIR /function
COPY --from=build-stage /go/src/func/func /function/
ENTRYPOINT ["./func"]
`

const funcYaml = `name: fn_test_hello_docker_runtime
version: 0.0.1
runtime: docker`

func withGoFunction(h *testharness.CLIHarness) {

	h.CopyFiles(map[string]string{
		"simplefunc/vendor":  "vendor",
		"simplefunc/func.go": "func.go",
		"simplefunc/go.sum":  "go.sum",
		"simplefunc/go.mod":  "go.mod",
	})
}

func TestDockerRuntimeInit(t *testing.T) {
	t.Parallel()
	tctx := testharness.Create(t)
	defer tctx.Cleanup()

	appName := tctx.NewAppName()
	tctx.Fn("create", "app", appName).AssertSuccess()
	fnName := tctx.NewFuncName(appName)
	tctx.MkDir(fnName)
	tctx.Cd(fnName)
	withGoFunction(tctx)
	tctx.WithFile("Dockerfile", dockerFile, 0644)

	tctx.Fn("init").AssertSuccess()
	tctx.Fn("--verbose", "build").AssertSuccess()
	tctx.Fn("--registry", "test", "deploy", "--local", "--app", appName).AssertSuccess()
	tctx.Fn("invoke", appName, fnName).AssertSuccess()
}

func TestDockerRuntimeBuildFailsWithNoDockerfile(t *testing.T) {
	t.Parallel()
	tctx := testharness.Create(t)
	defer tctx.Cleanup()

	appName := tctx.NewAppName()
	fnName := tctx.NewFuncName(appName)
	tctx.MkDir(fnName)
	tctx.Cd(fnName)
	withGoFunction(tctx)
	tctx.WithFile("func.yaml", funcYaml, 0644)

	tctx.Fn("--verbose", "build").AssertFailed().AssertStderrContains("Dockerfile does not exist")

}
