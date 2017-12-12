package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
)

var fnTestBin string

// setup
func init() {
	fnTestBin = path.Join(os.TempDir(), "fn-test")
	res, err := exec.Command("go", "build", "-o", fnTestBin).CombinedOutput()
	fmt.Println(string(res))
	if err != nil {
		log.Fatal(err)
	}
}

func cdToTmp(t *testing.T) string {
	tmp := os.TempDir() + "/functest"
	var err error
	if _, err = os.Stat(tmp); os.IsExist(err) {
		err := os.RemoveAll(tmp)
		if err != nil {
			t.Fatal(err)
		}
	}
	err = os.MkdirAll(tmp, 0700)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	return tmp
}

func TestMainCommands(t *testing.T) {
	testCommands := []string{
		"init",
		"apps",
		"routes",
		"images",
		"lambda",
		"version",
		"build",
		"bump",
		"deploy",
		"run",
		"push",
		"logs",
		"calls",
		"call",
	}
	tmp := cdToTmp(t)
	defer os.RemoveAll(tmp)

	for _, cmd := range testCommands {
		res, err := exec.Command(fnTestBin, strings.Split(cmd, " ")...).CombinedOutput()
		if bytes.Contains(res, []byte("command not found")) {
			t.Error(err)
		}
	}
}

// very simple non-failure test for java versioning on fn init
func TestJavaVersioning(t *testing.T) {
	testname := "TestJavaVersioning"

	testdir, err := ioutil.TempDir("", testname)
	if err != nil {
		t.Fatalf("ERROR: Failed to make tmp test directory: err: %v", err)
	}
	defer os.RemoveAll(testdir)

	testsub1, err := ioutil.TempDir(testdir, "8")
	if err != nil {
		t.Fatalf("ERROR: Failed to make tmp test directory: err: %v", err)
	}
	checkInit(t, testsub1, "8", fnTestBin)

	testsub2, err := ioutil.TempDir(testdir, "9")
	if err != nil {
		t.Fatalf("ERROR: Failed to make tmp test directory: err: %v", err)
	}
	checkInit(t, testsub2, "9", fnTestBin)

	testsub3, err := ioutil.TempDir(testdir, "noversion")
	if err != nil {
		t.Fatalf("ERROR: Failed to make tmp test directory: err: %v", err)
	}
	checkInit(t, testsub3, "", fnTestBin)
}

func checkInit(t *testing.T, testdir, version, bin string) {
	err := os.Chdir(testdir)
	if err != nil {
		t.Errorf("ERROR: Failed to cd to tmp test directory: err: %v", err)
	}

	runtime := fmt.Sprintf("--runtime=java%s", version)

	_, err = exec.Command(bin, "init", runtime).CombinedOutput()
	if err != nil {
		t.Errorf("ERROR: Failed to run fn init with --runtime=java%s. err: %v", version, err)
	}

	if _, err = os.Stat(fmt.Sprintf("%s/src/main", testdir)); err != nil && os.IsNotExist(err) {
		t.Errorf("ERROR: failed to create java project with --runtime=java%s. err: %v", version, err)
	}
}
