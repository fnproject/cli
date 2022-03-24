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

// PythonLangHelper - python-specific init helper
type PythonLangHelper struct {
	BaseHelper
	Version string
}

// CustomMemory - python is a hungry beast so specify a higher base memory here.
func (h *PythonLangHelper) CustomMemory() uint64 {
	return 256
}

// HasBoilerplate - yep, we have boilerplate...
func (h *PythonLangHelper) HasBoilerplate() bool { return true }

// GenerateBoilerplate - ...and here it is.
func (h *PythonLangHelper) GenerateBoilerplate(path string) error {
	codeFile := filepath.Join(path, "func.py")
	if exists(codeFile) {
		return errors.New("func.py already exists, canceling init")
	}
	if err := ioutil.WriteFile(codeFile, []byte(helloPythonSrcBoilerplate), os.FileMode(0644)); err != nil {
		return err
	}
	depFile := "requirements.txt"
	fdkVersionInfo, err := h.GetLatestFDKVersion()
	if err != nil {
		fmt.Println("Unable to get latest FDK version, using default")
	}
	if len(fdkVersionInfo) != 0 {
		fdkVersionInfo = fmt.Sprintf(">=%s", fdkVersionInfo)
	}
	if err := ioutil.WriteFile(depFile, []byte(fmt.Sprintf(reqsPythonSrcBoilerplate, fdkVersionInfo)), os.FileMode(0644)); err != nil {
		return err
	}

	return nil
}

func (h *PythonLangHelper) Handles(lang string) bool {
	return defaultHandles(h, lang)
}
func (h *PythonLangHelper) Runtime() string {
	return h.LangStrings()[0]
}

func (h *PythonLangHelper) LangStrings() []string {
	return []string{"python", fmt.Sprintf("python%s", h.Version)}
}

func (h *PythonLangHelper) Extensions() []string {
	return []string{".py"}
}

func (h *PythonLangHelper) BuildFromImage() (string, error) {
	return fmt.Sprintf("fnproject/python:%s-dev", h.Version), nil
}

func (h *PythonLangHelper) RunFromImage() (string, error) {
	return fmt.Sprintf("fnproject/python:%s", h.Version), nil
}

func (h *PythonLangHelper) Entrypoint() (string, error) {
	return "/python/bin/fdk /function/func.py handler", nil
}

func (h *PythonLangHelper) DockerfileBuildCmds() []string {
	var r []string
	if exists("requirements.txt") {
		pip_cmd := `RUN pip3 install --target /python/  --no-cache --no-cache-dir`
		if exists(".pip_cache") {
			r = append(r, "ADD .pip_cache /function/.pip_cache")
			pip_cmd += " --no-index --find-links /function/.pip_cache"
		}
		r = append(r, "ADD requirements.txt /function/")
		r = append(r, fmt.Sprintf(`
			%v -r requirements.txt &&\
			    rm -fr ~/.cache/pip /tmp* requirements.txt func.yaml Dockerfile .venv &&\
			    chmod -R o+r /python`, pip_cmd))
	}
	r = append(r, "ADD . /function/")
	if exists("setup.py") {
		r = append(r, "python setup.py install")
	}
	r = append(r, "RUN rm -fr /function/.pip_cache")

	return r
}

func (h *PythonLangHelper) IsMultiStage() bool {
	return true
}

const (
	helloPythonSrcBoilerplate = `import io
import json
import logging

from fdk import response


def handler(ctx, data: io.BytesIO = None):
    name = "World"
    try:
        body = json.loads(data.getvalue())
        name = body.get("name")
    except (Exception, ValueError) as ex:
        logging.getLogger().info('error parsing json payload: ' + str(ex))

    logging.getLogger().info("Inside Python Hello World function")
    return response.Response(
        ctx, response_data=json.dumps(
            {"message": "Hello {0}".format(name)}),
        headers={"Content-Type": "application/json"}
    )
`
	reqsPythonSrcBoilerplate = `fdk%s`
)

func (h *PythonLangHelper) DockerfileCopyCmds() []string {
	return []string{
		"COPY --from=build-stage /python /python",
		"COPY --from=build-stage /function /function",
		"RUN chmod -R o+r /function",
		"ENV PYTHONPATH=/function:/python",
	}
}

func (h *PythonLangHelper) FixImagesOnInit() bool {
	return true
}

type pypiResponseStruct struct {
	Info struct {
		Version string `json:"version"`
	} `json:"info"`
}

func (h *PythonLangHelper) GetLatestFDKVersion() (string, error) {
	resp, err := http.Get("https://pypi.org/pypi/fdk/json")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	responseBody := &pypiResponseStruct{}
	err = json.NewDecoder(resp.Body).Decode(responseBody)
	return responseBody.Info.Version, err
}
