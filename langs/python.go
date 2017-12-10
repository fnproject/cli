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

func (lh *PythonLangHelper) HasBoilerplate() bool { return true }

func (lh *PythonLangHelper) GenerateBoilerplate() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	codeFile := filepath.Join(wd, "func.py")
	if exists(codeFile) {
		return ErrBoilerplateExists
	}

	if err := ioutil.WriteFile(codeFile, []byte(helloPythonSrcBoilerplate), os.FileMode(0644)); err != nil {
		return err
	}

	return nil
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

const (
	helloPythonSrcBoilerplate = `
	import sys
	import os
	import json
	
	sys.stderr.write("Starting Python Function\n")
	
	name = "I speak Python too"
	
	try:
	  if not os.isatty(sys.stdin.fileno()):
		try:
		  obj = json.loads(sys.stdin.read())
		  if obj["name"] != "":
			name = obj["name"]
		except ValueError:
		  # ignore it
		  sys.stderr.write("no input, but that's ok\n")
	except:
	  pass
	
	print "Hello, " + name + "!"
`
)
