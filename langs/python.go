package langs

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// PythonLangHelper - python-specific init helper
type PythonLangHelper struct {
	BaseHelper
	Version string
}

// CustomMemory - python is a hungry beast so specify a higher base memory here.
func (h *PythonLangHelper) CustomMemory() uint64 {
	return 256
}

// HasBoilerplate - yep, we have boilerplate...
func (h *PythonLangHelper) HasBoilerplate() bool { return true }

// GenerateBoilerplate - ...and here it is.
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
	return "/python/bin/fdk func.py handler", nil
}

func (h *PythonLangHelper) DockerfileBuildCmds() []string {
	var r []string
	if exists("requirements.txt") {
		r = append(r, "ADD requirements.txt /function/")
		r = append(r, `
RUN pip3 install --target /python/  --no-cache --no-cache-dir -r requirements.txt &&\
    rm -fr ~/.cache/pip /tmp* requirements.txt func.yaml Dockerfile .venv`)

	}
	r = append(r, "ADD . /function/")
	if exists("setup.py") {
		r = append(r, "python setup.py install")
	}
	return r
}

func (h *PythonLangHelper) IsMultiStage() bool {
	return true
}

const (
	helloPythonSrcBoilerplate = `import io
import json

from fdk import response


def handler(ctx, data: io.BytesIO=None):
    name = "World"
    try:
        body = json.loads(data.getvalue())
        name = body.get("name")
    except (Exception, ValueError) as ex:
        print(str(ex))

    return response.Response(
        ctx, response_data=json.dumps(
            {"message": "Hello {0}".format(name)}),
        headers={"Content-Type": "application/json"}
    )
`
	reqsPythonSrcBoilerplate = `fdk`
)

func (h *PythonLangHelper) DockerfileCopyCmds() []string {
	return []string{
		"COPY --from=build-stage /function /function",
		"COPY --from=build-stage /python /python",
		"ENV PYTHONPATH=/python",
		"WORKDIR /function",
	}
}
