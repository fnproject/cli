package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/fnproject/cli/client"
	functions "github.com/funcy/functions_go"
	"github.com/onsi/gomega"
	"github.com/urfave/cli"
)

type testStruct struct {
	Tests []fftest `yaml:"tests,omitempty" json:"tests,omitempty"`
}

func testfn() cli.Command {
	cmd := testcmd{RoutesApi: functions.NewRoutesApi()}
	return cli.Command{
		Name:   "test",
		Usage:  "run functions test if present",
		Flags:  cmd.flags(),
		Action: cmd.test,
	}
}

type testcmd struct {
	*functions.RoutesApi

	build  bool
	remote string
}

func (t *testcmd) flags() []cli.Flag {
	return []cli.Flag{
		// cli.BoolFlag{
		// 	Name:        "b",
		// 	Usage:       "build before test",
		// 	Destination: &t.build,
		// },
		cli.StringFlag{
			Name:        "remote",
			Usage:       "run tests by calling the function on `appname`",
			Destination: &t.remote,
		},
	}
}

func (t *testcmd) test(c *cli.Context) error {
	gomega.RegisterFailHandler(func(message string, callerSkip ...int) {
		fmt.Println("In gomega FailHandler:", message)
	})

	wd := getWd()

	fpath, ff, err := findAndParseFuncfile(wd)
	if err != nil {
		return err
	}
	// get name from directory if it's not defined
	if ff.Name == "" {
		ff.Name = filepath.Base(filepath.Dir(fpath)) // todo: should probably make a copy of ff before changing it
	}

	ff, err = buildfunc(fpath, ff, false)
	ff, envVars, err := preRun(c)
	if err != nil {
		return err
	}

	var tests []fftest

	// Look for test.json file too
	tfile := "test.json"
	if exists(tfile) {
		f, err := os.Open(tfile)
		if err != nil {
			return fmt.Errorf("could not open %s for parsing. %v", tfile, err)
		}
		ts := &testStruct{}
		err = json.NewDecoder(f).Decode(ts)
		if err != nil {
			fmt.Println("Invalid tests.json file:", err)
			return err
		}
		tests = ts.Tests
	} else {
		tests = ff.Tests
	}
	if len(tests) == 0 {
		return errors.New("no tests found for this function")
	}

	fmt.Printf("Running %v tests...", len(tests))

	target := ff.ImageName()
	runtest := runlocaltest
	if t.remote != "" {
		if ff.Path == "" {
			return errors.New("execution of tests on remote server demand that this function has a `path`.")
		}
		if err := resetBasePath(t.Configuration); err != nil {
			return fmt.Errorf("error setting endpoint: %v", err)
		}
		baseURL, err := url.Parse(t.Configuration.BasePath)
		if err != nil {
			return fmt.Errorf("error parsing base path: %v", err)
		}

		u, err := url.Parse("../")
		u.Path = path.Join(u.Path, "r", t.remote, ff.Path)
		target = baseURL.ResolveReference(u).String()
		runtest = runremotetest
	}

	errorCount := 0
	fmt.Println("running tests on", ff.ImageName(), ":")
	for i, tt := range tests {
		fmt.Printf("\nTest %v\n", i+1)
		start := time.Now()
		var err error
		err = runtest(target, tt.Input, tt.Output, tt.Err, envVars)
		if err != nil {
			fmt.Print("FAILED")
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
			fmt.Print("PASSED")
		}
		fmt.Println(" - ", tt.Name, " (", time.Since(start), ")")

	}
	fmt.Printf("\n%v tests passed, %v tests failed.\n", len(tests)-errorCount, errorCount)
	if errorCount > 0 {
		return errors.New("tests failed, errors found")
	}
	return nil
}

func runlocaltest(target string, in *inputMap, expectedOut *outputMap, expectedErr *string, envVars []string) error {
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

	ff := &Funcfile{Name: target}
	if err := runff(ff, stdin, &stdout, &stderr, "", envVars, nil, DefaultFormat, 1); err != nil {
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

func runremotetest(target string, in *inputMap, expectedOut *outputMap, expectedErr *string, envVars []string) error {
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

	if err := client.CallFN(target, stdin, &stdout, "", envVars, false); err != nil {
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
