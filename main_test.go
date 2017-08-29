package main

import (
	"bytes"
	"fmt"
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
