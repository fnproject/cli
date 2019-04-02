package main

import (
	"flag"
	"io/ioutil"
	"os"
	"testing"

	"github.com/fnproject/cli/commands"
	"github.com/fnproject/cli/common"
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
	helper.GenerateBoilerplate(testdir)

	app := newFn()
	err = app.Command("init").Run(cli.NewContext(app, &flag.FlagSet{}, nil))
	if err != nil {
		t.Fatalf("ERROR: Failed run `init` command: err: %v", err)
	}

	ffname := "func.yaml"
	b, err := ioutil.ReadFile(ffname)
	if err != nil {
		t.Fatalf("Could not open %s for parsing. Error: %v", ffname, err)
	}
	ff := &common.FuncFile{}
	err = yaml.Unmarshal(b, ff)
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

func TestFuncNameValidation(t *testing.T) {
	tc := []struct {
		name        string
		value       string
		errExpected bool
	}{
		{
			name:        "strings only",
			value:       "somename",
			errExpected: false,
		},
		{
			name:        "numbers only",
			value:       "12345",
			errExpected: false,
		},
		{
			name:        "uppercase letters",
			value:       "HelloFuncName",
			errExpected: true,
		},
		{
			name:        "starts with separator",
			value:       ".myfunc",
			errExpected: true,
		},
		{
			name:        "starts with separator (*)",
			value:       "*myfunc",
			errExpected: true,
		},
		{
			name:        "starts with separator (-)",
			value:       "-myfunc",
			errExpected: true,
		},
		{
			name:        "ends with separator",
			value:       "myfunc-",
			errExpected: true,
		},
		{
			name:        "ends with separator (.)",
			value:       "myfunc.",
			errExpected: true,
		},
		{
			name:        "ends with separator (*)",
			value:       "myfunc*",
			errExpected: true,
		},
		{
			name:        "contains spaces",
			value:       "my func blah",
			errExpected: true,
		},
		{
			name:        "hypen in name",
			value:       "my-func-blah",
			errExpected: false,
		},
		{
			name:        "underscore in name",
			value:       "my_funcblah",
			errExpected: false,
		},
		{
			name:        "dot in name",
			value:       "my.funcblah",
			errExpected: false,
		},
		{
			name:        "multiple separators in row",
			value:       "my___funcblah",
			errExpected: true,
		},
		{
			name:        "multiple different separators in row",
			value:       "my_.funcblah",
			errExpected: true,
		},
		{
			name:        "multiple separators in the name",
			value:       "my_func-blah.test",
			errExpected: false,
		},
	}

	t.Log("Test func name validation")
	{
		for i, tst := range tc {
			t.Logf("\tTest %d: \t%s", i, tst.name)
			{
				err := commands.ValidateFuncName(tst.value)
				if err != nil {
					if !tst.errExpected {
						t.Fatalf("\tValidateFuncName failed : got unexpected error [%v]\n", err)
					}
				} else {
					if tst.errExpected {
						t.Fatalf("\tValidateFuncName failed : error expected for value '%s' but was nil\n", tst.value)
					}
				}
			}
		}
	}
}
