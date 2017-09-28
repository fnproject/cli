package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type LangHelp struct {
	LangBuilder
	LangInitialiser
}

func LangHelper(rt, cmdFlag string) (*LangHelp, error) {
	pwd, _ := os.Getwd()
	lh := &LangHelp{}
	langHelperImage := getLangHelperImageName(rt)
	//TODO: cmdFlag needs changing; very stupid solution at present
	//Was done this way because it was the quickest way of using two commands in one docker image
	cmd := exec.Command("docker", "run", "-i", "--mount", "type=bind,src="+pwd+",target=/dummy", langHelperImage, cmdFlag)
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
