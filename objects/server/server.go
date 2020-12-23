package server

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/fnxproject/cli/common"
	"github.com/urfave/cli/v2"
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
		log.Fatalln("Starting command failed:", err)
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
		log.Println("Interrupt caught, exiting")
		err = cmd.Process.Kill()
		if err != nil {
			log.Println("Error: could not kill process")
		}
	case err := <-done:
		if err != nil {
			log.Println("Processed finished with error:", err)
		} else {
			log.Println("Process finished gracefully")
		}
	}
	return nil
}
