package test

import (
	"fmt"
	"github.com/fnproject/cli/langs"
	"github.com/fnproject/cli/testharness"
	"os"
	"strings"
	"testing"
)

const (
	rubySrcBoilerplate = `require 'fdk'

def myfunction(context:, input:)
  # input_value = input.respond_to?(:fetch) ? input.fetch('name') : input
  # name = input_value.to_s.strip.empty? ? 'World' : input_value
  # FDK.log(entry: "Inside Ruby Hello World function")
  "ruby#{RUBY_VERSION}"
end

FDK.handle(target: :myfunction)
`

	rubyGemfileBoilerplate = `source 'https://rubygems.org' do
  gem 'fdk', '>= %s'
end
`
	funcYamlContent = `schema_version: 20180708
name: %s 
version: 0.0.1
runtime: ruby 
entrypoint: ruby func.rb`
)

// Fallback version scenario for older cli clients
func TestFnBuildWithOlderRuntimeWithoutVersion(t *testing.T) {
	t.Run("`fn invoke` should return the fallback ruby version", func(t *testing.T) {
		t.Parallel()
		h := testharness.Create(t)
		defer h.Cleanup()

		appName := h.NewAppName()
		funcName := h.NewFuncName(appName)
		dirName := funcName + "_dir"
		fmt.Println(appName + " " + funcName)
		h.Fn("create", "app", appName).AssertSuccess()
		h.Fn("init", "--runtime", "ruby", "--name", funcName, dirName).AssertSuccess()
		h.Cd(dirName)

		mod := os.FileMode(int(0777))
		h.WithFile("func.rb", rubySrcBoilerplate, mod)

		oldClientYamlFile := fmt.Sprintf(funcYamlContent, funcName)
		h.WithFile("func.yaml", oldClientYamlFile, mod)

		h.Fn("--verbose", "build").AssertSuccess()
		h.Fn("--registry", "test", "deploy", "--local", "--app", appName).AssertSuccess()
		result := h.Fn("invoke", appName, funcName).AssertSuccess()

		// get the returned version from ruby image
		imageVersion := result.Stdout

		fallBackHandler := langs.GetFallbackLangHelper("ruby")
		fallBackVersion := fallBackHandler.LangStrings()[1]
		fmt.Println("Ruby version returned by image :" + imageVersion)
		fmt.Println("Fallback ruby version :" + fallBackVersion)
		match := strings.Contains(imageVersion, fallBackVersion)
		if !match {
			err := fmt.Sprint("Versions do not match, `ruby` image version #{imageVersion} does not match with fallback version #{fallbackVersion}")
			panic(err)
		}
	})
}