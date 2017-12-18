package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/fnproject/cli/langs"
	"github.com/urfave/cli"
	yaml "gopkg.in/yaml.v2"
)

func TestInit(t *testing.T) {

	testname := "test-init"
	testdir, err := ioutil.TempDir("", testname)
	if err != nil {
		t.Fatalf("ERROR: Failed to make tmp test directory: err: %v", err)
	}
	defer os.RemoveAll(testdir)

	err = os.Chdir(testdir)
	if err != nil {
		t.Fatalf("ERROR: Failed to cd to tmp test directory: err: %v", err)
	}

	helper := &langs.GoLangHelper{}
	helper.GenerateBoilerplate()

	app := newFn()
	err = app.Command("init").Run(cli.NewContext(app, &flag.FlagSet{}, nil))
	if err != nil {
		t.Fatalf("ERROR: Failed run `init` command: err: %v", err)
	}

	ffname := "func.yaml"
	b, err := ioutil.ReadFile(ffname)
	if err != nil {
		t.Fatalf("could not open %s for parsing. Error: %v", ffname, err)
	}
	ff := &funcfile{}
	err = yaml.Unmarshal(b, ff)
	if err != nil {
		t.Fatalf("could not parse %s. Error: %v", ffname, err)
	}
	// should have version, runtime and entrypoint
	if ff.Version == "" {
		t.Errorf("no version found in generated %s", ffname)
	}
	if ff.Runtime == "" {
		t.Errorf("no runtime found in generated %s", ffname)
	}
	if ff.Entrypoint == "" {
		t.Errorf("no entrypoint found in generated %s", ffname)
	}
}
