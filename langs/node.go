package langs

import (
	"os"
	"path/filepath"
	"io/ioutil"
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

const funcJsContent = "const fdk = require('@fnproject/fdk') \n\n" +
	"fdk.handle(function (input, ctx) { \n" +
	"  return 'Hello ' + JSON.stringify(input) \n" +
	"}) \n"

const packageJsonContent = "{\n" +
	"   \"name\": \"hellofn\", \n" +
	"   \"version\": \"1.0.0\", \n" +
	"   \"description\": \"example function\",\n" +
	"   \"main\": \"func.js\",\n" +
	"   \"author\": \"\",\n" +
	"   \"license\": \"Apache-2.0\", \n" +
	"   \"dependencies\": { \n" +
	"      \"@fnproject/fdk\": \"0.x\"\n " +
	"    }\n" +
	"}\n"

func (lh *NodeLangHelper) GenerateBoilerplate() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	pathToPackageJsonFile := filepath.Join(wd, "package.json")
	pathToFuncJs := filepath.Join(wd, "func.js")

	if exists(pathToPackageJsonFile) || exists(pathToFuncJs) {
		return ErrBoilerplateExists
	}

	err = ioutil.WriteFile(pathToPackageJsonFile, []byte(packageJsonContent), os.FileMode(0644))
	if err != nil {
		return err
	}
	return ioutil.WriteFile(pathToFuncJs, []byte(funcJsContent), os.FileMode(0644))
}

func (lh *NodeLangHelper) HasBoilerplate() bool { return true }

func (lh *NodeLangHelper) DefaultFormat() string { return "json" }

func (lh *NodeLangHelper) Entrypoint() (string, error) {
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
	if exists("package.json") &&  !exists("node_modules") {
		r = append(r, "COPY --from=build-stage /function/node_modules/ /function/node_modules/")
	}


	return r
}
