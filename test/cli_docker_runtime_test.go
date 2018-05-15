package test

import (
	"testing"

	"github.com/fnproject/cli/testharness"
)

const dockerFile = `FROM golang:latest
RUN mkdir /app
ADD . /app
WORKDIR /app
RUN go build -o hello .
CMD ["./hello"]
`
const goFuncDotGo = `package main

import (
	"fmt"
)

func main() {
	fmt.Println("Hello from Fn for func file 'docker' runtime test !")
}`

const funcYaml = `name: fn_test_hello_docker_runtime
version: 0.0.1
runtime: docker
path: /fn_test_hello_docker_runtime`

func TestDockerRuntimeInit(t *testing.T) {
	t.Parallel()
	tctx := testharness.Create(t)
	defer tctx.Cleanup()
	fnName := tctx.NewFuncName()
	tctx.MkDir(fnName)
	tctx.Cd(fnName)

	tctx.WithFile("Dockerfile", dockerFile, 0644)
	tctx.WithFile("func.go", goFuncDotGo, 0644)

	tctx.Fn("init").AssertSuccess()
	tctx.Fn("--verbose", "build").AssertSuccess()
	tctx.Fn("run").AssertSuccess()

}

func TestDockerRuntimeBuildFailsWithNoDockerfile(t *testing.T) {
	t.Parallel()
	tctx := testharness.Create(t)
	defer tctx.Cleanup()

	fnName := tctx.NewFuncName()
	tctx.MkDir(fnName)
	tctx.Cd(fnName)

	tctx.WithFile("func.yaml", funcYaml, 0644)
	tctx.WithFile("func.go", goFuncDotGo, 0644)

	tctx.Fn("--verbose", "build").AssertFailed().AssertStderrContains("Dockerfile does not exist")

}
