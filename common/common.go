package common

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"time"

	"github.com/fatih/color"
)

func UberExec(verbose bool, dir string, cmd string, args []string) error {
	cancel := make(chan os.Signal, 3)
	signal.Notify(cancel, os.Interrupt) // and others perhaps
	defer signal.Stop(cancel)

	result := make(chan error, 1)

	buildOut := ioutil.Discard
	buildErr := ioutil.Discard

	quit := make(chan struct{})
	if verbose {
		fmt.Println()
		buildOut = os.Stdout
		buildErr = os.Stderr
	} else {
		// print dots. quit channel explanation: https://stackoverflow.com/a/16466581/105562
		ticker := time.NewTicker(1 * time.Second)
		go func() {
			for {
				select {
				case <-ticker.C:
					fmt.Fprintf(os.Stderr, ".")
				case <-quit:
					ticker.Stop()
					return
				}
			}
		}()
	}

	go func(done chan<- error) {
		cmd := exec.Command(cmd, args...)
		cmd.Dir = dir
		cmd.Stderr = buildErr // Doesn't look like there's any output to stderr on docker build, whether it's successful or not.
		cmd.Stdout = buildOut
		done <- cmd.Run()
	}(result)

	select {
	case err := <-result:
		close(quit)
		fmt.Fprintln(os.Stderr)
		if err != nil {
			fmt.Printf("%v Run with `--verbose` flag to see what went wrong. eg: `fn --verbose CMD`\n", color.RedString("Error during build."))
			return fmt.Errorf("error running docker build: %v", err)
		}
	case signal := <-cancel:
		close(quit)
		fmt.Fprintln(os.Stderr)
		return fmt.Errorf("build cancelled on signal %v", signal)
	}
	return nil
}
