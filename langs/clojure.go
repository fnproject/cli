package langs

import (
    "bytes"
    "encoding/json"
    "errors"
    "fmt"
    "io/ioutil"
    "net/http"
    "os"
    "path/filepath"
)

// ClojureLangHelper provides a set of helper methods for the lifecycle of Clojure Leiningen projects
type ClojureLangHelper struct {
    BaseHelper
    latestFdkVersion string
    version          string
}

func (h *ClojureLangHelper) Handles(lang string) bool {
    for _, s := range h.LangStrings() {
        if lang == s {
            return true
        }
    }
    return false
}
func (h *ClojureLangHelper) Runtime() string {
    return h.LangStrings()[0]
}

func (lh *ClojureLangHelper) LangStrings() []string {
    return []string{"clojure"}

}
func (lh *ClojureLangHelper) Extensions() []string {
    return []string{".clj"}
}

func (lh *ClojureLangHelper) BuildFromImage() (string, error) {
    return "clojure:lein", nil
}

// RunFromImage returns the Docker image used to run the Clojure function.
func (lh *ClojureLangHelper) RunFromImage() (string, error) {
    return "openjdk:9-jre", nil
}

// HasBoilerplate returns whether the Java runtime has boilerplate that can be generated.
func (lh *ClojureLangHelper) HasBoilerplate() bool { return true }

// Clojure defaults to CloudEvents
func (lh *ClojureLangHelper) DefaultFormat() string { return "cloudevent" }

// GenerateBoilerplate will generate function boilerplate for a Java runtime.
// The default boilerplate is for a Leiningen project.
func (lh *ClojureLangHelper) GenerateBoilerplate(path string) error {
    pathToProjectFile := filepath.Join(path, "project.clj")
    if exists(pathToProjectFile) {
        return ErrBoilerplateExists
    }

    apiVersion, err := lh.getFDKAPIVersion()
    if err != nil {
        return err
    }

    if err := ioutil.WriteFile(pathToProjectFile,
        []byte(clojureProjFileContent(lh.version, apiVersion)),
        os.FileMode(0644)); err != nil {
        return err
    }

    pathToTestJson := filepath.Join(path, "test.json")
    if exists(pathToTestJson) {
        return ErrBoilerplateExists
    }
    if err := ioutil.WriteFile(pathToTestJson,
        []byte(fnTestBoilerplate),
        os.FileMode(0644)); err != nil {
        return err
    }

    mkDirAndWriteFile := func(dir, filename, content string) error {
        fullPath := filepath.Join(path, dir)
        if err = os.MkdirAll(fullPath, os.FileMode(0755)); err != nil {
            return err
        }

        fullFilePath := filepath.Join(fullPath, filename)
        return ioutil.WriteFile(fullFilePath, []byte(content), os.FileMode(0644))
    }

    err = mkDirAndWriteFile("src/func", "core.clj", helloClojureSrcBoilerplate)
    if err != nil {
        return err
    }

    return mkDirAndWriteFile("test/func", "core_test.clj", helloClojureTestBoilerplate)
}

func (lh *ClojureLangHelper) Entrypoint() (string, error) {
    return "java -jar /function/app/func.jar", nil
}

// DockerfileCopyCmds returns the Docker COPY command to copy the compiled Clojure standalone jar.
func (lh *ClojureLangHelper) DockerfileCopyCmds() []string {
    return []string{
        `COPY --from=build-stage /function/target/com.fdk.func-1.0.0-standalone.jar /function/app/func.jar`,
    }
}

// DockerfileBuildCmds returns the build stage steps to compile the Leiningen function project.
func (lh *ClojureLangHelper) DockerfileBuildCmds() []string {
    return []string{
        `ADD project.clj /function/project.clj`,
        `ADD src /function/src`,
        `RUN ["lein", "uberjar"]`,
    }
}

// HasPreBuild returns whether the Leiningen runtime has a pre-build step.
func (lh *ClojureLangHelper) HasPreBuild() bool { return true }

// PreBuild ensures that the expected the function is based is a leiningen project.
func (lh *ClojureLangHelper) PreBuild() error {
    wd, err := os.Getwd()
    if err != nil {
        return err
    }

    if !exists(filepath.Join(wd, "project.clj")) {
        return errors.New("Could not find project.clj - are you sure this is a Leiningen project?")
    }

    return nil
}

func clojureProjFileContent(clojureVersion string, FDKVersion string) string {
    return fmt.Sprintf(clojureProjFile, clojureVersion, FDKVersion)
}

func (lh *ClojureLangHelper) getFDKAPIVersion() (string, error) {

    if lh.latestFdkVersion != "" {
        return lh.latestFdkVersion, nil
    }

    const versionURL = "https://clojars.org/api/artifacts/unpause/fdk-clj"
    const versionEnv = "FN_CLOJURE_FDK_VERSION"
    fetchError := fmt.Errorf("Failed to fetch latest Clojure FDK version from %v. Check your network settings or manually override the version by setting %s", versionURL, versionEnv)

    type parsedResponse struct {
        Version string `json:"latest_version"`
    }
    version := os.Getenv(versionEnv)
    if version != "" {
        return version, nil
    }
    resp, err := http.Get(versionURL)
    if err != nil || resp.StatusCode != 200 {
        return "", fetchError
    }

    buf := bytes.Buffer{}
    _, err = buf.ReadFrom(resp.Body)
    if err != nil {
        return "", fetchError
    }

    var parsedResp parsedResponse
    err = json.Unmarshal(buf.Bytes(), &parsedResp)
    if err != nil {
        fmt.Print(err)
        return "", fetchError
    }

    lh.latestFdkVersion = parsedResp.Version
    return lh.latestFdkVersion, nil
}

func (lh *ClojureLangHelper) FixImagesOnInit() bool {
    return true
}

const (
    clojureProjFile = `(defproject com.fdk.func "1.0.0"
  :description "Clojure FDK for Fn"
  :url "https://github.com/fnproject"
  :license {:name "Apache License Version 2.0"
        :url "http://www.apache.org/licenses/LICENSE-2.0.txt"}
  :dependencies [[org.clojure/clojure "%s"]
                [unpause/fdk-clj "%s"]
                [org.clojure/test.check "0.9.0"]]
  :main func.core
  :jvm-opts ["-Duser.timezone=UTC"]
  :profiles {:uberjar { :aot :all }}
  :test-paths ["test"])
`

    helloClojureSrcBoilerplate = `
(ns func.core 
  (:require [fdk-clj.core :as fdk]) 
  (:gen-class))

(defn handler [ctx data]
  { :message (str "Hello " (get data :name "World")) })

(defn -main [& args] (fdk/handle handler))
`

    helloClojureTestBoilerplate = `
(ns func.core-test
  (:refer-clojure :exclude [extend second])
  (:require [clojure.test :refer :all]
            [func.core :refer :all]))

(deftest handler-test
  (is (= (handler nil { :name "Johnny" }) { :message "Hello Johnny" }))
  (is (= (handler nil nil) {:message "Hello World"})))
`
    fnTestBoilerplate = `{
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
}`
)
