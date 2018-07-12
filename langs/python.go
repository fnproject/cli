package langs

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type PythonLangHelper struct {
	BaseHelper
	Version string
}

func (h *PythonLangHelper) DefaultFormat() string {
	return "json"
}

func (h *PythonLangHelper) HasBoilerplate() bool { return true }

func (h *PythonLangHelper) GenerateBoilerplate(path string) error {
	codeFile := filepath.Join(path, "func.py")
	if exists(codeFile) {
		return errors.New("func.py already exists, canceling init")
	}
	if err := ioutil.WriteFile(codeFile, []byte(helloPythonSrcBoilerplate), os.FileMode(0644)); err != nil {
		return err
	}
	depFile := "requirements.txt"
	if err := ioutil.WriteFile(depFile, []byte(reqsPythonSrcBoilerplate), os.FileMode(0644)); err != nil {
		return err
	}

	testFile := filepath.Join(path, "test.json")
	if exists(testFile) {
		fmt.Println("test.json already exists, skipping")
	} else {
		if err := ioutil.WriteFile(testFile, []byte(pythonTestBoilerPlate), os.FileMode(0644)); err != nil {
			return err
		}
	}
	return nil
}

func (h *PythonLangHelper) Handles(lang string) bool {
	return defaultHandles(h, lang)
}
func (h *PythonLangHelper) Runtime() string {
	return h.LangStrings()[0]
}

func (h *PythonLangHelper) LangStrings() []string {
	return []string{"python", "python3.6"}
}

func (h *PythonLangHelper) Extensions() []string {
	return []string{".py"}
}

func (h *PythonLangHelper) BuildFromImage() (string, error) {
	return fmt.Sprintf("python:%s-slim-stretch", h.Version), nil
}

func (h *PythonLangHelper) RunFromImage() (string, error) {
	return fmt.Sprintf("python:%s-slim-stretch", h.Version), nil
}

func (h *PythonLangHelper) Entrypoint() (string, error) {
	return "python3 func.py", nil
}

func (h *PythonLangHelper) DockerfileBuildCmds() []string {
	pip := "pip3"
	r := []string{}
	r = append(r, "RUN apt-get update && apt-get install --no-install-recommends -qy build-essential gcc")
	if exists("requirements.txt") {
		r = append(r,
			"ADD requirements.txt /function/",
			fmt.Sprintf("RUN %v install --no-cache --no-cache-dir -r requirements.txt", pip),
		)
	}
	r = append(r, "ADD . /function/")
	r = append(r, "RUN rm -fr ~/.cache/pip /tmp* requirements.txt func.yaml Dockerfile .venv")
	return r
}

func (h *PythonLangHelper) IsMultiStage() bool {
	return false
}

const (
	helloPythonSrcBoilerplate = `import fdk
import json


def handler(ctx, data=None, loop=None):
    name = "World"
    if data and len(data) > 0:
        body = json.loads(data)
        name = body.get("name")
    return {"message": "Hello {0}".format(name)}



if __name__ == "__main__":
    fdk.handle(handler)

`
	reqsPythonSrcBoilerplate = `fdk`
	pythonTestBoilerPlate    = `{
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
