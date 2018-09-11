package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/fnproject/cli/commands"
	"github.com/fnproject/cli/common"
	"github.com/fnproject/cli/testharness"
	yaml "gopkg.in/yaml.v2"
)

func TestInit(t *testing.T) {
	h := testharness.Create(t)
	defer h.Cleanup()

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

	funcName := h.NewFuncName()
	h.Fn("init", "--runtime", "go").AssertSuccess()
	fmt.Println("ffName: ", funcName)

	ffname := "func.yaml"
	b, err := ioutil.ReadFile(ffname)
	if err != nil {
		t.Fatalf("Could not open %s for parsing. Error: %v", ffname, err)
	}
	ff := &common.FuncFileV20180708{}
	err = yaml.Unmarshal(b, ff)

	fmt.Println("FF.Runtime: ", ff.Runtime)
	if err != nil {
		t.Fatalf("Could not parse %s. Error: %v", ffname, err)
	}
	// should have version, runtime and entrypoint
	if ff.Version == "" {
		t.Errorf("No version found in generated %s", ffname)
	}
	if ff.Runtime == "" {
		t.Errorf("No runtime found in generated %s", ffname)
	}
	if ff.Entrypoint == "" {
		t.Errorf("No entrypoint found in generated %s", ffname)
	}
}

func funcNameValidation(name string, t *testing.T) {
	err := commands.ValidateFuncName("fooFunc")
	if err == nil {
		t.Error("Expected validation error for function name")
	}
}

func TestFuncNameWithUpperCase(t *testing.T) {
	funcNameValidation("fooMyFunc", t)
}

func TestFuncNameWithColon(t *testing.T) {
	funcNameValidation("foo:myfunc", t)
}
