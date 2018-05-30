package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/fnproject/cli/client"
	"github.com/onsi/gomega"
	"github.com/urfave/cli"
	"github.com/fnproject/fn_go/provider"
)

type testStruct struct {
	Tests []fftest `yaml:"tests,omitempty" json:"tests,omitempty"`
}

func testfn() cli.Command {
	cmd := testcmd{}
	return cli.Command{
		Name:   "test",
		Usage:  "run functions test if present",
		Flags:  cmd.flags(),
		Action: cmd.test,
		Before: func(cxt *cli.Context) error {
			prov,err := client.CurrentProvider()
			if err !=nil {
				return err
			}
			cmd.provider = prov
			return nil

		},
	}
}

type testcmd struct {
	provider provider.Provider
	build  bool
	remote string
}

func (t *testcmd) flags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:        "remote",
			Usage:       "run tests against a remote fn server",
			Destination: &t.remote,
		},
		cli.BoolFlag{
			Name:  "all",
			Usage: "if in root directory containing `app.yaml`, this will deploy all functions",
		},
	}
}

func (t *testcmd) test(c *cli.Context) error {
	gomega.RegisterFailHandler(func(message string, callerSkip ...int) {
		fmt.Println("In gomega FailHandler:", message)
	})

	wd := getWd()

	if c.Bool("all") {
		fmt.Println("Testing all functions in this directory and all sub directories.")
		return t.testAll(c, wd)
	}

	_, _, err := t.testSingle(c, wd)
	return err
}

func (t *testcmd) testAll(c *cli.Context, wd string) error {
	testCount := 0
	errorCount := 0
	err := walkFuncs(wd, func(path string, ff *funcfile, err error) error {
		if err != nil { // probably some issue with funcfile parsing, can decide to handle this differently if we'd like
			return err
		}
		dir := filepath.Dir(path)
		if dir == wd {
			// TODO: needed for tests?
			// setRootFuncInfo(ff, appName)
		} else {
			// change dirs
			err = os.Chdir(dir)
			if err != nil {
				return err
			}
			p2 := strings.TrimPrefix(dir, wd)
			if ff.Name == "" {
				ff.Name = strings.Replace(p2, "/", "-", -1)
				if strings.HasPrefix(ff.Name, "-") {
					ff.Name = ff.Name[1:]
				}
				// todo: should we prefix appname too?
			}
			if ff.Path == "" {
				ff.Path = p2
			}
		}
		tc, ec, err := t.testSingle(c, dir)
		if err != nil {
			fmt.Printf("test error on %s: %v\n", path, err)
			// TOOD: store these logs and print them at the end?
		}
		testCount += tc
		errorCount += ec
		now := time.Now()
		os.Chtimes(path, now, now)
		// funcFound = true
		return nil
	})
	if err != nil {
		return err
	}
	errmsg := "0 FAILED"
	if errorCount > 0 {
		errmsg = color.RedString(fmt.Sprintf("%v FAILED", errorCount))
	}
	passed := testCount - errorCount
	fmt.Printf("\nAll %v tests finished.\n%v\n%v\n", testCount, color.GreenString(fmt.Sprintf("%v PASSED", passed)), errmsg)
	if errorCount > 0 {
		return fmt.Errorf("%v tests failed", errorCount)
	}
	return nil
}

func (t *testcmd) testSingle(c *cli.Context, wd string) (totalTests, errorCount int, err error) {
	// TODO: prerun should take a wd
	fpath, ff, envVars, err := preRun(c)
	if err != nil {
		return 0, 0, err
	}

	// get name from directory if it's not defined
	if ff.Name == "" {
		ff.Name = filepath.Base(filepath.Dir(fpath)) // todo: should probably make a copy of ff before changing it
	}

	var tests []fftest

	// Look for test.json file too
	tfile := "test.json"
	if exists(tfile) {
		f, err := os.Open(tfile)
		if err != nil {
			return 0, 0, fmt.Errorf("could not open %s for parsing. %v", tfile, err)
		}
		ts := &testStruct{}
		err = json.NewDecoder(f).Decode(ts)
		if err != nil {
			fmt.Println("Invalid tests.json file:", err)
			return 0, 0, err
		}
		tests = ts.Tests
	} else {
		tests = ff.Tests
	}
	if len(tests) == 0 {
		return 0, 0, errors.New("no tests found for this function")
	}

	runtest := runlocaltest
	if t.remote != "" {
		runtest = t.runremotetest
	}

	// todo: make path here relative to the app root
	fmt.Printf("Running %v tests on %v (image: %v):", len(tests), fpath, ff.ImageName())
	for i, tt := range tests {
		fmt.Printf("\nTest %v\n", i+1)
		start := time.Now()
		var err error
		err = runtest(ff, tt.Input, tt.Output, tt.Err, envVars)
		if err != nil {
			fmt.Print(color.RedString("FAILED"))
			errorCount++
			scanner := bufio.NewScanner(strings.NewReader(err.Error()))
			for scanner.Scan() {
				fmt.Println("\t\t", scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				fmt.Fprintln(os.Stderr, "reading test result:", err)
				break
			}
		} else {
			fmt.Print(color.GreenString("PASSED"))
		}
		fmt.Println(" - ", tt.Name, " (", time.Since(start), ")")

	}
	passed := len(tests) - errorCount
	errmsg := "0 failed"
	if errorCount > 0 {
		errmsg = color.RedString(fmt.Sprintf("%v failed", errorCount))
	}
	fmt.Printf("\ntests run: %v, %v\n", color.GreenString(fmt.Sprintf("%v passed", passed)), errmsg)
	if errorCount > 0 {
		return len(tests), errorCount, fmt.Errorf("%v tests failed", errorCount)
	}
	return len(tests), errorCount, nil
}

func runlocaltest(ff *funcfile, in *inputMap, expectedOut *outputMap, expectedErr *string, envVars []string) error {
	inBytes, err := json.Marshal(in.Body)
	if err != nil {
		return err
	}
	if string(inBytes) == "\"\"" {
		// marshalling this: `"body": ""` turns into double quotes, not an empty string as you might expect.
		// may be a better way to handle this?
		inBytes = []byte{} // empty string
	}
	stdin := &bytes.Buffer{}
	if in != nil {
		stdin = bytes.NewBuffer(inBytes)
	}
	expectedB, err := json.Marshal(expectedOut.Body)
	if err != nil {
		return err
	}
	expectedString := string(expectedB)

	var stdout, stderr bytes.Buffer

	if err := runff(ff, stdin, &stdout, &stderr, "", envVars, nil, "", 1, "application/json"); err != nil {
		return fmt.Errorf("%v\nstdout:%s\nstderr:%s\n", err, stdout.String(), stderr.String())
	}

	out := stdout.String()
	if expectedOut == nil && out != "" {
		return fmt.Errorf("unexpected output found: %s", out)
	}
	if gomega.Expect(out).To(gomega.MatchJSON(expectedString)) {
		// PASS!
		return nil
	}

	return fmt.Errorf("mismatched output found.\nexpected:\n%s\ngot:\n%s\nlogs:\n%s\n", expectedString, out, stderr.String())
}

func (t *testcmd) runremotetest(ff *funcfile, in *inputMap, expectedOut *outputMap, expectedErr *string, envVars []string) error {
	if ff.Path == "" {
		return errors.New("execution of tests on remote server demand that this function has a `path`")
	}

	inBytes, err := json.Marshal(in)
	if err != nil {
		return err
	}
	stdin := &bytes.Buffer{}
	if in != nil {
		stdin = bytes.NewBuffer(inBytes)
	}
	expectedString, err := json.Marshal(expectedOut.Body)
	if err != nil {
		return err
	}
	var stdout bytes.Buffer

	if err := client.CallFN(t.provider,t.remote, ff.Path, stdin, &stdout, "", envVars, "application/json", false); err != nil {
		return fmt.Errorf("%v\nstdout:%s\n", err, stdout.String())
	}

	out := stdout.String()
	if expectedOut == nil && out != "" {
		return fmt.Errorf("unexpected output found: %s", out)
	}
	if gomega.Expect(out).To(gomega.MatchJSON(expectedString)) {
		// PASS!
		return nil
	}

	return nil
}
