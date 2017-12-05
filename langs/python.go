package langs

import (
	"fmt"
	"strings"
)

type PythonLangHelper struct {
	BaseHelper
	Version string
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

// The multi-stage build didn't work, pip seems to be required for it to load the modules
// func (h *PythonLangHelper) DockerfileCopyCmds() []string {
// return []string{
// "ADD . /function/",
// "COPY --from=build-stage /root/.cache/pip/ /root/.cache/pip/",
// }
// }
