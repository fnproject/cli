package langs

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type GoLangHelper struct {
	BaseHelper
}

func (h *GoLangHelper) Handles(lang string) bool {
	return defaultHandles(h, lang)
}
func (h *GoLangHelper) Runtime() string {
	return h.LangStrings()[0]
}

func (lh *GoLangHelper) LangStrings() []string {
	return []string{"go"}
}
func (lh *GoLangHelper) Extensions() []string {
	return []string{".go"}
}

func (lh *GoLangHelper) BuildFromImage() (string, error) {
	return "fnproject/go:dev", nil
}

func (lh *GoLangHelper) RunFromImage() (string, error) {
	return "fnproject/go", nil
}

func (h *GoLangHelper) DockerfileBuildCmds() []string {
	r := []string{}
	// more info on Go multi-stage builds: https://medium.com/travis-on-docker/multi-stage-docker-builds-for-creating-tiny-go-images-e0e1867efe5a
	r = append(r, "ADD . /go/src/func/")
	vendor := exists("vendor/")
	// skip dep/glide tool install if vendor is there
	if !vendor {
		if exists("Gopkg.toml") && exists("Gopkg.lock") {
			r = append(r, "RUN go get -u github.com/golang/dep/cmd/dep",
				"RUN cd /go/src/func/ && dep ensure -v")
		}
	}

	r = append(r, "RUN cd /go/src/func/ && go build -o func")

	return r
}

func (h *GoLangHelper) DockerfileCopyCmds() []string {
	return []string{
		"COPY --from=build-stage /go/src/func/func /function/",
	}
}

func (lh *GoLangHelper) Entrypoint() (string, error) {
	return "./func", nil
}

func (lh *GoLangHelper) HasBoilerplate() bool { return true }

func (lh *GoLangHelper) GenerateBoilerplate(properties ...string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	codeFile := filepath.Join(wd, "func.go")
	if exists(codeFile) {
		return errors.New("func.go already exists, canceling init.")
	}
	if err := ioutil.WriteFile(codeFile, []byte(helloGoSrcBoilerplate), os.FileMode(0644)); err != nil {
		return err
	}

	testFile := filepath.Join(wd, "test.json")
	if exists(testFile) {
		fmt.Println("test.json already exists, skipping")
	} else {
		if err := ioutil.WriteFile(testFile, []byte(goTestBoilerPlate), os.FileMode(0644)); err != nil {
			return err
		}
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
