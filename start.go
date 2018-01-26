package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/urfave/cli"
)

func startCmd() cli.Command {
	return cli.Command{
		Name:   "start",
		Usage:  "start a functions server",
		Action: start,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "log-level",
				Usage: "--log-level debug to enable debugging",
			},
			cli.BoolFlag{
				Name:  "detach, d",
				Usage: "Run container in background.",
			},
			cli.StringFlag{
				Name:  "env-file",
				Usage: "Path to Fn server configuration file.",
			},
		},
	}
}

func start(c *cli.Context) error {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln("Getwd failed:", err)
	}
	args := []string{"run", "--rm", "-i",
		"--name", "fnserver",
		"-v", fmt.Sprintf("%s/data:/app/data", wd),
		"-v", "/var/run/docker.sock:/var/run/docker.sock",
		"--privileged",
		"-p", "8080:8080",
		"--entrypoint", "./fnserver",
	}
	if c.String("log-level") != "" {
		args = append(args, "-e", fmt.Sprintf("FN_LOG_LEVEL=%v", c.String("log-level")))
	}
	if c.String("env-file") != "" {
		args = append(args, "--env-file", c.String("env-file"))
	}
	if c.Bool("detach") {
		args = append(args, "-d")
	}
	args = append(args, functionsDockerImage)
	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
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

	for {
		select {
		case <-sigC:
			log.Println("interrupt caught, exiting")
			err = cmd.Process.Signal(syscall.SIGTERM)
			if err != nil {
				log.Println("error: could not kill process:", err)
				return err
			}
		case err := <-done:
			if err != nil {
				log.Println("error: processed finished with error", err)
			}
		}
		return err
	}

	return nil
}
