package common

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"encoding/base32"
	"os/signal"
	"path/filepath"
	"crypto/rand"
	"strings"
	"time"
	"unicode"

	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"

	"github.com/coreos/go-semver/semver"
	"github.com/fatih/color"
	"github.com/fnproject/cli/config"
	"github.com/fnproject/cli/langs"
	"github.com/urfave/cli"
)

// Global docker variables.
const (
	FunctionsDockerImage     = "fnproject/fnserver"
	FuncfileDockerRuntime    = "docker"
	MinRequiredDockerVersion = "17.5.0"
	RequestID = "request-id"
	RequestedLength = 16
)

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

func GetRequestID() string {
	byteArr := make([]byte, RequestedLength)
	_, err := rand.Read(byteArr)
	if err != nil {
		log.Fatalf("failed to generate random number for requestID")
	}

	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(byteArr)
}

// BuildFunc bumps version and builds function.
func BuildFunc(c *cli.Context, fpath string, funcfile *FuncFile, buildArg []string, noCache bool) (*FuncFile, error) {
	var err error
	if funcfile.Version == "" {
		funcfile, err = BumpIt(fpath, Patch)
		if err != nil {
			return nil, err
		}
	}

	if err := localBuild(fpath, funcfile.Build); err != nil {
		return nil, err
	}

	if err := dockerBuild(c, fpath, funcfile, buildArg, noCache); err != nil {
		return nil, err
	}

	return funcfile, nil
}

// BuildFunc bumps version and builds function.
func BuildFuncV20180707(c *cli.Context, fpath string, funcfile *FuncFileV20180707, buildArg []string, noCache bool) (*FuncFileV20180707, error) {
	var err error

	if funcfile.Version == "" {
		funcfile, err = BumpItV20180707(fpath, Patch)
		if err != nil {
			return nil, err
		}
	}

	if err := localBuild(fpath, funcfile.Build); err != nil {
		return nil, err
	}

	if err := dockerBuildV20180707(c, fpath, funcfile, buildArg, noCache); err != nil {
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

func dockerBuild(c *cli.Context, fpath string, ff *FuncFile, buildArgs []string, noCache bool) error {
	err := dockerVersionCheck()
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
		defer os.Remove(dockerfile)
		if helper.HasPreBuild() {
			err := helper.PreBuild()
			if err != nil {
				return err
			}
		}
	}
	err = RunBuild(c, dir, ff.ImageName(), dockerfile, buildArgs, noCache)
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

func dockerBuildV20180707(c *cli.Context, fpath string, ff *FuncFileV20180707, buildArgs []string, noCache bool) error {
	err := dockerVersionCheck()
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
		dockerfile, err = writeTmpDockerfileV20180707(helper, dir, ff)
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
	err = RunBuild(c, dir, ff.ImageNameV20180707(), dockerfile, buildArgs, noCache)
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

// RunBuild runs function from func.yaml/json/yml.
func RunBuild(c *cli.Context, dir, imageName, dockerfile string, buildArgs []string, noCache bool) error {
	cancel := make(chan os.Signal, 3)
	signal.Notify(cancel, os.Interrupt) // and others perhaps
	defer signal.Stop(cancel)

	result := make(chan error, 1)

	buildOut := ioutil.Discard
	buildErr := ioutil.Discard

	quit := make(chan struct{})
	fmt.Fprintf(os.Stderr, "Building image %v ", imageName)
	if c.GlobalBool("verbose") {
		fmt.Println()
		buildOut = os.Stdout
		buildErr = os.Stderr
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
		args := []string{
			"build",
			"-t", imageName,
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
		cmd := exec.Command("docker", args...)
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
			fmt.Printf("%v Run with `--verbose` flag to see what went wrong. eg: `fn --verbose CMD`\n", color.RedString("Error during build."))
			return fmt.Errorf("error running docker build: %v", err)
		}
	case signal := <-cancel:
		close(quit)
		fmt.Fprintln(os.Stderr)
		return fmt.Errorf("build cancelled on signal %v", signal)
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
	vMin, err := semver.NewVersion(MinRequiredDockerVersion)
	if err != nil {
		return fmt.Errorf("our bad, sorry... please make an issue, detailed error: %v", err)
	}
	if v.LessThan(*vMin) {
		return fmt.Errorf("please upgrade your version of Docker to %s or greater", MinRequiredDockerVersion)
	}
	return nil
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

func writeTmpDockerfileV20180707(helper langs.LangHelper, dir string, ff *FuncFileV20180707) (string, error) {
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

// ExtractEnvConfig extract env vars from configuration.
func ExtractEnvConfig(configs []string) map[string]string {
	c := make(map[string]string)
	for _, v := range configs {
		kv := strings.SplitN(v, "=", 2)
		if len(kv) == 2 {
			c[kv[0]] = os.ExpandEnv(kv[1])
		}
	}
	return c
}

// DockerPush pushes to docker registry.
func DockerPush(ff *FuncFile) error {
	_, err := ValidateImageName(ff.ImageName())
	if err != nil {
		return err
	}
	fmt.Printf("Pushing %v to docker registry...", ff.ImageName())
	cmd := exec.Command("docker", "push", ff.ImageName())
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running docker push, are you logged into docker?: %v", err)
	}
	return nil
}

// DockerPush pushes to docker registry.
func DockerPushV20180707(ff *FuncFileV20180707) error {
	_, err := ValidateImageName(ff.ImageNameV20180707())
	if err != nil {
		return err
	}
	fmt.Printf("Pushing %v to docker registry...", ff.ImageNameV20180707())
	cmd := exec.Command("docker", "push", ff.ImageNameV20180707())
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running docker push, are you logged into docker?: %v", err)
	}
	return nil
}

// ValidateImageName validates that the full image name (REGISTRY/name:tag) is allowed for push
// remember that private registries must be supported here
func ValidateImageName(n string) (string, error) {
	parts := strings.Split(n, "/")
	if len(parts) < 2 {
		if viper.GetString("registry") == "" {
			return "", errors.New("image name must have a dockerhub owner or private registry. Be sure to set FN_REGISTRY env var, pass in --registry or configure your context file")
		}
		n = viper.GetString("registry") + "/" + n
	}
	lastParts := strings.Split(parts[len(parts)-1], ":")
	if len(lastParts) != 2 {
		return "", errors.New("image name must have a tag")
	}
	return n, nil
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
