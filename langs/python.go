package langs

import (
	"fmt"
	"strings"
)

type PythonLangHelper struct {
	BaseHelper
	Version string
}

func (lh *PythonLangHelper) BuildFromImage() string {
	return fmt.Sprintf("fnproject/python:%v", lh.Version)
}

func (lh *PythonLangHelper) RunFromImage() string {
	return fmt.Sprintf("fnproject/python:%v", lh.Version)
}

func (lh *PythonLangHelper) Entrypoint() string {
	return "python2 func.py"
}

func (h *PythonLangHelper) DockerfileBuildCmds() []string {
	pip := "pip"
	if strings.HasPrefix(h.Version, "2.7") {
		pip = "pip2"
	}
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
