package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
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
				Usage: "--log-level DEBUG to enable debugging",
			},
			cli.StringFlag{
				Name:  "docker-mode",
				Value: "docker-in-docker",
				Usage: "docker-mode - docker-in-docker or docker-bind-socket",
			},
			cli.BoolFlag{
				Name:  "detach, d",
				Usage: "Run container in background.",
			},
			cli.IntFlag{
				Name:  "port, p",
				Value: 8080,
				Usage: "Specify port number to bind to on the host.",
			},
		},
	}
}

func start(c *cli.Context) error {
	denvs := []string{}
	if c.String("log-level") != "" {
		denvs = append(denvs, "GIN_MODE="+c.String("log-level"))
	}

	// docker-in-docker or socker bind mount?
	isDind := true
	if c.String("docker-mode") == "docker-in-docker" && runtime.GOOS == "windows" {
		log.Println("docker-in-docker unavailable in Windows, auto-reverting to docker-bind-socket mode")
		isDind = false
	} else if c.String("docker-mode") == "docker-bind-socket" {
		isDind = false
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln("Getwd failed:", err)
	}

	args := []string{"run", "--rm", "-i",
		"--name", "functions",
		"-v", fmt.Sprintf("%s/data:/app/data", wd),
		"-p", fmt.Sprintf("%d:8080", c.Int("port")),
	}

	if isDind {
		args = append(args, "--privileged")
	} else {
		args = append(args, "-v", "/var/run/docker.sock:/var/run/docker.sock")
	}

	for _, v := range denvs {
		args = append(args, "-e", v)
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

	select {
	case <-sigC:
		log.Println("interrupt caught, exiting")
		err = cmd.Process.Kill()
		if err != nil {
			log.Println("error: could not kill process:", err)
		}
	case err := <-done:
		if err != nil {
			log.Println("error: processed finished with error", err)
		}
	}
	return nil
}
