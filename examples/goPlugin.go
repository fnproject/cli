package XXXXXlanghelper //replace with "main"

import "github.com/fnproject/cli/langhelper"

type XXXXXHelper struct {
	langhelper.BaseHelper
}

func main() {
	//you probably don't want to change this
	helper := &XXXXXHelper{}
	helper.Plug = helper
	helper.Main()
}

func (h *XXXXXHelper) Init(flags map[string]string) {
	//change one or both of these values
	cmd := ""
	entrypoint := ""

	//optional: generateBoilerplate() - create boilerplate code for your runtime

	//leave this line as it is
	h.LangInitialiser = &langhelper.LangInitialiser{Cmd: cmd, Entrypoint: entrypoint}
}

func (h *XXXXXHelper) Build(flags map[string]string) {
	//change these as needed
	buildImage := ""
	runImage := ""
	isMultiStage := true
	dockerFileCopyCmds := []string{}
	dockerFileBuildCmds := []string{}

	//optional: preBuild() - run any prebuild stages for your runtime

	//leave this line as it is
	h.LangBuilder = &langhelper.LangBuilder{
		BuildImage:          buildImage,
		RunImage:            runImage,
		IsMultiStage:        isMultiStage,
		DockerfileCopyCmds:  dockerFileCopyCmds,
		DockerfileBuildCmds: dockerFileBuildCmds,
	}
}

func generateBoilerplate() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	codeFile := filepath.Join(wd, "func.go")
	if exists(codeFile) {
		return ErrBoilerplateExists
	}
	testFile := filepath.Join(wd, "test.json")
	if exists(testFile) {
		return ErrBoilerplateExists
	}

	if err := ioutil.WriteFile(codeFile, []byte(helloGoSrcBoilerplate), os.FileMode(0644)); err != nil {
		return err
	}

	if err := ioutil.WriteFile(testFile, []byte(goTestBoilerPlate), os.FileMode(0644)); err != nil {
		return err
	}
	return nil
}

const (
	helloGoSrcBoilerplate = `package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Person struct {
	Name string
}

func main() {
	p := &Person{Name: "World"}
	json.NewDecoder(os.Stdin).Decode(p)
	mapD := map[string]string{"message": fmt.Sprintf("Hello %s", p.Name)}
	mapB, _ := json.Marshal(mapD)
	fmt.Println(string(mapB))
}
`

	// Could use same test for most langs
	goTestBoilerPlate = `{
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
