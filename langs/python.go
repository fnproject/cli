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

func (h *PythonLangHelper) GenerateBoilerplate() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	codeFile := filepath.Join(wd, "func.py")
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

	testFile := filepath.Join(wd, "test.json")
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
	return []string{"python3.6"}
}

func (h *PythonLangHelper) Extensions() []string {
	return []string{".py"}
}

func (h *PythonLangHelper) BuildFromImage() (string, error) {
	return fmt.Sprintf("python:%v", h.Version), nil
}

func (h *PythonLangHelper) RunFromImage() (string, error) {
	return fmt.Sprintf("fnproject/python:%v", h.Version), nil
}

func (h *PythonLangHelper) Entrypoint() (string, error) {
	return "python3 func.py", nil
}

func (h *PythonLangHelper) DockerfileBuildCmds() []string {
	pip := "pip3"
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

const (
	helloPythonSrcBoilerplate = `
import fdk


@fdk.coerce_input_to_content_type
def handler(context, data=None, loop=None):
    """
    This is just an echo function
    :param context: request context
    :type context: hotfn.http.request.RequestContext
    :param data: request body
    :type data: object
    :param loop: asyncio event loop
    :type loop: asyncio.AbstractEventLoop
    :return: echo of request body
    :rtype: object
    """
    return "Hello {0}".format(data.get("name", "World"))


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
