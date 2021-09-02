/*
 * Copyright (c) 2019, 2020 Oracle and/or its affiliates. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package langs

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

type GoLangHelper struct {
	BaseHelper
	Version string
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
	return []string{"go", fmt.Sprintf("go%s", lh.Version)}
}
func (lh *GoLangHelper) Extensions() []string {
	return []string{".go"}
}

func (lh *GoLangHelper) BuildFromImage() (string, error) {
	return fmt.Sprintf("fnproject/go:%s-dev", lh.Version), nil
}

func (lh *GoLangHelper) RunFromImage() (string, error) {
	return fmt.Sprintf("fnproject/go:%s", lh.Version), nil
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
	fdkVersion, _ := lh.GetLatestFDKVersion()
	if err := ioutil.WriteFile(modFile, []byte(fmt.Sprintf(modBoilerplate, fdkVersion)), os.FileMode(0644)); err != nil {
		return err
	}

	return nil
}

type githubTagResponse struct {
	Name string `json:"name"`
}

func (h *GoLangHelper) GetLatestFDKVersion() (string, error) {
	// Github API has limit on number of calls
	resp, err := http.Get("https://api.github.com/repos/fnproject/fdk-go/tags")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	responseBody := []githubTagResponse{}
	if err = json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return "", err
	}
	if len(responseBody) == 0 {
		return "", errors.New("Could not read latest version of FDK from tags")
	}
	return responseBody[0].Name, nil
}

const (
	helloGoSrcBoilerplate = `package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"

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
	log.Print("Inside Go Hello World function")
	json.NewEncoder(out).Encode(&msg)
}
`

	modBoilerplate = `
module func

require github.com/fnproject/fdk-go %s
`
)

func (h *GoLangHelper) FixImagesOnInit() bool {
	return true
}
