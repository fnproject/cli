package langs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type RubyLangHelper struct {
	BaseHelper
}

func (h *RubyLangHelper) Handles(lang string) bool {
	return defaultHandles(h, lang)
}
func (h *RubyLangHelper) Runtime() string {
	return h.LangStrings()[0]
}

func (h *RubyLangHelper) LangStrings() []string {
	return []string{"ruby"}
}
func (h *RubyLangHelper) Extensions() []string {
	return []string{".rb"}
}

// CustomMemory - no memory override here.
func (h *RubyLangHelper) CustomMemory() uint64 {
	return 0
}
func (h *RubyLangHelper) BuildFromImage() (string, error) {
	return "fnproject/ruby:dev", nil
}

func (h *RubyLangHelper) RunFromImage() (string, error) {
	return "fnproject/ruby", nil
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
		"ADD . /function/",
	}
}

func (h *RubyLangHelper) Entrypoint() (string, error) {
	return "ruby func.rb", nil
}

func (h *RubyLangHelper) HasBoilerplate() bool { return true }

func (h *RubyLangHelper) GenerateBoilerplate(path string) error {
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

	if err := ioutil.WriteFile(gemFile, []byte(rubyGemfileBoilerplate), os.FileMode(0644)); err != nil {
		return err
	}

	return nil
}

const (
	rubySrcBoilerplate = `require 'fdk'

def myfunction(context:, input:)
  input_value = input.respond_to?(:fetch) ? input.fetch('name') : input
  name = input_value.to_s.strip.empty? ? 'World' : input_value
  { message: "Hello #{name}!" }
end

FDK.handle(target: :myfunction)
`

	rubyGemfileBoilerplate = `source 'https://rubygems.org' do
  gem 'fdk', '>= 0.0.18'
end
`
)
