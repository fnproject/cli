package langs

import (
	"errors"
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

func (h *GoLangHelper) CustomMemory() uint64 {
	return 0
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
	// TODO: if we keep the go.sum on user's drive, we can put this after the dep commands and then the dep layers will be cached.
	vendor := exists("vendor/")
	// skip dep tool install if vendor is there
	if !vendor && exists("Gopkg.toml") {
		r = append(r, "RUN go get -u github.com/golang/dep/cmd/dep")
		if exists("Gopkg.lock") {
			r = append(r, "ADD Gopkg.* /go/src/func/")
			r = append(r, "RUN cd /go/src/func/ && dep ensure --vendor-only")
			r = append(r, "ADD . /go/src/func/")
		} else {
			r = append(r, "ADD . /go/src/func/")
			r = append(r, "RUN cd /go/src/func/ && dep ensure")
		}
	} else if exists("go.mod") {
		r = append(r, "WORKDIR /go/src/func/")
		r = append(r, "ENV GO111MODULE=on")
		if vendor {
			r = append(r, "ENV GOFLAGS=\"-mod=vendor\"")
		}
		r = append(r, "COPY . .")
	} else {
		r = append(r, "ADD . /go/src/func/")
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

func (lh *GoLangHelper) GenerateBoilerplate(path string) error {
	codeFile := filepath.Join(path, "func.go")
	if exists(codeFile) {
		return errors.New("func.go already exists, canceling init")
	}
	if err := ioutil.WriteFile(codeFile, []byte(helloGoSrcBoilerplate), os.FileMode(0644)); err != nil {
		return err
	}
	modFile := "go.mod"
	if err := ioutil.WriteFile(modFile, []byte(modBoilerplate), os.FileMode(0644)); err != nil {
		return err
	}

	return nil
}

const (
	helloGoSrcBoilerplate = `package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	fdk "github.com/fnproject/fdk-go"
)

func main() {
	fdk.Handle(fdk.HandlerFunc(myHandler))
}

type Person struct {
	Name string ` + "`json:\"name\"`" + `
}

func myHandler(ctx context.Context, in io.Reader, out io.Writer) {
	p := &Person{Name: "World"}
	json.NewDecoder(in).Decode(p)
	msg := struct {
		Msg string ` + "`json:\"message\"`" + `
	}{
		Msg: fmt.Sprintf("Hello %s", p.Name),
	}
	json.NewEncoder(out).Encode(&msg)
}
`

	modBoilerplate = `module func

require github.com/fnproject/fdk-go master
`
)
