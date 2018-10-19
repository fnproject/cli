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
	return "http-stream"
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

	return nil
}

func (h *PythonLangHelper) Handles(lang string) bool {
	return defaultHandles(h, lang)
}
func (h *PythonLangHelper) Runtime() string {
	return h.LangStrings()[0]
}

func (h *PythonLangHelper) LangStrings() []string {
	return []string{"python", fmt.Sprintf("python%s", h.Version)}
}

func (h *PythonLangHelper) Extensions() []string {
	return []string{".py"}
}

func (h *PythonLangHelper) BuildFromImage() (string, error) {
	return fmt.Sprintf("fnproject/python:%s-dev", h.Version), nil
}

func (h *PythonLangHelper) RunFromImage() (string, error) {
	return fmt.Sprintf("fnproject/python:%s", h.Version), nil
}

func (h *PythonLangHelper) Entrypoint() (string, error) {
	return "python3 func.py", nil
}

func (h *PythonLangHelper) DockerfileBuildCmds() []string {
	r := []string{}
	r = append(r, "ADD . /function/")
	if exists("requirements.txt") {
		r = append(r, `
RUN pip3 install --target /python/  --no-cache --no-cache-dir -r requirements.txt &&\
    rm -fr ~/.cache/pip /tmp* requirements.txt func.yaml Dockerfile .venv`)
	}
	return r
}

func (h *PythonLangHelper) IsMultiStage() bool {
	return true
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
)

func (h *PythonLangHelper) DockerfileCopyCmds() []string {
	return []string{
		"COPY --from=build-stage /function /function",
		"COPY --from=build-stage /python /python",
		"ENV PYTHONPATH=/python",
	}
}
