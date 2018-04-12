package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
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
	cname := "fnserver"
	args := []string{"run", "--rm", "-i",
		"--name", cname,
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

	// check for newer version:
	go checkForNewerVersion(cname)

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

type TagResponse struct {
	Results []Tag `json:"results"`
}
type Tag struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	FullSize int64  `json:"full_size"`
}

func checkForNewerVersion(cname string) {
	ctx := context.Background()
	time.Sleep(3 * time.Second)

	// Grab latest tags
	resp, err := http.Get("https://registry.hub.docker.com/v2/repositories/fnproject/fnserver/tags")
	if err != nil {
		log.Println("Error getting tags: %s", err)
	}
	defer resp.Body.Close()
	tagR := &TagResponse{}
	err = json.NewDecoder(resp.Body).Decode(tagR)
	fmt.Printf("tagR: %+v\n", tagR)

	dc, err := client.NewEnvClient()
	if err != nil {
		log.Println("Error instantiating Docker client: %s", err)
		return
	}

	cJSON, err := dc.ContainerInspect(ctx, cname)
	if err != nil {
		log.Println("Error getting fnserver container info: %s", err)
		return
	}
	fmt.Printf("cJSON: %+v\n", cJSON)

	args := filters.NewArgs()
	args.Add("reference", "fnproject/fnserver")
	imgList, err := dc.ImageList(ctx, types.ImageListOptions{
		Filters: args,
	})
	if err != nil {
		log.Printf("Error getting image list: %s", err)
		return
	}
	for _, img := range imgList {
		fmt.Printf("image: %+v\n", img)
		if len(img.RepoTags) > 0 {
			if img.RepoTags[0] == "fnproject/fnserver:latest" {

			}
		}
	}

	// newImageInfo, _, err := dc.ImageInspectWithRaw(ctx, functionsDockerImage)
	// if err != nil {
	// 	log.Println("Error inspecting image: %s", err)
	// 	return
	// }

	// if newImageInfo.ID != cJSON.Image {
	// 	log.Println("Found new fnserver image, try updating with `fn update` to get latest version.")
	// 	return
	// }
}
