package test

import (
	"testing"
	"github.com/fnproject/cli/testharness"
	"log"
	"strings"
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
	tctx.WithFile("Dockerfile", dockerFile,0644)
	tctx.WithFile("func.go", goFuncDotGo,0644)

	tctx.Fn("init").AssertSuccess()
	tctx.Fn("build").AssertSuccess()
	tctx.Fn("run").AssertSuccess()

}

func TestDockerRuntimeBuildFailsWithNoDockerfile(t *testing.T) {
	t.Parallel()
	tctx := testharness.Create(t)
	defer tctx.Cleanup()

	tctx.WithFile("func.yaml", funcYaml,0644)
	tctx.WithFile("func.go", goFuncDotGo,0644)

	res := tctx.Fn("build")

	if res.Success {
		log.Fatalf("Build should have failed")
	}
	if !strings.Contains(res.Stderr, "Dockerfile does not exist") {
		log.Fatalf("Expected error message not found in result: %v", res)
	}
}
