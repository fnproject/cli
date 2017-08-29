package tests

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"testing"
)

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
	mylog(t, "INFO", testname, testname+" SUCCESS")
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
	mylog(t, "INFO", testname, testname+" SUCCESS")
}

func cleanup(t *testing.T, currdir, testdir string) {
	err := os.Chdir(currdir)
	if err != nil {
		t.Fatalf("ERROR: Failed to cd "+currdir+" directory: err: %v", err)
	}
	os.Remove(testdir)
}

func setupTestFiles(t *testing.T,
	testname, currdir, testdir string, testfiles []string) string {

	mylog(t, "INFO", testname, "Current directory is "+currdir)

	testfilesdir := path.Join(currdir, "testfiles")
	for _, testfile := range testfiles {
		err := copyFile(path.Join(testfilesdir, testfile), path.Join(testdir, testfile))
		if err != nil {
			t.Fatalf("ERROR: Failed to copy "+testfile+" to test directory: err: %v", err)
		}
	}

	fnTestBin := path.Join(testdir, "fn")
	err := os.Chdir("../../")
	if err != nil {
		t.Fatalf("ERROR: Failed to cd ../../ directory: err: %v", err)
	}
	res, err := exec.Command("go", "build", "-o", fnTestBin).CombinedOutput()
	if err != nil {
		mylog(t, "INFO", testname, string(res))
		t.Fatalf("ERROR: Failed to build fn: err: %v", err)
	}

	mylog(t, "INFO", testname, "cd test dir "+testdir)
	if err := os.Chdir(testdir); err != nil {
		t.Fatalf("ERROR: Failed to cd test dir "+testdir+": err: %v", err)
	}
	return fnTestBin
}

func runFnInit(t *testing.T, testname, fnTestBin string) {
	dockeruser := os.Getenv("DOCKER_USER")
	if dockeruser == "" {
		t.Fatalf("ERROR: DOCKER_USER not set")
	}
	mylog(t, "INFO", testname, "DOCKER_USER= "+dockeruser)
	var imagename string = dockeruser + "/" + "fn_test_hello_docker_runtime"

	mylog(t, "INFO", testname, "Run fn init "+imagename)
	res, err := exec.Command(fnTestBin, "init", imagename).CombinedOutput()
	if err != nil {
		mylog(t, "INFO", testname, string(res))
		t.Fatalf("ERROR: Failed run fn init: err: %v", err)
	}
}

func runFnBuild(t *testing.T, testname, fnTestBin string) {
	mylog(t, "INFO", testname, "Run fn build")
	res, err := exec.Command(fnTestBin, "build").CombinedOutput()
	if err != nil {
		mylog(t, "INFO", testname, string(res))
		t.Fatalf("ERROR: Failed run fn build: err: %v", err)
	}
}

func runFnBuildNoDockerfile(t *testing.T, testname, fnTestBin string) {
	mylog(t, "INFO", testname, "Run fn build")
	res, err := exec.Command(fnTestBin, "build").CombinedOutput()
	if err != nil {
		if bytes.Contains(res, []byte("Dockerfile not exists")) {
			mylog(t, "INFO", testname, "Got expected error message: "+string(res))
			return
		}
	}
	mylog(t, "INFO", testname, string(res))
	t.Fatalf("ERROR: Didn't get expected failure on fn build: res: %s, err: %v", string(res), err)
}

func runFnRun(t *testing.T, testname, fnTestBin string) {
	mylog(t, "INFO", testname, "Run fn run")
	res, err := exec.Command(fnTestBin, "run").CombinedOutput()
	if err != nil {
		mylog(t, "INFO", testname, string(res))
		t.Fatalf("ERROR: Failed run fn run: err: %v", err)
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

func mylog(t *testing.T, level, testname, msg string) {
	prefix := testname + " " + level + ": "
	t.Log(prefix + msg)
}
