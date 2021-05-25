package test

import (
	"fmt"
	"github.com/fnproject/cli/testharness"
	"os"
	"testing"
)

const (
	helloPythonSrcBoilerplate = `import io
import json
import logging
import platform

from fdk import response


def handler(ctx, data: io.BytesIO = None):
    logging.getLogger().info("Inside Python Hello World function")
	version = platform.sys.version
    return response.Response(
        ctx, response_data=json.dumps(
            {"message": "Version {0}".format(version)}),
        headers={"Content-Type": "application/json"}
    )
`
	reqsPythonSrcBoilerplate = `fdk`
	funcYamlContent = `schema_version: 20180708
name: %s 
version: 0.0.1
runtime: python
entrypoint: /python/bin/fdk /function/func.py handler`
)

/*
	Build a yaml file for runtimes:
	Node
	PYthon
	Go
	Ruby
	Java
	Kotlin
	Without any version in runtime, now build should use fallback runtimes
	To test:
	Verify dockerfile
*/
func TestFnBuildWithOlderRuntimeWithoutVersion(t *testing.T) {
	t.Run("`fn init --name` should set the name in func.yaml", func(t *testing.T) {
		t.Parallel()
		h := testharness.Create(t)
		defer h.Cleanup()

		appName := h.NewAppName()
		funcName := h.NewFuncName(appName)
		dirName := funcName + "_dir"
		h.Fn("create", "app", appName).AssertSuccess()
		h.Fn("init", "--runtime", "python", "--name", funcName, dirName).AssertSuccess()
		h.Cd(dirName)

		mod := os.FileMode(int(0777))
		h.WithFile("func.py", helloPythonSrcBoilerplate, mod)
		content := h.GetFile("func.py")

		oldClientYamlFile := fmt.Sprintf(funcYamlContent, funcName)
		h.WithFile("func.yaml", oldClientYamlFile, mod)
		content = h.GetFile("func.yaml")
		fmt.Println(content)

		h.Fn("--verbose", "build").AssertSuccess()
		h.Fn("--registry", "test", "deploy", "--local", "--app", appName).AssertSuccess()
		result := h.Fn("invoke", appName, funcName).AssertSuccess()
		out := result.Stdout
		fmt.Println(out)
	})
}