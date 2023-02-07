/*
 * Copyright (c) 2019, 2020 Oracle and/or its affiliates. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package common

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/coreos/go-semver/semver"
	"github.com/fatih/color"
	"github.com/fnproject/cli/config"
	"github.com/fnproject/cli/langs"
	fnclient "github.com/fnproject/fn_go/clientv2"
	apifns "github.com/fnproject/fn_go/clientv2/fns"
	apitriggers "github.com/fnproject/fn_go/clientv2/triggers"
	"github.com/fnproject/fn_go/modelsv2"
	"github.com/spf13/viper"
	"github.com/urfave/cli"
)

// Global docker variables.
const (
	//FunctionsDockerImage     = "fnproject/fnserver"
	FunctionsDockerImage     = "greendragons/testfnproject"
	FuncfileDockerRuntime    = "docker"
	MinRequiredDockerVersion = "17.5.0"
	BuildxBuilderInstance    = "oci_fn_builder"
)

var GlobalVerbose bool
var CommandVerbose bool

var architectureMap = map[string]string{
	"x86": "linux/amd64",
	"arm": "linux/arm64",
}

func IsVerbose() bool {
	return GlobalVerbose || CommandVerbose
}

// DefaultBashComplete prints the list of all sub commands
// of the current command (without alias names)
func DefaultBashComplete(c *cli.Context) {
	for _, command := range c.App.Commands {
		if command.Hidden {
			continue
		}
		if command.Name != "help" {
			fmt.Println(command.Name)
		}
	}
}

// GetWd returns working directory.
func GetWd() string {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln("Couldn't get working directory:", err)
	}
	return wd
}

// GetDir returns the dir if defined as a flag in cli.Context
func GetDir(c *cli.Context) string {
	var dir string
	if c.String("working-dir") != "" {
		dir = c.String("working-dir")
	} else {
		dir = GetWd()
	}

	return dir
}

// BuildFunc bumps version and builds function.
func BuildFunc(verbose bool, fpath string, funcfile *FuncFile, buildArg []string, noCache bool) (*FuncFile, error) {
	var err error
	if funcfile.Version == "" {
		funcfile, err = BumpIt(fpath, Patch)
		if err != nil {
			return nil, err
		}
	}

	funcfile, err = imageStampFuncFile(fpath, funcfile)
	if err != nil {
		return nil, err
	}
	if err := localBuild(fpath, funcfile.Build); err != nil {
		return nil, err
	}

	if err := containerEngineBuild(verbose, fpath, funcfile, buildArg, noCache); err != nil {
		return nil, err
	}

	return funcfile, nil
}

// BuildFunc bumps version and builds function.
func BuildFuncV20180708(verbose bool, fpath string, funcfile *FuncFileV20180708, buildArg []string, noCache bool, architectures []string) (*FuncFileV20180708, error) {
	var err error

	if funcfile.Version == "" {
		funcfile, err = BumpItV20180708(fpath, Patch)
		if err != nil {
			return nil, err
		}
	}

	funcfile, err = imageStampFuncFileV20180708(fpath, funcfile)
	if err != nil {
		return nil, err
	}

	if err := localBuild(fpath, funcfile.Build); err != nil {
		return nil, err
	}
	if err := containerEngineBuildV20180708(verbose, fpath, funcfile, buildArg, noCache, architectures); err != nil {
		return nil, err
	}

	return funcfile, nil
}

func imageStampFuncFile(fpath string, funcfile *FuncFile) (*FuncFile, error) {

	dir := filepath.Dir(fpath)
	dockerfile := filepath.Join(dir, "Dockerfile")

	// detect if build and run image both are absent and runtime is not docker then update them
	if !Exists(dockerfile) && funcfile.Runtime != FuncfileDockerRuntime && funcfile.BuildImage == "" && funcfile.RunImage == "" {

		helper := langs.GetLangHelper(funcfile.Runtime)

		// For case when runtime is overriden with unsupported lang runtime in func.yaml manually
		if helper == nil {
			return funcfile, fmt.Errorf("runtime [%s] is not supported", funcfile.Runtime)
		}

		// get the lang without any version
		langRuntime := helper.Runtime()

		// if fallback is not supported then continue with same lang helper
		if langs.IsFallbackSupported(langRuntime) {
			helper = langs.GetFallbackLangHelper(langRuntime)
		}

		// update func file accordingly
		bi, err := helper.BuildFromImage()
		if err != nil {
			return funcfile, err
		}
		funcfile.BuildImage = bi

		// In case of multistage build there will be a runtime image
		if helper.IsMultiStage() {
			ri, err := helper.RunFromImage()
			if err != nil {
				return funcfile, err
			}
			funcfile.RunImage = ri
		}

		// fill back yaml file
		err = EncodeFuncfileYAML(fpath, funcfile)
		if err != nil {
			return funcfile, err
		}
	}

	return funcfile, nil
}

// Stamping funcfile is only valid for functions with runtime lang not for docker runtime
func imageStampFuncFileV20180708(fpath string, funcfile *FuncFileV20180708) (*FuncFileV20180708, error) {

	dir := filepath.Dir(fpath)
	dockerfile := filepath.Join(dir, "Dockerfile")

	// detect if build and run image both are absent and runtime is not docker then update them
	if !Exists(dockerfile) && funcfile.Runtime != FuncfileDockerRuntime && funcfile.Build_image == "" && funcfile.Run_image == "" {

		helper := langs.GetLangHelper(funcfile.Runtime)

		// For case when runtime is overriden with unsupported lang runtime in func.yaml manually
		if helper == nil {
			return funcfile, fmt.Errorf("runtime [%s] is not supported", funcfile.Runtime)
		}

		// get the lang without any version
		langRuntime := helper.Runtime()

		// if fallback is not supported then continue with same lang helper
		if langs.IsFallbackSupported(langRuntime) {
			helper = langs.GetFallbackLangHelper(langRuntime)
		}

		bi, err := helper.BuildFromImage()
		if err != nil {
			return funcfile, err
		}
		funcfile.Build_image = bi

		// In case of multistage build there will be a runtime image
		if helper.IsMultiStage() {
			ri, err := helper.RunFromImage()
			if err != nil {
				return funcfile, nil
			}
			funcfile.Run_image = ri
		}

		// fill back yaml file
		err = EncodeFuncFileV20180708YAML(fpath, funcfile)
		if err != nil {
			return funcfile, err
		}
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

func PrintContextualInfo() {
	var registry, currentContext string
	registry = viper.GetString(config.EnvFnRegistry)
	if registry == "" {
		registry = "FN_REGISTRY is not set."
	}
	fmt.Println("FN_REGISTRY: ", registry)

	currentContext = viper.GetString(config.CurrentContext)
	if currentContext == "" {
		currentContext = "No context currently in use."
	}
	fmt.Println("Current Context: ", currentContext)
}

func PrintDockerfileContent(dockerfile string, stdout io.Writer) {
	file, err := os.Open(dockerfile)
	if err != nil {
		fmt.Fprintf(stdout, err.Error())
		fmt.Fprintf(stdout, "\n")
		return
	}
	defer file.Close()
	fmt.Fprintf(stdout, "Dockerfile content\n-----------------------------------\n")
	_, _ = io.Copy(stdout, file)
	fmt.Fprintf(stdout, "-----------------------------------\n")
}

func GetContainerEngineType() (string, error) {
	containerEngineType := viper.GetString(config.ContainerEngineType)
	err := config.ValidateContainerEngineType(containerEngineType)
	if err != nil {
		return "", err
	}
	if containerEngineType == "" {
		return "docker", nil
	}
	return containerEngineType, nil
}

func containerEngineBuild(verbose bool, fpath string, ff *FuncFile, buildArgs []string, noCache bool) error {
	containerEngineType, err := GetContainerEngineType()
	if err != nil {
		return err
	}

	fmt.Println("Using Container engine", containerEngineType)
	err = containerEngineVersionCheck(containerEngineType)
	if err != nil {
		return err
	}

	dir := filepath.Dir(fpath)

	var helper langs.LangHelper
	dockerfile := filepath.Join(dir, "Dockerfile")

	if !Exists(dockerfile) {
		if ff.Runtime == FuncfileDockerRuntime {
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
		//defer os.Remove(dockerfile)
		if helper.HasPreBuild() {
			err := helper.PreBuild()
			if err != nil {
				return err
			}
		}
	}
	err = RunBuild(verbose, dir, ff.ImageName(), dockerfile, buildArgs, noCache, containerEngineType, make([]string, 0))
	if err != nil {
		return err
	}

	if helper != nil {
		err := helper.AfterBuild()
		if err != nil {
			return err
		}
	}
	return nil
}

func containerEngineBuildV20180708(verbose bool, fpath string, ff *FuncFileV20180708, buildArgs []string, noCache bool, architectures []string) error {
	containerEngineType, err := GetContainerEngineType()
	if err != nil {
		return err
	}

	fmt.Println("Using Container engine", containerEngineType)
	err = containerEngineVersionCheck(containerEngineType)
	if err != nil {
		return err
	}

	dir := filepath.Dir(fpath)

	var helper langs.LangHelper
	dockerfile := filepath.Join(dir, "Dockerfile")
	if !Exists(dockerfile) {
		if ff.Runtime == FuncfileDockerRuntime {
			return fmt.Errorf("Dockerfile does not exist for 'docker' runtime")
		}

		helper = langs.GetLangHelper(ff.Runtime)
		if helper == nil {
			return fmt.Errorf("Cannot build, no language helper found for %v", ff.Runtime)
		}
		dockerfile, err = writeTmpDockerfileV20180708(helper, dir, ff)
		if err != nil {
			return err
		}
		//defer os.Remove(dockerfile)
		if helper.HasPreBuild() {
			err := helper.PreBuild()
			if err != nil {
				return err
			}
		}
	}

	err = RunBuild(verbose, dir, ff.ImageNameV20180708(), dockerfile, buildArgs, noCache, containerEngineType, architectures)
	if err != nil {
		return err
	}

	if helper != nil {
		err := helper.AfterBuild()
		if err != nil {
			return err
		}
	}
	return nil
}

func buildXDockerCommand(imageName, dockerfile string, buildArgs []string, noCache bool, architectures []string) []string {
	var buildCommand = "buildx"
	var name = imageName

	args := []string{
		buildCommand,
		"build",
		"-t", name,
		"-f", dockerfile,
		"--platform", strings.Join(architectures, ","),
	}

	if noCache {
		args = append(args, "--no-cache")
	}

	if len(buildArgs) > 0 {
		for _, buildArg := range buildArgs {
			args = append(args, "--build-arg", buildArg)
		}
	}

	if len(architectures) > 0 {
		var arg = "ARCH=" + strings.Join(architectures, ",")
		var label = "imageName=" + imageName
		args = append(args, "--build-arg", arg)
		args = append(args, "--label", label)
		args = append(args, "--push")
	}

	args = append(args,
		"--build-arg", "HTTP_PROXY",
		"--build-arg", "HTTPS_PROXY",
		".")

	// Container engine type would be optional here
	cmd := strings.Join(args, " ")
	fmt.Println(cmd)

	return args
}

func buildDockerCommand(imageName, dockerfile string, buildArgs []string, noCache bool) []string {
	var name = imageName
	args := []string{
		"build",
		"-t", name,
		"-f", dockerfile,
	}

	if noCache {
		args = append(args, "--no-cache")
	}

	if len(buildArgs) > 0 {
		for _, buildArg := range buildArgs {
			args = append(args, "--build-arg", buildArg)
		}
	}

	args = append(args,
		"--build-arg", "HTTP_PROXY",
		"--build-arg", "HTTPS_PROXY",
		".")

	// Container engine type would be optional here
	return args
}

// RunBuild runs function from func.yaml/json/yml.
func RunBuild(verbose bool, dir, imageName, dockerfile string, buildArgs []string, noCache bool, containerEngineType string, architectures []string) error {
	cancel := make(chan os.Signal, 3)
	signal.Notify(cancel, os.Interrupt) // and others perhaps
	defer signal.Stop(cancel)

	result := make(chan error, 1)

	buildOut := ioutil.Discard
	buildErr := ioutil.Discard

	quit := make(chan struct{})
	fmt.Fprintf(os.Stderr, "Building image %v ", imageName)
	if verbose {
		fmt.Println()
		buildOut = os.Stdout
		buildErr = os.Stderr
		PrintDockerfileContent(dockerfile, buildOut)
		PrintContextualInfo()
	} else {
		// print dots. quit channel explanation: https://stackoverflow.com/a/16466581/105562
		ticker := time.NewTicker(1 * time.Second)
		go func() {
			for {
				select {
				case <-ticker.C:
					fmt.Fprintf(os.Stderr, ".")
				case <-quit:
					ticker.Stop()
					return
				}
			}
		}()
	}

	go func(done chan<- error) {
		var dockerBuildCmdArgs []string
		// Depending whether architecture list is passed or not trigger docker buildx or docker build accordingly
		if architectures != nil && len(architectures) != 0 {

			var mappedArchitectures []string
			for _, arch := range architectures {
				mappedArch, ok := architectureMap[arch]
				if !ok {
					err := errors.New(fmt.Sprintf("invalid architecture type : %v", arch))
					done <- err
					return
				}
				mappedArchitectures = append(mappedArchitectures, mappedArch)
			}

			err := initializeContainerBuilder(containerEngineType, mappedArchitectures)
			if err != nil {
				done <- err
				return
			}

			dockerBuildCmdArgs = buildXDockerCommand(imageName, dockerfile,  buildArgs, noCache, mappedArchitectures)
			// perform cleanup
			defer cleanupContainerBuilder(containerEngineType)
		} else {
			dockerBuildCmdArgs = buildDockerCommand(imageName, dockerfile,  buildArgs, noCache)
		}

		cmd := exec.Command(containerEngineType, dockerBuildCmdArgs...)
		cmd.Dir = dir
		cmd.Stderr = buildErr // Doesn't look like there's any output to stderr on docker build, whether it's successful or not.
		cmd.Stdout = buildOut
		done <- cmd.Run()
	}(result)

	select {
	case err := <-result:
		close(quit)
		fmt.Fprintln(os.Stderr)
		if err != nil {
			if verbose == false {
				fmt.Printf("%v Run with `--verbose` flag to see what went wrong. eg: `fn --verbose CMD`\n", color.RedString("Error during build."))
			}
			return fmt.Errorf("error running docker build: %v", err)
		}
	case signal := <-cancel:
		close(quit)
		fmt.Fprintln(os.Stderr)
		return fmt.Errorf("build cancelled on signal %v", signal)
	}
	return nil
}

func containerEngineVersionCheck(containerEngineType string) error {
	out, err := exec.Command(containerEngineType, "version", "--format", "{{.Server.Version}}").Output()
	if err != nil {
		return fmt.Errorf("Cannot connect to the %v, make sure you have it installed and running: %v", containerEngineType, err)
	}
	// dev / test builds append '-ce', trim this
	trimmed := strings.TrimRightFunc(string(out), func(r rune) bool { return r != '.' && !unicode.IsDigit(r) })

	v, err := semver.NewVersion(trimmed)
	if err != nil {
		return fmt.Errorf("could not check %v version: %v", containerEngineType, err)
	}
	if containerEngineType == "docker" {
		vMin, err := semver.NewVersion(MinRequiredDockerVersion)
		if err != nil {
			return fmt.Errorf("our bad, sorry... please make an issue, detailed error: %v", err)
		}
		if v.LessThan(*vMin) {
			return fmt.Errorf("please upgrade your version of Docker to %s or greater", MinRequiredDockerVersion)
		}
	}
	return nil
}

func initializeContainerBuilder(containerEngineType string, platforms []string) error {
	// ignoring the error as there could be no such existing builder
	// valid error would be caught later while creating a builder
	_ = cleanupContainerBuilder(containerEngineType)

	var args []string
	args = append(args, "buildx")
	args = append(args, "create")
	args = append(args, "--name", BuildxBuilderInstance)
	args = append(args, "--use")
	args = append(args, "--platform", strings.Join(platforms, ","))

	buildKitVersion := os.Getenv("BUILDKIT_VERSION")
	if len(buildKitVersion) != 0 {
		args = append(args, "--driver-opt image=moby/buildkit:" + buildKitVersion)
	}

	_, err := exec.Command(containerEngineType, args...).Output()
	if err != nil {
		return fmt.Errorf("Cannot create/use builder instance %v for %v : %v", containerEngineType,
			BuildxBuilderInstance, err)
	}

	return nil
}

func cleanupContainerBuilder(containerEngineType string) error {
	var args []string
	args = append(args, "buildx")
	args = append(args, "rm")
	args = append(args, BuildxBuilderInstance)

	//remove existing builder instance
	_, err := exec.Command(containerEngineType, args...).Output()
	return err
}

// Exists check file exists.
func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func writeTmpDockerfile(helper langs.LangHelper, dir string, ff *FuncFile) (string, error) {
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

	/* This check will always be false for non docker runtime
	   as all funcfiles are already image stamped
	   can we remove it in future.
	*/
	if bi == "" {
		bi, err = helper.BuildFromImage()
		if err != nil {
			return "", err
		}
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

		/*
		  Now this check will always be false for non docker runtime
		  as all funcfiles are already image stamped
		  can we remove it in future.
		*/
		if ri == "" {
			ri, err = helper.RunFromImage()
			if err != nil {
				return "", err
			}
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

func writeTmpDockerfileV20180708(helper langs.LangHelper, dir string, ff *FuncFileV20180708) (string, error) {
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
	bi := ff.Build_image

	if bi == "" {
		bi, err = helper.BuildFromImage()
		if err != nil {
			return "", err
		}
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
		ri := ff.Run_image
		if ri == "" {
			ri, err = helper.RunFromImage()
			if err != nil {
				return "", err
			}
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

// ExtractConfig parses key-value configuration into a map
func ExtractConfig(configs []string) map[string]string {
	c := make(map[string]string)
	for _, v := range configs {
		kv := strings.SplitN(v, "=", 2)
		if len(kv) == 2 {
			c[kv[0]] = kv[1]
		}
	}
	return c
}

// Push pushes to docker registry.
func Push(ff *FuncFile) error {
	containerEngineType, err := GetContainerEngineType()
	if err != nil {
		return err
	}
	err = ValidateFullImageName(ff.ImageName())
	if err != nil {
		return err
	}
	fmt.Printf("Pushing %v to docker registry...", ff.ImageName())
	cmd := exec.Command(containerEngineType, "push", ff.ImageName())
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running %v push, are you logged?: %v", containerEngineType, err)
	}
	return nil
}

// Push pushes to docker registry.
func PushV20180708(ff *FuncFileV20180708) error {
	containerEngineType, err := GetContainerEngineType()
	if err != nil {
		return err
	}
	err = ValidateFullImageName(ff.ImageNameV20180708())
	if err != nil {
		return err
	}
	fmt.Println("Using Container engine", containerEngineType, "to push")
	fmt.Printf("Pushing %v to docker registry...", ff.ImageNameV20180708())
	cmd := exec.Command(containerEngineType, "push", ff.ImageNameV20180708())
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running %v push, are you logged?: %v", containerEngineType, err)
	}
	return nil
}

// ValidateFullImageName validates that the full image name (REGISTRY/name:tag) is allowed for push
// remember that private registries must be supported here
func ValidateFullImageName(n string) error {
	parts := strings.Split(n, "/")
	fmt.Println("Parts: ", parts)
	if len(parts) < 2 {
		return errors.New("image name must have a dockerhub owner or private registry. Be sure to set FN_REGISTRY env var, pass in --registry or configure your context file")

	}
	return ValidateTagImageName(n)
}

// ValidateTagImageName validates that the last part of the image name (name:tag) is allowed for create/update
func ValidateTagImageName(n string) error {
	parts := strings.Split(n, "/")
	lastParts := strings.Split(parts[len(parts)-1], ":")
	if len(lastParts) != 2 {
		return errors.New("image name must have a tag")
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

// ExtractAnnotations extract annotations from command flags.
func ExtractAnnotations(c *cli.Context) map[string]interface{} {
	annotations := make(map[string]interface{})
	for _, s := range c.StringSlice("annotation") {
		parts := strings.Split(s, "=")
		if len(parts) == 2 {
			var v interface{}
			err := json.Unmarshal([]byte(parts[1]), &v)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to parse annotation value '%v'. Annotations values must be valid JSON strings.\n", parts[1])
			} else {
				annotations[parts[0]] = v
			}
		} else {
			fmt.Fprintf(os.Stderr, "Annotations must be specified in the form key='value', where value is a valid JSON string")
		}
	}
	return annotations
}

func ExtractArchitecturesTypeEnum(architectureString string) ([]modelsv2.ArchitectureType, error) {
	var architectures []modelsv2.ArchitectureType
	err := json.Unmarshal([]byte(architectureString), &architectures)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to parse architectures value '%v'. Architectures values must be a valid JSON string.\n", architectureString))
	}

	architectureMap := make(map[modelsv2.ArchitectureType]bool, 0)
	for _, arch := range architectures {
		//verify the string value for architecture type is a valid value
		archType := modelsv2.ArchitectureType(arch)
		if err := archType.Validate(nil); err != nil {
			return nil, errors.New(fmt.Sprintf("invalid architecture type : %s", arch))
		}

		if _, ok := architectureMap[archType]; ok {
			return nil, errors.New(fmt.Sprintf("duplicate architecture type found : %s", arch))
		}
		architectureMap[archType] = true
	}
	return  architectures, nil
}

func ExtractArchitecturesType(architectureString string) ([]string, error) {
	var architectures []string
	err := json.Unmarshal([]byte(architectureString), &architectures)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to parse architectures value '%v'. Architectures values must be a valid JSON string.\n", architectureString))
	}

	architectureMap := make(map[string]bool, 0)
	for _, arch := range architectures {
		//verify the string value for architecture type is a valid value
		archType := modelsv2.ArchitectureType(arch)
		if err := archType.Validate(nil); err != nil {
			return nil, errors.New(fmt.Sprintf("invalid architecture type : %s", arch))
		}

		if _, ok := architectureMap[arch]; ok {
			return nil, errors.New(fmt.Sprintf("duplicate architecture type found : %s", arch))
		}
		architectureMap[arch] = true
	}
	return  architectures, nil
}

func ReadInFuncFile() (map[string]interface{}, error) {
	wd := GetWd()

	fpath, err := FindFuncfile(wd)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadFile(fpath)
	if err != nil {
		return nil, fmt.Errorf("could not open %s for parsing. Error: %v", fpath, err)
	}

	var ff map[string]interface{}
	err = yaml.Unmarshal(b, &ff)
	if err != nil {
		return nil, err
	}

	return ff, nil
}

func GetFuncYamlVersion(oldFF map[string]interface{}) int {
	if _, ok := oldFF["schema_version"]; ok {
		return oldFF["schema_version"].(int)
	}
	return 1
}

// UserConfirmedMultiResourceDeletion will prompt the user for confirmation to delete all the the resources
func UserConfirmedMultiResourceDeletion(aps []*modelsv2.App, fns []*modelsv2.Fn, trs []*modelsv2.Trigger) bool {

	apsLen := len(aps)
	fnsLen := len(fns)
	trsLen := len(trs)

	fmt.Println("You are about to delete the following resources:")
	if apsLen > 0 {
		fmt.Println("   Applications:", apsLen)
	}
	if fnsLen > 0 {
		fmt.Println("   Functions:   ", fnsLen)
	}
	if trsLen > 0 {
		fmt.Println("   Triggers:    ", trsLen)
	}
	fmt.Println("This operation cannot be undone")
	fmt.Printf("Do you wish to proceed? Y/N: ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if strings.ToLower(input) == "y" {
		return true
	} else if strings.ToLower(input) == "n" {
		fmt.Println("Cancelling delete")
		return false
	} else {
		fmt.Println("Unrecognised input, should be Y/N")
		fmt.Println("Cancelling delete")
		return false
	}
	return true
}

// ListFnsAndTriggersInApp lists all the functions associated with an app and all the triggers associated with each of those functions
func ListFnsAndTriggersInApp(c *cli.Context, client *fnclient.Fn, app *modelsv2.App) ([]*modelsv2.Fn, []*modelsv2.Trigger, error) {
	fns, err := ListFnsInApp(c, client, app)
	if err != nil {
		return nil, nil, err
	}
	var trs []*modelsv2.Trigger
	for _, fn := range fns {
		t, err := ListTriggersInFunc(c, client, fn)
		if err != nil {
			return nil, nil, err
		}
		trs = append(trs, t...)
	}
	return fns, trs, nil
}

//DeleteFunctions deletes all the functions provided to it. if provided nil it is a no-op
func DeleteFunctions(c *cli.Context, client *fnclient.Fn, fns []*modelsv2.Fn) error {
	if fns == nil {
		return nil
	}
	for _, fn := range fns {
		params := apifns.NewDeleteFnParams()
		params.FnID = fn.ID
		_, err := client.Fns.DeleteFn(params)
		if err != nil {
			return fmt.Errorf("Failed to delete Function %s: %s", fn.Name, err)
		}
		fmt.Println("Function ", fn.Name, " deleted")
	}
	return nil
}

//DeleteTriggers deletes all the triggers provided to it. if provided nil it is a no-op
func DeleteTriggers(c *cli.Context, client *fnclient.Fn, triggers []*modelsv2.Trigger) error {
	if triggers == nil {
		return nil
	}
	for _, t := range triggers {
		params := apitriggers.NewDeleteTriggerParams()
		params.TriggerID = t.ID
		_, err := client.Triggers.DeleteTrigger(params)
		if err != nil {
			return fmt.Errorf("Failed to Delete trigger %s: %s", t.Name, err)
		}
		fmt.Println("Trigger ", t.Name, " deleted")
	}
	return nil
}

//ListFnsInApp gets all the functions associated with an app
func ListFnsInApp(c *cli.Context, client *fnclient.Fn, app *modelsv2.App) ([]*modelsv2.Fn, error) {
	params := &apifns.ListFnsParams{
		Context: context.Background(),
		AppID:   &app.ID,
	}

	var resFns []*modelsv2.Fn
	for {
		resp, err := client.Fns.ListFns(params)

		if err != nil {
			return nil, fmt.Errorf("Could not list functions in application %s: %s", app.Name, err)
		}
		n := c.Int64("n")

		resFns = append(resFns, resp.Payload.Items...)
		howManyMore := n - int64(len(resFns)+len(resp.Payload.Items))
		if howManyMore <= 0 || resp.Payload.NextCursor == "" {
			break
		}

		params.Cursor = &resp.Payload.NextCursor
	}

	return resFns, nil
}

//ListTriggersInFunc gets all the triggers associated with a function
func ListTriggersInFunc(c *cli.Context, client *fnclient.Fn, fn *modelsv2.Fn) ([]*modelsv2.Trigger, error) {
	params := &apitriggers.ListTriggersParams{
		Context: context.Background(),
		AppID:   &fn.AppID,
		FnID:    &fn.ID,
	}

	var resTriggers []*modelsv2.Trigger
	for {
		resp, err := client.Triggers.ListTriggers(params)
		if err != nil {
			return nil, fmt.Errorf("Could not list triggers in function %s: %s", fn.Name, err)
		}
		n := c.Int64("n")

		resTriggers = append(resTriggers, resp.Payload.Items...)
		howManyMore := n - int64(len(resTriggers)+len(resp.Payload.Items))
		if howManyMore <= 0 || resp.Payload.NextCursor == "" {
			break
		}
		params.Cursor = &resp.Payload.NextCursor
	}
	return resTriggers, nil
}

func RunInitImage(initImage string, fName string) error {
	containerEngineType, err := GetContainerEngineType()
	if err != nil {
		return err
	}
	fmt.Println("Running init-image: " + initImage)

	args := []string{"run", "--rm", "-e", "FN_FUNCTION_NAME=" + fName}
	args = append(args, proxyArgs()...)
	args = append(args, initImage)

	fmt.Printf("Executing %v command: %s\n", containerEngineType, strings.Join(args, " "))
	cmd := exec.Command(containerEngineType, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("Error attaching stdout pipe to init-image cmd: %v.", err)
	}

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("Error starting init-image: %v. Stderr: '%s'", err, stderr.String())
	}

	err = untarStream(stdout)
	if err != nil {
		return fmt.Errorf("Error un-tarring stdout from the init-image: %v", err)
	}

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("Error running init-image: %v. Stderr: '%s'", err, stderr.String())
	}

	return nil
}

// Get any proxy environment variables
func proxyArgs() []string {
	args := make([]string, 0)
	for _, a := range []string{"http_proxy", "https_proxy", "ftp_proxy", "no_proxy", "HTTP_PROXY", "HTTPS_PROXY", "FTP_PROXY", "NO_PROXY"} {
		if val, ok := os.LookupEnv(a); ok {
			args = append(args, "-e", a+"="+val)
		}
	}
	return args
}

// Untars an io.Reader into the cwd
func untarStream(r io.Reader) error {
	tr := tar.NewReader(r)
	for {
		header, err := tr.Next()

		if err == io.EOF {
			// if no more files are found we are finished
			return nil
		}

		if err != nil {
			return err
		}

		switch header.Typeflag {
		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(header.Name); err != nil {
				if err := os.MkdirAll(header.Name, 0755); err != nil {
					return err
				}
			}

		// if it's a file create it
		case tar.TypeReg:
			f, err := os.OpenFile(header.Name, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			// copy over contents
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}

			_ = f.Close()
		}
	}
}
