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

func (lh *RubyLangHelper) BuildFromImage() string {
	return "fnproject/ruby:dev"
}

func (lh *RubyLangHelper) RunFromImage() string {
	return "fnproject/ruby"
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

func (lh *RubyLangHelper) Entrypoint() string {
	return "ruby func.rb"
}

func (lh *RubyLangHelper) HasBoilerplate() bool { return true }

func (lh *RubyLangHelper) GenerateBoilerplate() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	msg := "%s already exists, can't generate boilerplate"
	codeFile := filepath.Join(wd, "func.rb")
	if exists(codeFile) {
		return fmt.Errorf(msg, "func.rb")
	}
	gemFile := filepath.Join(wd, "Gemfile")
	if exists(gemFile) {
		return fmt.Errorf(msg, "Gemfile")
	}
	testFile := filepath.Join(wd, "test.json")
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
	rubySrcBoilerplate = `require 'json'

# Default value(s)
name = "World"

# Parse input
payload = STDIN.read
if payload != ""
	payload = JSON.parse(payload)
	name = payload['name']
end

# Print response
x = {message:"Hello #{name}"}
# STDERR.puts x.inspect
puts x.to_json

# Logging
STDERR.puts "---> STDERR goes to server logs"
`

	rubyGemfileBoilerplate = `source 'https://rubygems.org'

gem 'json', '> 1.8.2'	
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
