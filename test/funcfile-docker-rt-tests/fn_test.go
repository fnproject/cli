package tests

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"testing"
)

func init() {
	dockerUser := os.Getenv("DOCKER_USER")
	var err error
	if dockerUser == "" {
		dockerUser = "funcster"
	}
	err = os.Setenv("FN_REGISTRY", dockerUser)
	if err != nil {
		log.Fatalf("couldn't set env var: %v", err)
	}
	return
}

func TestDockerRuntime(t *testing.T) {
	testname := "TestDockerRuntime"
	testfiles := []string{"Dockerfile", "func.go"}

	currdir, err := os.Getwd()
	if err != nil {
		t.Fatalf("ERROR: Failed to get current directory: err: %v", err)
	}
	testdir, err := ioutil.TempDir("", "cli_test_funcfile-docker-rt_"+testname)
	if err != nil {
		t.Fatalf("ERROR: Failed to make tmp test directory: err: %v", err)
	}
	defer cleanup(t, currdir, testdir)
	fnTestBin := setupTestFiles(t, testname, currdir, testdir, testfiles)
	runFnInit(t, testname, fnTestBin)
	runFnBuild(t, testname, fnTestBin)
	runFnRun(t, testname, fnTestBin)
	t.Logf("INFO: %s SUCCESS", testname)
}

func TestDockerRuntimeNoDockerfile(t *testing.T) {
	testname := "TestDockerRuntimeNoDockerfile"
	testfiles := []string{"func.yaml", "func.go"}

	currdir, err := os.Getwd()
	if err != nil {
		t.Fatalf("ERROR: Failed to get current directory: err: %v", err)
	}
	testdir, err := ioutil.TempDir("", "cli_test_funcfile-docker-rt_"+testname)
	if err != nil {
		t.Fatalf("ERROR: Failed to make tmp test directory: err: %v", err)
	}
	defer cleanup(t, currdir, testdir)
	fnTestBin := setupTestFiles(t, testname, currdir, testdir, testfiles)
	runFnBuildNoDockerfile(t, testname, fnTestBin)
	t.Logf("INFO: %s SUCCESS", testname)
}

func cleanup(t *testing.T, currdir, testdir string) {
	err := os.Chdir(currdir)
	if err != nil {
		t.Fatalf("ERROR: Failed to cd %s directory: err: %v", currdir, err)
	}
	os.Remove(testdir)
}

func setupTestFiles(t *testing.T,
	testname, currdir, testdir string, testfiles []string) string {

	t.Logf("INFO: %s Current directory is %s", testname, currdir)

	testfilesdir := path.Join(currdir, "testfiles")
	for _, testfile := range testfiles {
		err := copyFile(path.Join(testfilesdir, testfile), path.Join(testdir, testfile))
		if err != nil {
			t.Fatalf("ERROR: Failed to copy %s to test directory %s: err: %v", testfile, testdir, err)
		}
	}

	fnTestBin := path.Join(testdir, "fn")
	err := os.Chdir("../../")
	if err != nil {
		t.Fatalf("ERROR: Failed to cd ../../ directory: err: %v", err)
	}
	res, err := exec.Command("go", "build", "-o", fnTestBin).CombinedOutput()
	if err != nil {
		t.Fatalf("ERROR: Failed to build fn: res: %s, err: %v", string(res), err)
	}

	t.Logf("INFO: %s cd test directory %s", testname, testdir)
	if err := os.Chdir(testdir); err != nil {
		t.Fatalf("ERROR: Failed to cd test directory %s: err: %v", testdir, err)
	}
	return fnTestBin
}

func runFnInit(t *testing.T, testname, fnTestBin string) {
	var imagename string = "fn_test_hello_docker_runtime"

	t.Logf("INFO: %s Run fn init --name %s", testname, imagename)
	res, err := exec.Command(fnTestBin, "init", "--name", imagename).CombinedOutput()
	if err != nil {
		t.Fatalf("ERROR: Failed run fn init: res: %s, err: %v", string(res), err)
	}
}

func runFnBuild(t *testing.T, testname, fnTestBin string) {
	t.Logf("INFO: %s Run fn build", testname)
	res, err := exec.Command(fnTestBin, "build").CombinedOutput()
	if err != nil {
		t.Fatalf("ERROR: Failed run fn build: res: %s, err: %v", string(res), err)
	}
}

func runFnBuildNoDockerfile(t *testing.T, testname, fnTestBin string) {
	t.Logf("INFO: %s Run fn build", testname)
	res, err := exec.Command(fnTestBin, "build").CombinedOutput()
	if err != nil {
		if bytes.Contains(res, []byte("Dockerfile not exists")) {
			t.Logf("INFO: %s Got expected error message %s", testname, string(res))
			return
		}
	}
	t.Fatalf("ERROR: Didn't get expected failure on fn build: res: %s, err: %v", string(res), err)
}

func runFnRun(t *testing.T, testname, fnTestBin string) {
	t.Logf("INFO: %s Run fn run", testname)
	res, err := exec.Command(fnTestBin, "run").CombinedOutput()
	if err != nil {
		t.Fatalf("ERROR: Failed run fn run: res: %s, err: %v", string(res), err)
	}

}

func copyFile(src, dst string) error {
	b, err := ioutil.ReadFile(src)
	if err != nil {
		return fmt.Errorf("copyFile(): failed to read src file "+src+": %s", src)
	}

	err = ioutil.WriteFile(dst, b, 0777)
	if err != nil {
		return fmt.Errorf("copyFile(): failed to read dst file "+dst+": %s", dst)
	}
	return nil
}
