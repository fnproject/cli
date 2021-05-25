package langs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

type RubyLangHelper struct {
	BaseHelper
	Version string
}

func (h *RubyLangHelper) Handles(lang string) bool {
	return defaultHandles(h, lang)
}

func (h *RubyLangHelper) Runtime() string {
	return h.LangStrings()[0]
}

func (h *RubyLangHelper) LangStrings() []string {
	return []string{"ruby", fmt.Sprintf("ruby%s", h.Version)}
}
func (h *RubyLangHelper) Extensions() []string {
	return []string{".rb"}
}

// CustomMemory - no memory override here.
func (h *RubyLangHelper) CustomMemory() uint64 {
	return 0
}
func (h *RubyLangHelper) BuildFromImage() (string, error) {
	return fmt.Sprintf("fnproject/ruby:%s-dev", h.Version), nil
}

func (h *RubyLangHelper) RunFromImage() (string, error) {
	// return fmt.Sprintf("fnproject/ruby:%s", h.Version), nil
	return fmt.Sprintf("fnproject/ruby:%s", h.Version), nil
}

func (h *RubyLangHelper) DockerfileBuildCmds() []string {
	r := []string{}
	if exists("Gemfile") {
		r = append(r,
			"ADD Gemfile* /function/",
			"RUN bundle install",
		)
	}
	return r
}

func (h *RubyLangHelper) DockerfileCopyCmds() []string {
	return []string{
		"COPY --from=build-stage /usr/lib/ruby/gems/ /usr/lib/ruby/gems/", // skip this if no Gemfile?  Does it matter?
		"COPY . /function/",
		"RUN chmod -R o+r /function",
	}
}

func (h *RubyLangHelper) Entrypoint() (string, error) {
	return "ruby func.rb", nil
}

func (h *RubyLangHelper) HasBoilerplate() bool { return true }

func (h *RubyLangHelper) GenerateBoilerplate(path string) error {
	fdkVersion, err := h.GetLatestFDKVersion()
	if err != nil {
		return err
	}

	msg := "%s already exists, can't generate boilerplate"
	codeFile := filepath.Join(path, "func.rb")
	if exists(codeFile) {
		return fmt.Errorf(msg, "func.rb")
	}
	gemFile := filepath.Join(path, "Gemfile")
	if exists(gemFile) {
		return fmt.Errorf(msg, "Gemfile")
	}

	if err := ioutil.WriteFile(codeFile, []byte(rubySrcBoilerplate), os.FileMode(0644)); err != nil {
		return err
	}

	gemfileBoilerplate := fmt.Sprintf(rubyGemfileBoilerplate, fdkVersion)
	if err := ioutil.WriteFile(gemFile, []byte(gemfileBoilerplate), os.FileMode(0644)); err != nil {
		return err
	}

	return nil
}

func (h *RubyLangHelper) GetLatestFDKVersion() (string, error) {

	const versionURL = "https://rubygems.org/api/v1/versions/fdk/latest.json"
	const versionEnv = "FN_RUBY_FDK_VERSION"
	fetchError := fmt.Errorf("failed to fetch latest Ruby FDK version from %v. "+
		"Check your network settings or manually override the Ruby FDK version by setting %s", versionURL, versionEnv)

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
		Version string `json:"version"`
	}{}
	err = json.Unmarshal(body, &parsedResp)
	if err != nil {
		return "", fetchError
	}

	return parsedResp.Version, nil
}

func (h *RubyLangHelper) FixImagesOnInit() bool {
	return true
}

const (
	rubySrcBoilerplate = `require 'fdk'

def myfunction(context:, input:)
  input_value = input.respond_to?(:fetch) ? input.fetch('name') : input
  name = input_value.to_s.strip.empty? ? 'World' : input_value
  FDK.log(entry: "Inside Ruby Hello World function")
  { message: "Hello #{name}!" }
end

FDK.handle(target: :myfunction)
`

	rubyGemfileBoilerplate = `source 'https://rubygems.org' do
  gem 'fdk', '>= %s'
end
`
)
