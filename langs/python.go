package langs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type PythonLangHelper struct {
	BaseHelper
	Version string
}

func (h *PythonLangHelper) Handles(lang string) bool {
	return defaultHandles(h, lang)
}
func (h *PythonLangHelper) Runtime() string {
	return h.LangStrings()[0]
}

// TODO: I feel like this whole versioning thing here could be done better. eg: `runtime: python:2` where we have a single lang, but support versions in tags (like docker tags).
func (lh *PythonLangHelper) LangStrings() []string {
	if strings.HasPrefix(lh.Version, "3.6") {
		return []string{"python3.6"}
	}
	return []string{"python", "python2.7"}
}
func (lh *PythonLangHelper) Extensions() []string {
	return []string{".py"}
}

func (lh *PythonLangHelper) BuildFromImage() (string, error) {
	return fmt.Sprintf("fnproject/python:%v", lh.Version), nil
}

func (lh *PythonLangHelper) RunFromImage() (string, error) {
	return fmt.Sprintf("fnproject/python:%v", lh.Version), nil
}

func (lh *PythonLangHelper) Entrypoint() (string, error) {
	python := "python2"
	if strings.HasPrefix(lh.Version, "3.6") {
		python = "python3"
	}
	return fmt.Sprintf("%v func.py", python), nil
}

func (h *PythonLangHelper) DockerfileBuildCmds() []string {
	pip := "pip"
	if strings.HasPrefix(h.Version, "3.6") {
		pip = "pip3"
	}
	r := []string{}
	if exists("requirements.txt") {
		r = append(r,
			"ADD requirements.txt /function/",
			fmt.Sprintf("RUN %v install -r requirements.txt", pip),
		)
	}
	r = append(r, "ADD . /function/")
	return r
}

func (h *PythonLangHelper) IsMultiStage() bool {
	return false
}

// HasBoilerplate return whether the Python runtime has boilerplate that can be generated
func (h *PythonLangHelper) HasBoilerplate() bool { return true }

// GenerateBoilerplate creates the func file and test file for the Python runtime
func (h *PythonLangHelper) GenerateBoilerplate() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	codeFile := filepath.Join(wd, "func.py")
	if exists(codeFile) {
		return ErrBoilerplateExists
	}
	testFile := filepath.Join(wd, "test.json")
	if exists(testFile) {
		return ErrBoilerplateExists
	}

	if err := ioutil.WriteFile(codeFile, []byte(helloPythonSrcBoilerplate), os.FileMode(0644)); err != nil {
		return err
	}

	if err := ioutil.WriteFile(testFile, []byte(pythonTestBoilerPlate), os.FileMode(0644)); err != nil {
		return err
	}

	return nil
}

const (
	helloPythonSrcBoilerplate = `
import sys
import json


def sayHello(name=None):
    salutation = {"message": "Hello World"}
    if name is not None:
        salutation["message"]= "Hello " + name
    return json.dumps(salutation)


try:
    data = json.loads(sys.stdin.read())
    print(sayHello(data["name"]))
except ValueError:
    print(sayHello())

`

	pythonTestBoilerPlate = `{
    "tests": [
        {
            "input": {
                "body": {
                    "name": "Johnny"
                }
            },
            "output": {
                "body": {
                    "message": "Hello Johnny"
                }
            }
        },
        {
            "input": {
                "body": ""
            },
            "output": {
                "body": {
                    "message": "Hello World"
                }
            }
        }
    ]
}
`
)

// The multi-stage build didn't work, pip seems to be required for it to load the modules
// func (h *PythonLangHelper) DockerfileCopyCmds() []string {
// return []string{
// "ADD . /function/",
// "COPY --from=build-stage /root/.cache/pip/ /root/.cache/pip/",
// }
// }
