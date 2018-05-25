package main

import (
	"errors"
	"fmt"
	"os/exec"

	"github.com/urfave/cli"
)

func stopCmd() cli.Command {
	return cli.Command{
		Name:   "stop",
		Usage:  "stops a functions server",
		Action: stop,
	}
}
func stop(c *cli.Context) error {
	cmd := exec.Command("docker", "stop", "fnserver")
	err := cmd.Run()
	if err != nil {
		return errors.New("failed to stop 'fnserver'")
	}

	fmt.Println("Successfully stopped 'fnserver'")

	return err
}
