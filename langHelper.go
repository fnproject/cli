package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fnproject/cli/langhelper"
)

func LangHelper(rt string, flags []string) (*langhelper.LangHelpResults, error) {
	pwd, _ := os.Getwd()
	lh := &langhelper.LangHelpResults{}
	langHelperImage := getLangHelperImageName(rt)
	//TODO: cmdFlag needs changing; very stupid solution at present
	//Was done this way because it was the quickest way of using two commands in one docker image
	commands := []string{"run", "-i", "--mount", "type=bind,src=" + pwd + ",target=/dummy", langHelperImage, "-runtime=" + rt}
	for _, val := range flags {
		commands = append(commands, val)
	}
	cmd := exec.Command("docker", commands...)
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

func getLangHelperImageName(rt string) string {
	if strings.Contains(rt, "/") {
		return rt
	}
	return fmt.Sprintf("fnproject/lang-%s:latest", rt)
}
