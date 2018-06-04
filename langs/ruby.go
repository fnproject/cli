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
func (h *RubyLangHelper) DefaultFormat() string {
	return "json"
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
	testFile := filepath.Join(path, "test.json")
	if exists(testFile) {
		return fmt.Errorf(msg, "test.json")
	}

	if err := ioutil.WriteFile(codeFile, []byte(rubySrcBoilerplate), os.FileMode(0644)); err != nil {
		return err
	}

	if err := ioutil.WriteFile(gemFile, []byte(rubyGemfileBoilerplate), os.FileMode(0644)); err != nil {
		return err
	}

	if err := ioutil.WriteFile(testFile, []byte(rubyTestBoilerPlate), os.FileMode(0644)); err != nil {
		return err
	}
	return nil
}

const (
	rubySrcBoilerplate = `require 'fdk'

def myhandler(context, input)
	STDERR.puts "call_id: " + context.call_id
	name = "World"
	if input != nil
		if context.content_type == "application/json"
			nin = input['name']
			if nin && nin != ""
				name = nin
			end
		elsif context.content_type == "text/plain"
			name = input
		else
			raise "Invalid input, expecting JSON!"
		end
	end
	return {message: "Hello " + name.to_s + "!"}
end

FDK.handle(:myhandler)
`

	rubyGemfileBoilerplate = `source 'https://rubygems.org'

gem 'json', '~> 2.0'
gem 'fdk', '>= 0.0.11', '< 2.0.0'
gem 'yajl-ruby', require: 'yajl'
`

	// Could use same test for most langs
	rubyTestBoilerPlate = `{
    "tests": [
        {
            "input": {
                "body": {
                    "name": "Johnny"
                }
            },
            "output": {
                "body": {
                    "message": "Hello Johnny!"
                }
            }
        },
        {
            "input": {
                "body": ""
            },
            "output": {
                "body": {
                    "message": "Hello World!"
                }
            }
        }
    ]
}
`
)
