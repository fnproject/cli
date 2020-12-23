package commands

import (
	"errors"
	"fmt"
	"os/exec"

	"github.com/urfave/cli/v2"
)

// StopCommand returns stop server cli.command
func StopCommand() *cli.Command {
	return &cli.Command{
		Name:        "stop",
		Usage:       "Stop a function server",
		Category:    "SERVER COMMANDS",
		Description: "This command stops a Fn server.",
		Action:      stop,
	}
}
func stop(c *cli.Context) error {
	cmd := exec.Command("docker", "stop", "fnserver")
	err := cmd.Run()
	if err != nil {
		return errors.New("Failed to stop 'fnserver'")
	}

	fmt.Println("Successfully stopped 'fnserver'")

	return err
}
