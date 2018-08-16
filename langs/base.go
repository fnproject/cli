package langs

import (
	"errors"
	"fmt"
	"os"
)

// not a map because some helpers can handle multiple keys
var helpers = []LangHelper{}

func init() {
	registerHelper(&GoLangHelper{})
	registerHelper(&JavaLangHelper{version: "1.8"})
	registerHelper(&JavaLangHelper{version: "9"})
	registerHelper(&NodeLangHelper{})
	registerHelper(&PythonLangHelper{Version: "3.6"})
	registerHelper(&PythonLangHelper{Version: "3.7"})
	registerHelper(&RubyLangHelper{})
	registerHelper(&RustLangHelper{})
	registerHelper(&KotlinLangHelper{})
	registerHelper(&ClojureLangHelper{version: "1.9.0"})

}

func registerHelper(h LangHelper) {
	helpers = append(helpers, h)
}

func Helpers() []LangHelper {
	return helpers
}

var (
	ErrBoilerplateExists = errors.New("Function boilerplate already exists")
)

// GetLangHelper returns a LangHelper for the passed in language
func GetLangHelper(lang string) LangHelper {
	for _, h := range helpers {
		if h.Handles(lang) {
			return h
		}
	}
	return nil
}

// LangHelper is the interface that language helpers must implement.
type LangHelper interface {
	// Handles return whether it can handle the passed in lang string or not
	Handles(string) bool
	// LangStrings returns list of supported language strings user can use for runtime
	LangStrings() []string
	// Extension is the file extension this helper supports. Eg: .java, .go, .js
	Extensions() []string
	// Runtime that will be used for the build (includes version)
	Runtime() string
	// BuildFromImage is the base image to build off, typically fnproject/LANG:dev
	BuildFromImage() (string, error)
	// RunFromImage is the base image to use for deployment (usually smaller than the build images)
	RunFromImage() (string, error)
	// If set to false, it will use a single Docker build step, rather than multi-stage
	IsMultiStage() bool
	// Dockerfile build lines for building dependencies or anything else language specific
	DockerfileBuildCmds() []string
	// DockerfileCopyCmds will run in second/final stage of multi-stage build to copy artifacts form the build stage
	DockerfileCopyCmds() []string
	// Entrypoint sets the Docker Entrypoint. One of Entrypoint or Cmd is required.
	Entrypoint() (string, error)
	// Cmd sets the Docker command. One of Entrypoint or Cmd is required.
	Cmd() (string, error)
	// DefaultFormat provides the default fn format to set in func.yaml fn init, return "" for an empty format.
	DefaultFormat() string
	HasPreBuild() bool
	PreBuild() error
	AfterBuild() error
	// HasBoilerplate indicates whether a language has support for generating function boilerplate.
	HasBoilerplate() bool
	// GenerateBoilerplate generates basic function boilerplate. Returns ErrBoilerplateExists if the function file
	// already exists.
	GenerateBoilerplate(string) error
	// FixImagesOnInit determines if images should be fixed on initialization - BuildFromImage and RunFromImage will be written to func.yaml
	FixImagesOnInit() bool
}

func defaultHandles(h LangHelper, lang string) bool {
	for _, s := range h.LangStrings() {
		if lang == s {
			return true
		}
	}
	return false
}

// BaseHelper is empty implementation of LangHelper for embedding in implementations.
type BaseHelper struct {
}

func (h *BaseHelper) IsMultiStage() bool               { return true }
func (h *BaseHelper) DockerfileBuildCmds() []string    { return []string{} }
func (h *BaseHelper) DockerfileCopyCmds() []string     { return []string{} }
func (h *BaseHelper) Entrypoint() (string, error)      { return "", nil }
func (h *BaseHelper) Cmd() (string, error)             { return "", nil }
func (h *BaseHelper) HasPreBuild() bool                { return false }
func (h *BaseHelper) PreBuild() error                  { return nil }
func (h *BaseHelper) AfterBuild() error                { return nil }
func (h *BaseHelper) HasBoilerplate() bool             { return false }
func (h *BaseHelper) GenerateBoilerplate(string) error { return nil }
func (h *BaseHelper) DefaultFormat() string            { return "" }
func (h *BaseHelper) FixImagesOnInit() bool            { return false }

// exists checks if a file exists
func exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func dockerBuildError(err error) error {
	return fmt.Errorf("error running docker build: %v", err)
}
