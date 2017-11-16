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
				Usage: "--log-level DEBUG to enable debugging",
			},
			cli.BoolFlag{
				Name:  "detach, d",
				Usage: "Run container in background.",
			},
			cli.BoolFlag{
				Name:  "selinux",
				Usage: "Run container with the right privileges on a SELinux system.",
			},
		},
	}
}

func start(c *cli.Context) error {
	denvs := []string{}
	if c.String("log-level") != "" {
		denvs = append(denvs, "GIN_MODE="+c.String("log-level"))
	}
	// Socket mount: docker run --rm -it --name functions -v ${PWD}/data:/app/data:Z -v /var/run/docker.sock:/var/run/docker.sock:Z -p 8080:8080 funcy/functions
	// OR dind: docker run --rm -it --name functions -v ${PWD}/data:/app/data:Z --privileged -p 8080:8080 funcy/functions
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln("Getwd failed:", err)
	}
	args := []string{"run", "--rm", "-i", "--name", "functions",
		"-v", "/var/run/docker.sock:/var/run/docker.sock",
		"-p", "8080:8080",
	}
	if c.Bool("selinux") {
		// Require security privileges on SELinux systems and securely mount the volume with :Z
		args = append(args, "--security-opt", "label=type:container_runtime_t")
		args = append(args, "-v", fmt.Sprintf("%s/data:/app/data:Z", wd))
	} else {
		// Just mount the volume as usual
		args = append(args, "-v", fmt.Sprintf("%s/data:/app/data", wd))
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
