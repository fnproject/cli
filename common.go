package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/coreos/go-semver/semver"
	"github.com/fnproject/cli/langs"
)

const (
	functionsDockerImage     = "fnproject/fnserver"
	funcfileDockerRuntime    = "docker"
	minRequiredDockerVersion = "17.5.0"
	envFnRegistry            = "FN_REGISTRY"
)

type HasRegistry interface {
	Registry() string
}

func setRegistryEnv(hr HasRegistry) {
	if hr.Registry() != "" {
		err := os.Setenv(envFnRegistry, hr.Registry())
		if err != nil {
			log.Fatalf("Couldn't set %s env var: %v\n", envFnRegistry, err)
		}
	}
}

func getWd() string {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln("Couldn't get working directory:", err)
	}
	return wd
}

func buildfunc(fpath string, funcfile *funcfile, noCache bool) (*funcfile, error) {
	var err error
	if funcfile.Version == "" {
		funcfile, err = bumpIt(fpath, Patch)
		if err != nil {
			return nil, err
		}
	}

	if err := localBuild(fpath, funcfile.Build); err != nil {
		return nil, err
	}

	if err := dockerBuild(fpath, funcfile, noCache); err != nil {
		return nil, err
	}

	return funcfile, nil
}

func localBuild(path string, steps []string) error {
	for _, cmd := range steps {
		exe := exec.Command("/bin/sh", "-c", cmd)
		exe.Dir = filepath.Dir(path)
		if err := exe.Run(); err != nil {
			return fmt.Errorf("error running command %v (%v)", cmd, err)
		}
	}

	return nil
}

func dockerBuild(fpath string, ff *funcfile, noCache bool) error {
	err := dockerVersionCheck()
	if err != nil {
		return err
	}

	dir := filepath.Dir(fpath)

	var helper langs.LangHelper
	dockerfile := filepath.Join(dir, "Dockerfile")
	if !exists(dockerfile) {
		if ff.Runtime == funcfileDockerRuntime {
			return fmt.Errorf("Dockerfile does not exist for 'docker' runtime")
		}
		helper = langs.GetLangHelper(ff.Runtime)
		if helper == nil {
			return fmt.Errorf("Cannot build, no language helper found for %v", ff.Runtime)
		}
		dockerfile, err = writeTmpDockerfile(helper, dir, ff)
		if err != nil {
			return err
		}
		defer os.Remove(dockerfile)
		if helper.HasPreBuild() {
			err := helper.PreBuild()
			if err != nil {
				return err
			}
		}
	}

	fmt.Printf("Building image %v\n", ff.ImageName())

	cancel := make(chan os.Signal, 3)
	signal.Notify(cancel, os.Interrupt) // and others perhaps
	defer signal.Stop(cancel)

	result := make(chan error, 1)

	go func(done chan<- error) {
		args := []string{
			"build",
			"-t", ff.ImageName(),
			"-f", dockerfile,
		}
		if noCache {
			args = append(args, "--no-cache")
		}
		args = append(args,
			"--build-arg", "HTTP_PROXY",
			"--build-arg", "HTTPS_PROXY",
			".")
		cmd := exec.Command("docker", args...)
		cmd.Dir = dir
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		done <- cmd.Run()
	}(result)

	select {
	case err := <-result:
		if err != nil {
			return fmt.Errorf("error running docker build: %v", err)
		}
	case signal := <-cancel:
		return fmt.Errorf("build cancelled on signal %v", signal)
	}

	if helper != nil {
		err := helper.AfterBuild()
		if err != nil {
			return err
		}
	}
	return nil
}

func dockerVersionCheck() error {
	out, err := exec.Command("docker", "version", "--format", "{{.Server.Version}}").Output()
	if err != nil {
		return fmt.Errorf("could not check Docker version: %v", err)
	}
	// dev / test builds append '-ce', trim this
	trimmed := strings.TrimRightFunc(string(out), func(r rune) bool { return r != '.' && !unicode.IsDigit(r) })

	v, err := semver.NewVersion(trimmed)
	if err != nil {
		return fmt.Errorf("could not check Docker version: %v", err)
	}
	vMin, err := semver.NewVersion(minRequiredDockerVersion)
	if err != nil {
		return fmt.Errorf("our bad, sorry... please make an issue.", err)
	}
	if v.LessThan(*vMin) {
		return fmt.Errorf("please upgrade your version of Docker to %s or greater", minRequiredDockerVersion)
	}
	return nil
}

func exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func writeTmpDockerfile(helper langs.LangHelper, dir string, ff *funcfile) (string, error) {
	if ff.Entrypoint == "" && ff.Cmd == "" {
		return "", errors.New("entrypoint and cmd are missing, you must provide one or the other")
	}

	fd, err := ioutil.TempFile(dir, "Dockerfile")
	if err != nil {
		return "", err
	}
	defer fd.Close()

	// multi-stage build: https://medium.com/travis-on-docker/multi-stage-docker-builds-for-creating-tiny-go-images-e0e1867efe5a
	dfLines := []string{}
	bi := ff.BuildImage
	if bi == "" {
		bi = helper.BuildFromImage()
	}
	if helper.IsMultiStage() {
		// build stage
		dfLines = append(dfLines, fmt.Sprintf("FROM %s as build-stage", bi))
	} else {
		dfLines = append(dfLines, fmt.Sprintf("FROM %s", bi))
	}
	dfLines = append(dfLines, "WORKDIR /function")
	dfLines = append(dfLines, helper.DockerfileBuildCmds()...)
	if helper.IsMultiStage() {
		// final stage
		ri := ff.RunImage
		if ri == "" {
			ri = helper.RunFromImage()
		}
		dfLines = append(dfLines, fmt.Sprintf("FROM %s", ri))
		dfLines = append(dfLines, "WORKDIR /function")
		dfLines = append(dfLines, helper.DockerfileCopyCmds()...)
	}
	if ff.Entrypoint != "" {
		dfLines = append(dfLines, fmt.Sprintf("ENTRYPOINT [%s]", stringToSlice(ff.Entrypoint)))
	}
	if ff.Cmd != "" {
		dfLines = append(dfLines, fmt.Sprintf("CMD [%s]", stringToSlice(ff.Cmd)))
	}
	err = writeLines(fd, dfLines)
	if err != nil {
		return "", err
	}
	return fd.Name(), err
}

func writeLines(w io.Writer, lines []string) error {
	writer := bufio.NewWriter(w)
	for _, l := range lines {
		_, err := writer.WriteString(l + "\n")
		if err != nil {
			return err
		}
	}
	writer.Flush()
	return nil
}

func stringToSlice(in string) string {
	epvals := strings.Fields(in)
	var buffer bytes.Buffer
	for i, s := range epvals {
		if i > 0 {
			buffer.WriteString(", ")
		}
		buffer.WriteString("\"")
		buffer.WriteString(s)
		buffer.WriteString("\"")
	}
	return buffer.String()
}

func extractEnvConfig(configs []string) map[string]string {
	c := make(map[string]string)
	for _, v := range configs {
		kv := strings.SplitN(v, "=", 2)
		if len(kv) == 2 {
			c[kv[0]] = os.ExpandEnv(kv[1])
		}
	}
	return c
}

func dockerPush(ff *funcfile) error {
	err := validateImageName(ff.ImageName())
	if err != nil {
		return err
	}
	fmt.Printf("Pushing %v to docker registry...", ff.ImageName())
	cmd := exec.Command("docker", "push", ff.ImageName())
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running docker push: %v", err)
	}
	return nil
}

func validateImageName(n string) error {
	// must have at least owner name and a tag
	split := strings.Split(n, ":")
	if len(split) < 2 {
		return errors.New("image name must have a tag")
	}
	split2 := strings.Split(split[0], "/")
	if len(split2) < 2 {
		return errors.New("image name must have an owner and name, eg: username/myfunc. Be sure to set FN_REGISTRY env var or pass in --registry.")
	}
	return nil
}

func appNamePath(img string) (string, string) {
	sep := strings.Index(img, "/")
	if sep < 0 {
		return "", ""
	}
	tag := strings.Index(img[sep:], ":")
	if tag < 0 {
		tag = len(img[sep:])
	}
	return img[:sep], img[sep : sep+tag]
}
