package langs

import (
	"fmt"
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
	return []string{"python3.6"}
}

func (lh *PythonLangHelper) Extensions() []string {
	return []string{".py"}
}

func (lh *PythonLangHelper) BuildFromImage() (string, error) {
	return fmt.Sprintf("python:%v", lh.Version), nil
}

func (lh *PythonLangHelper) RunFromImage() (string, error) {
	return fmt.Sprintf("fnproject/python:%v", lh.Version), nil
}

func (lh *PythonLangHelper) Entrypoint() (string, error) {
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

// The multi-stage build didn't work, pip seems to be required for it to load the modules
// func (h *PythonLangHelper) DockerfileCopyCmds() []string {
// return []string{
// "ADD . /function/",
// "COPY --from=build-stage /root/.cache/pip/ /root/.cache/pip/",
// }
// }
