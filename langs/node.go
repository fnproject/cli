package langs

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

type NodeLangHelper struct {
	BaseHelper
}

func (h *NodeLangHelper) Handles(lang string) bool {
	return defaultHandles(h, lang)
}
func (h *NodeLangHelper) Runtime() string {
	return h.LangStrings()[0]
}

func (lh *NodeLangHelper) LangStrings() []string {
	return []string{"node"}
}
func (lh *NodeLangHelper) Extensions() []string {
	// this won't be chosen by default
	return []string{".js"}
}

func (lh *NodeLangHelper) BuildFromImage() (string, error) {
	return "fnproject/node:dev", nil
}
func (lh *NodeLangHelper) RunFromImage() (string, error) {
	return "fnproject/node", nil
}

const funcJsContent = `const fdk=require('@fnproject/fdk');

fdk.handle(function(input){
  let name = 'World';
  if (input.name) {
    name = input.name;
  }
  return {'message': 'Hello ' + name}
})
`

const packageJsonContent = `{
	"name": "hellofn",
    "version": "1.0.0",
	"description": "example function",
	"main": "func.js",
	"author": "",
	"license": "Apache-2.0",
	"dependencies": {
		"@fnproject/fdk": ">=0.0.13"
	}
}
`

func (h *NodeLangHelper) GenerateBoilerplate(path string) error {
	pathToPackageJsonFile := filepath.Join(path, "package.json")
	pathToFuncJs := filepath.Join(path, "func.js")

	if exists(pathToPackageJsonFile) || exists(pathToFuncJs) {
		return ErrBoilerplateExists
	}

	err := ioutil.WriteFile(pathToPackageJsonFile, []byte(packageJsonContent), os.FileMode(0644))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(pathToFuncJs, []byte(funcJsContent), os.FileMode(0644))
	if err != nil {
		return err
	}

	return nil
}

func (h *NodeLangHelper) HasBoilerplate() bool { return true }

// CustomMemory - no memory override here.
func (h *NodeLangHelper) CustomMemory() uint64 { return 0 }

func (h *NodeLangHelper) Entrypoint() (string, error) {
	return "node func.js", nil
}

func (h *NodeLangHelper) DockerfileBuildCmds() []string {
	r := []string{}
	// skip npm -install if node_modules is local - allows local development
	if exists("package.json") && !exists("node_modules") {
		if exists("package-lock.json") {
			r = append(r, "ADD package-lock.json /function/")
		}

		r = append(r,
			"ADD package.json /function/",
			"RUN npm install",
		)
	}
	return r
}

func (h *NodeLangHelper) DockerfileCopyCmds() []string {
	// excessive but content could be anything really
	r := []string{"ADD . /function/"}
	if exists("package.json") && !exists("node_modules") {
		r = append(r, "COPY --from=build-stage /function/node_modules/ /function/node_modules/")
	}

	return r
}
