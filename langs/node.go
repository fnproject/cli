package langs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

type NodeLangHelper struct {
	BaseHelper
	Version string
}

func (h *NodeLangHelper) Handles(lang string) bool {
	return defaultHandles(h, lang)
}
func (h *NodeLangHelper) Runtime() string {
	return h.LangStrings()[0]
}

func (lh *NodeLangHelper) LangStrings() []string {
	return []string{"node", fmt.Sprintf("node%s", lh.Version)}
}
func (lh *NodeLangHelper) Extensions() []string {
	// this won't be chosen by default
	return []string{".js"}
}
func (lh *NodeLangHelper) BuildFromImage() (string, error) {
	return fmt.Sprintf("roneet101/nodey:%s-dev", lh.Version), nil
}
func (lh *NodeLangHelper) RunFromImage() (string, error) {
	return fmt.Sprintf("roneet101/nodey:%s", lh.Version), nil
}

const funcJsContent = `const fdk=require('@fnproject/fdk');

fdk.handle(function(input){
  let name = 'World';
  if (input.name) {
    name = input.name;
  }
  console.log('\nInside Node Hello World function')
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
		"@fnproject/fdk": ">=%s"
	}
}
`

func (h *NodeLangHelper) GenerateBoilerplate(path string) error {
	fdkVersion, err := h.GetLatestFDKVersion()
	if err != nil {
		return err
	}

	pathToPackageJsonFile := filepath.Join(path, "package.json")
	pathToFuncJs := filepath.Join(path, "func.js")

	if exists(pathToPackageJsonFile) || exists(pathToFuncJs) {
		return ErrBoilerplateExists
	}

	packageJsonBoilerplate := fmt.Sprintf(packageJsonContent, fdkVersion)
	err = ioutil.WriteFile(pathToPackageJsonFile, []byte(packageJsonBoilerplate), os.FileMode(0644))
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
	r = append(r, "RUN chmod -R o+r /function")

	return r
}

func (h *NodeLangHelper) GetLatestFDKVersion() (string, error) {

	const versionURL = "https://registry.npmjs.org/@fnproject/fdk"
	const versionEnv = "FN_NODE_FDK_VERSION"
	fetchError := fmt.Errorf("failed to fetch latest Node FDK version from %v. "+
		"Check your network settings or manually override the Node FDK version by setting %s", versionURL, versionEnv)

	version := os.Getenv(versionEnv)
	if version != "" {
		return version, nil
	}

	resp, err := http.Get(versionURL)
	if err != nil || resp.StatusCode != 200 {
		return "", fetchError
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fetchError
	}

	parsedResp := struct {
		DistTags struct {
			Latest string `json:"latest"`
		} `json:"dist-tags"`
	}{}
	err = json.Unmarshal(body, &parsedResp)
	if err != nil {
		return "", fetchError
	}

	return parsedResp.DistTags.Latest, nil
}

func (h *NodeLangHelper) FixImagesOnInit() bool {
	return true
}
