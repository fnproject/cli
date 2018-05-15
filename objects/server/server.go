package server

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/fnproject/cli/common"
	"github.com/urfave/cli"
)

func update(c *cli.Context) error {
	args := []string{"pull",
		common.FunctionsDockerImage,
	}
	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		log.Fatalln("starting command failed:", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	// catch ctrl-c and kill
	sigC := make(chan os.Signal, 2)
	signal.Notify(sigC, os.Interrupt, syscall.SIGTERM)
	select {
	case <-sigC:
		log.Println("interrupt caught, exiting")
		err = cmd.Process.Kill()
		if err != nil {
			log.Println("error: could not kill process")
		}
	case err := <-done:
		if err != nil {
			log.Println("processed finished with error:", err)
		} else {
			log.Println("process finished gracefully")
		}
	}
	return nil
}
