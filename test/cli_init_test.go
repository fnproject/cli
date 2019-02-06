package test

import (
	"fmt"
	"testing"

	"github.com/fnproject/cli/testharness"
)

func TestSettingFuncName(t *testing.T) {
	t.Run("`fn init --name` should set the name in func.yaml", func(t *testing.T) {
		t.Parallel()
		h := testharness.Create(t)
		defer h.Cleanup()

		appName := h.NewAppName()
		funcName := h.NewFuncName(appName)
		dirName := funcName + "_dir"
		h.Fn("init", "--runtime", "go", "--name", funcName, dirName).AssertSuccess()

		h.Cd(dirName)

		yamlFile := h.GetYamlFile("func.yaml")
		if yamlFile.Name != funcName {
			t.Fatalf("Name was not set to %s in func.yaml", funcName)
		}
	})
}

func TestInitImage(t *testing.T) {

	// NB this test creates a function with `rn init --runtime` then creates an init-image from that
	// This will not be necessary when there are init-images publicly available to pull during this test

	t.Run("`fn init --init-image=<...>` should produce a working function template", func(t *testing.T) {
		h := testharness.Create(t)
		var err error

		// Create the init-image
		appName := h.NewAppName()
		h.Fn("create", "app", appName).AssertSuccess()
		origFuncName := h.NewFuncName(appName)
		h.Fn("init", "--runtime", "go", origFuncName)
		h.Cd(origFuncName)

		origYaml := h.GetYamlFile("func.yaml")
		origYaml.Name = ""
		origYaml.Version = ""
		h.WriteYamlFile("func.init.yaml", origYaml)

		err = h.Exec("tar", "-cf", "go.tar", "func.go", "func.init.yaml", "Gopkg.toml")
		if err != nil {
			fmt.Println(err)
			t.Fatal("Failed to create tarball for init-image")
		}

		const initDockerFile = `FROM alpine:latest
                                        COPY go.tar /
                                        CMD ["cat", "/go.tar"]
                                        `
		h.WithFile("Dockerfile", initDockerFile, 0600)

		err = h.Exec("docker", "build", "-t", origFuncName+"-init", ".")
		if err != nil {
			fmt.Println(err)
			t.Fatal("Failed to create init-image")
		}

		// Hooray we have an init-image!!
		// Lets use it
		h.Cd("")
		newFuncName := h.NewFuncName(appName)

		h.Fn("init", "--init-image", origFuncName+"-init", newFuncName)
		h.Cd(newFuncName)
		h.Fn("--registry", "test", "deploy", "--local", "--no-bump", "--app", appName).AssertSuccess()
		h.Fn("invoke", appName, newFuncName).AssertSuccess()

		newYaml := h.GetYamlFile("func.yaml")
		if newYaml.Name != newFuncName {
			t.Fatalf("generated function name is %s not %s", newYaml.Name, newFuncName)
		}

		if newYaml.Version != "0.0.1" {
			t.Fatalf("generated function version is %s not 0.0.1", newYaml.Version)
		}

	})
}
