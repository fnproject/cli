package common

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
	"time"
	"unicode"

	"github.com/coreos/go-semver/semver"
	"github.com/fatih/color"
	"github.com/fnproject/cli/langs"
	fnclient "github.com/fnproject/fn_go/client"
	"github.com/urfave/cli"
)

const (
	functionsDockerImage     = "fnproject/fnserver"
	funcfileDockerRuntime    = "docker"
	minRequiredDockerVersion = "17.5.0"
)

type FnClient struct {
	Client *fnclient.Fn
}

func getWd() string {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln("Couldn't get working directory:", err)
	}
	return wd
}

func buildfunc(c *cli.Context, fpath string, funcfile *FuncFile, buildArg []string, noCache bool) (*FuncFile, error) {
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

	if err := dockerBuild(c, fpath, funcfile, buildArg, noCache); err != nil {
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

func dockerBuild(c *cli.Context, fpath string, ff *FuncFile, buildArgs []string, noCache bool) error {
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
	err = runBuild(c, dir, ff.ImageName(), dockerfile, buildArgs, noCache)
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

func runBuild(c *cli.Context, dir, imageName, dockerfile string, buildArgs []string, noCache bool) error {
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
	vMin, err := semver.NewVersion(minRequiredDockerVersion)
	if err != nil {
		return fmt.Errorf("our bad, sorry... please make an issue, detailed error: %v", err)
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

func dockerPush(ff *FuncFile) error {
	err := validateImageName(ff.ImageName())
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

// validateImageName validates that the full image name (REGISTRY/name:tag) is allowed for push
// remember that private registries must be supported here
func validateImageName(n string) error {
	parts := strings.Split(n, "/")
	if len(parts) < 2 {
		return errors.New("image name must have a dockerhub owner or private registry. Be sure to set FN_REGISTRY env var, pass in --registry or configure your context file")
	}
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
