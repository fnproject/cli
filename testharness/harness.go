package testharness

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/jmoiron/jsonq"
)

// Max duration a command can run for before being killed
var commandTimeout = 5 * time.Minute

//CLIHarness encapsulates a single CLI session
type CLIHarness struct {
	t        *testing.T
	cliPath  string
	appNames []string
	testDir  string
	homeDir  string
	cwd      string

	env     map[string]string
	history []string
}

//CmdResult wraps the result of a single command - this includes the diagnostic history of commands that led to this command in a given context for easier test diagnosis
type CmdResult struct {
	t               *testing.T
	OriginalCommand string
	Cwd             string
	ExtraEnv        []string
	Stdout          string
	Stderr          string
	Input           string
	ExitState       *os.ProcessState
	Success         bool
	History         []string
}

func (cr *CmdResult) String() string {
	return fmt.Sprintf(`COMMAND: %s
INPUT:%s
RESULT: %t
STDERR:
%s
STDOUT:%s
HISTORY:
%s
EXTRAENV:
%s
EXITSTATE: %s
`, cr.OriginalCommand,
		cr.Input,
		cr.Success,
		cr.Stderr,
		cr.Stdout,
		strings.Join(cr.History, "\n"),
		strings.Join(cr.ExtraEnv, "\n"),
		cr.ExitState)
}

//AssertSuccess checks the command was success
func (cr *CmdResult) AssertSuccess() *CmdResult {
	if !cr.Success {
		cr.t.Fatalf("Command failed but should have succeeded: \n%s", cr.String())
	}
	return cr
}

// AssertStdoutContains asserts that the string appears somewhere in the stdout
func (cr *CmdResult) AssertStdoutContains(match string) *CmdResult {
	if !strings.Contains(cr.Stdout, match) {
		log.Fatalf("Expected stdout  message (%s) not found in result: %v", match, cr)
	}
	return cr
}

// AssertStdoutContains asserts that the string appears somewhere in the stderr
func (cr *CmdResult) AssertStderrContains(match string) *CmdResult {
	if !strings.Contains(cr.Stderr, match) {
		log.Fatalf("Expected sdterr message (%s) not found in result: %v", match, cr)
	}
	return cr
}

// AssertFailed asserts that the command did not succeed
func (cr *CmdResult) AssertFailed() *CmdResult {
	if cr.Success {
		cr.t.Fatalf("Command succeeded but should have failed: \n%s", cr.String())
	}
	return cr
}

// AssertStdoutEmpty fails if there was output to stdout
func (cr *CmdResult) AssertStdoutEmpty() {
	if cr.Stdout != "" {
		cr.t.Fatalf("Expecting empty stdout, got %v", cr)
	}
}

func randString(n int) string {

	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// Create creates a CLI test harness that that runs CLI operations in a test directory
// test harness operations will propagate most environment variables to tests (with the exception of HOME, which is faked)
func Create(t *testing.T) *CLIHarness {
	testDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal("Failed to create temp dir")
	}

	homeDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal("Failed to create home dir")
	}

	cliPath := os.Getenv("TEST_CLI_BINARY")

	if cliPath == "" {
		wd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get CWD, %v", err)
		}

		cliPath = path.Join(wd, "../fn")
	}
	ctx := &CLIHarness{
		t:       t,
		cliPath: cliPath,
		testDir: testDir,
		homeDir: homeDir,
		cwd:     testDir,
		env: map[string]string{
			"HOME": homeDir,
		},
	}
	ctx.pushHistoryf("cd %s", ctx.cwd)
	return ctx
}

// Cleanup removes any temporary files and tries to delete any apps that (may) have been created during a test
func (h *CLIHarness) Cleanup() {

	h.Cd("")
	for _, app := range h.appNames {
		h.Fn("delete", "apps", app)
	}

	os.RemoveAll(h.testDir)
	os.RemoveAll(h.homeDir)

}

//NewAppName creates a new, valid app name and registers it for deletion
func (h *CLIHarness) NewAppName() string {
	appName := randString(8)
	h.appNames = append(h.appNames, appName)
	return appName
}

//WithEnv sets additional enironment variables in the test , these overlay the ambient environment
func (h *CLIHarness) WithEnv(key string, value string) {
	h.env[key] = value
}

func copyAll(src, dest string) error {
	srcinfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if srcinfo.IsDir() {

		os.MkdirAll(dest, srcinfo.Mode())
		directory, err := os.Open(src)
		if err != nil {
			return fmt.Errorf("Failed to open directory %s: %v ", src, err)

		}

		objects, err := directory.Readdir(-1)
		if err != nil {
			return fmt.Errorf("Failed to read directory %s: %v ", src, err)
		}

		for _, obj := range objects {
			srcPath := path.Join(src, obj.Name())
			destPath := path.Join(dest, obj.Name())
			err := copyAll(srcPath, destPath)
			if err != nil {
				return err
			}
		}
	} else {

		dstDir := filepath.Dir(dest)
		srcDir := filepath.Dir(src)

		srcDirInfo, err := os.Stat(srcDir)
		if err != nil {
			return err
		}

		os.MkdirAll(dstDir, srcDirInfo.Mode())

		b, err := ioutil.ReadFile(src)
		if err != nil {
			return fmt.Errorf("Failed to read src file %s: %v ", src, err)
		}

		err = ioutil.WriteFile(dest, b, srcinfo.Mode())
		if err != nil {
			return fmt.Errorf("Failed to read dst file %s: %v ", dest, err)
		}
	}
	return nil
}

// CopyFiles copies files and directories from the test source dir into the testing root directory
func (h *CLIHarness) CopyFiles(files map[string]string) {

	for src, dest := range files {
		h.pushHistoryf("cp -r %s %s", src, dest)
		err := copyAll(src, path.Join(h.testDir, dest))
		if err != nil {
			h.t.Fatalf("Failed to copy %s -> %s : %v", src, dest, err)
		}
	}

}

// WithFile creates a file relative to the cwd
func (h *CLIHarness) WithFile(rPath string, content string, perm os.FileMode) {

	fullPath := h.relativeToCwd(rPath)

	err := ioutil.WriteFile(fullPath, []byte(content), perm)
	if err != nil {
		h.t.Fatalf("Failed to create file %s", fullPath)
	}
	h.pushHistoryf("echo `%s` > %s", content, fullPath)

}

// FnWithInput runs the Fn ClI with an input string
// If a command takes more than a certain timeout then this will send a SIGQUIT to the process resulting in a stacktrace on stderr
func (h *CLIHarness) FnWithInput(input string, args ...string) *CmdResult {

	stdOut := bytes.Buffer{}
	stdErr := bytes.Buffer{}

	args = append([]string{"--verbose"}, args...)
	cmd := exec.Command(h.cliPath, args...)
	cmd.Stderr = &stdErr
	cmd.Stdout = &stdOut

	stdIn := bytes.NewBufferString(input)

	cmd.Dir = h.cwd
	envRegex := regexp.MustCompile("([^=]+)=(.*)")

	mergedEnv := map[string]string{}

	for _, e := range os.Environ() {
		m := envRegex.FindStringSubmatch(e)
		if len(m) != 3 {
			panic("Invalid env entry")
		}
		mergedEnv[m[1]] = m[2]
	}

	extraEnv := make([]string, 0, len(h.env))

	for k, v := range h.env {
		mergedEnv[k] = v
		extraEnv = append(extraEnv, fmt.Sprintf("%s=%s", k, v))
	}
	env := make([]string, 0, len(mergedEnv))

	for k, v := range mergedEnv {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	cmd.Env = env
	cmd.Stdin = stdIn
	cmdString := h.cliPath + " " + strings.Join(args, " ")

	if input != "" {
		h.pushHistoryf("echo '%s' | %s", input, cmdString)
	} else {
		h.pushHistoryf("%s", cmdString)
	}
	done := make(chan interface{})
	timer := time.NewTimer(commandTimeout)

	// If the CLI stalls for more than commandTimeout we send a SIQQUIT which should result in a stack trace in stderr
	go func() {
		select {
		case <-done:
			return
		case <-timer.C:
			h.t.Errorf("Command timed out - killing CLI with SIGQUIT - see STDERR log for stack trace of where it was stalled")

			cmd.Process.Signal(syscall.SIGQUIT)
		}
	}()

	err := cmd.Run()
	close(done)

	cmdResult := &CmdResult{
		OriginalCommand: cmdString,
		Stdout:          stdOut.String(),
		Stderr:          stdErr.String(),
		ExtraEnv:        extraEnv,
		Cwd:             h.cwd,
		Input:           input,
		History:         h.history,
		ExitState:       cmd.ProcessState,
		t:               h.t,
	}

	if err, ok := err.(*exec.ExitError); ok {
		cmdResult.Success = false
	} else if err != nil {
		h.t.Fatalf("Failed to run cmd %v :  %v", args, err)
	} else {
		cmdResult.Success = true
	}

	return cmdResult
}

// Fn runs the Fn ClI with the specified arguments
func (h *CLIHarness) Fn(args ...string) *CmdResult {
	return h.FnWithInput("", args...)
}

// Writes the relevant files to the CWD to produce the smallest function that can be written
func (h *CLIHarness) WithMinimalFunctionSource() *CLIHarness {

	const dockerFile = `FROM busybox:1.28.3
RUN	mkdir /app
ADD	main.sh /app
WORKDIR /app
CMD ["./main.sh"]
`
	const mainSh = `#!/bin/sh
echo "hello world";
`

	const funcYaml = `version: 0.0.1
runtime: docker
`

	h.WithFile("func.yaml", funcYaml, 0644)
	h.WithFile("Dockerfile", dockerFile, 0644)
	h.WithFile("main.sh", mainSh, 0755)

	return h
}

//NewFuncName creates a valid function name
func (h *CLIHarness) NewFuncName() string {
	return randString(8)
}

func (h *CLIHarness) relativeToTestDir(dir string) string {
	absDir, err := filepath.Abs(path.Join(h.testDir, dir))
	if err != nil {
		h.t.Fatalf("Invalid path operation : %v", err)
	}

	if !strings.HasPrefix(absDir, h.testDir) {
		h.t.Fatalf("Cannot change directory to %s out of test directory %s", absDir, h.testDir)
	}
	return absDir
}

func (h *CLIHarness) relativeToCwd(dir string) string {
	absDir, err := filepath.Abs(path.Join(h.cwd, dir))
	if err != nil {
		h.t.Fatalf("Invalid path operation : %v", err)
	}

	if !strings.HasPrefix(absDir, h.testDir) {
		h.t.Fatalf("Cannot change directory to %s out of test directory %s", absDir, h.testDir)
	}
	return absDir
}

// Cd Changes the working directory for commands - passing "" resets this to the test directory
// You cannot Cd out of the test directory
func (h *CLIHarness) Cd(s string) {

	if s == "" {
		h.cwd = h.testDir
	} else {
		h.cwd = h.relativeToCwd(s)
	}

	h.pushHistoryf("cd %s", h.cwd)

}
func (h *CLIHarness) pushHistoryf(s string, args ...interface{}) {
	//log.Printf(s, args...)
	h.history = append(h.history, fmt.Sprintf(s, args...))

}

// MkDir creates a directory in the current cwd
func (h *CLIHarness) MkDir(dir string) {
	os.Mkdir(h.relativeToCwd(dir), 0777)

}

//FileAppend appends val to  an existing file
func (h *CLIHarness) FileAppend(file string, val string) {
	filePath := h.relativeToCwd(file)
	fileV, err := ioutil.ReadFile(filePath)

	if err != nil {
		h.t.Fatalf("Failed to read file %s: %v", file, err)
	}

	newV := string(fileV) + val
	err = ioutil.WriteFile(filePath, []byte(newV), 0555)
	if err != nil {
		h.t.Fatalf("Failed to write appended file %s", err)
	}

	h.pushHistoryf("echo '%s' >> %s", val, filePath)

}

// GetFile dumps the content of a file (relative to the  CWD)
func (h *CLIHarness) GetFile(s string) string {
	v, err := ioutil.ReadFile(h.relativeToCwd(s))
	if err != nil {
		h.t.Fatalf("File %s is not readable %v", s, err)
	}
	return string(v)

}

func (cr *CmdResult) AssertStdoutContainsJSON(query []string, value interface{}) {
	routeObj := map[string]interface{}{}
	err := json.Unmarshal([]byte(cr.Stdout), &routeObj)
	if err != nil {
		log.Fatalf("Failed to parse routes inspect as JSON %v, %v", err, cr)
	}

	q := jsonq.NewQuery(routeObj)

	val, err := q.Interface(query...)
	if err != nil {
		log.Fatalf("Failed to find path %v in json body %v", query, cr.Stdout)
	}

	if val != value {
		log.Fatalf("Expected %s to be %v but was %s, %v", strings.Join(query, "."), value, val, cr)
	}
}

func (cr *CmdResult) AssertStdoutMissingJSONPath(query []string) {
	routeObj := map[string]interface{}{}
	err := json.Unmarshal([]byte(cr.Stdout), &routeObj)
	if err != nil {
		log.Fatalf("Failed to parse routes inspect as JSON %v, %v", err, cr)
	}

	q := jsonq.NewQuery(routeObj)
	_, err = q.Interface(query...)
	if err == nil {
		log.Fatalf("Found path %v in json body %v when it was supposed to be missing", query, cr.Stdout)
	}
}
