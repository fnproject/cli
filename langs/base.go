package langs

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

//used to indicate the default supported version of java
const defaultJavaSupportedVersion = "9"

var (
	ErrBoilerplateExists = errors.New("Function boilerplate already exists")
)

type LangHelper struct {
	//NB: will be deconstructed and moved to common.go/init.go respectively

	BuildImage          string
	RunImage            string
	IsMultiStage        bool
	DockerfileBuildCmds []string
	DockerfileCopyCmds  []string
	Entrypoint          string
	Cmd                 string
}

func InitLangHelper(rt string) (*LangHelper, error) {
	//NB: will be moved to init.go eventually

	//download image

	//run container's "init" command - this will create boilerplate

	lh := &LangHelper{}
	//retrieve lh.Cmd, lh.Entrypoint
	pwd, _ := os.Getwd()
	//return lh, nil
	cmd := exec.Command("docker", "run", "-i", "--mount", "type=bind,src="+pwd+",target=/dummy", "ollerhll/dummy", "-init")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err = cmd.Start(); err != nil {
		return nil, err
	}

	if err = json.NewDecoder(stdout).Decode(lh); err != nil {
		return nil, err
	}

	if err = cmd.Wait(); err != nil {
		return nil, err
	}

	return lh, nil
}

func BuildLangHelper(rt string) (*LangHelper, error) {
	//NB: essentially a rewrite of dockerBuild in common.go, so will be moved there eventually

	//download image

	//run container's "build" command
	//	var lh LangHelper

	//retrieve lh.BuildImage, lh.RunImage, lh.IsMultiStage, DockerfileBuildCmds, DockerFileCopyCmds
	//if necessary, build tmp dockerfile (using code from common.go)

	//build the image

	//run container's "afterbuild" command
	pwd, _ := os.Getwd()
	lh := &LangHelper{}

	cmd := exec.Command("docker", "run", "-i", "--mount", "type=bind,src="+pwd+",target=/dummy", "ollerhll/dummy")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err = cmd.Start(); err != nil {
		return nil, err
	}

	if err = json.NewDecoder(stdout).Decode(lh); err != nil {
		return nil, err
	}

	if err = cmd.Wait(); err != nil {
		return nil, err
	}

	return lh, nil
}

// exists checks if a file exists
func exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func dockerBuildError(err error) error {
	return fmt.Errorf("error running docker build: %v", err)
}
