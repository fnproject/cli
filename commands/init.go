package commands

/*
usage: fn init --help

o If there's a Dockerfile found, this will generate a basic
function file with the image and 'docker' as 'runtime'
like following, for example:

name: hello
version: 0.0.1
runtime: docker
path: /hello

then exit; if 'runtime' is 'docker' in the function file
and no Dockerfile exists, print an error message then exit
o It will then try to decipher the runtime based on
the files in the current directory, if it can't figure it out,
it will print an error message then exit.
*/

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fnproject/cli/common"
	"github.com/fnproject/cli/langs"
	function "github.com/fnproject/cli/objects/fn"
	"github.com/fnproject/cli/objects/route"
	models "github.com/fnproject/fn_go/models"
	modelsV2 "github.com/fnproject/fn_go/modelsv2"
	"github.com/urfave/cli"
)

type initFnCmd struct {
	force       bool
	triggerType string
	wd          string
	ff          *common.FuncFile
	ffV20180707 *common.FuncFileV20180707
}

func initFlags(a *initFnCmd) []cli.Flag {
	fgs := []cli.Flag{
		cli.StringFlag{
			Name:  "name",
			Usage: "Name of the function. Defaults to directory name in lowercase.",
		},
		cli.BoolFlag{
			Name:        "force",
			Usage:       "Overwrite existing func.yaml",
			Destination: &a.force,
		},
		cli.StringFlag{
			Name:  "runtime",
			Usage: "Choose an existing runtime - " + langsList(),
		},
		cli.StringFlag{
			Name:  "init-image",
			Usage: "A Docker image which will create a function template",
		},
		cli.StringFlag{
			Name:  "entrypoint",
			Usage: "Entrypoint is the command to run to start this function - equivalent to Dockerfile ENTRYPOINT.",
		},
		cli.StringFlag{
			Name:  "cmd",
			Usage: "Command to run to start this function - equivalent to Dockerfile CMD.",
		},
		cli.StringFlag{
			Name:  "version",
			Usage: "Set initial function version",
			Value: common.InitialVersion,
		},
		cli.StringFlag{
			Name:        "working-dir,w",
			Usage:       "Specify the working directory to initialise a function, must be the full path.",
			Destination: &a.wd,
		},
		cli.StringFlag{
			Name:        "trigger",
			Usage:       "Specify the trigger type.",
			Destination: &a.triggerType,
		},
	}

	return append(fgs, route.RouteFlags...)
}

func langsList() string {
	allLangs := []string{}
	for _, h := range langs.Helpers() {
		allLangs = append(allLangs, h.LangStrings()...)
	}
	return strings.Join(allLangs, ", ")
}

// InitCommand returns init cli.command
func InitCommand() cli.Command {
	a := &initFnCmd{ff: &common.FuncFile{}, ffV20180707: &common.FuncFileV20180707{}}

	return cli.Command{
		Name:        "init",
		Usage:       "\tCreate a local func.yaml file",
		Category:    "DEVELOPMENT COMMANDS",
		Aliases:     []string{"in"},
		Description: "This command creates a func.yaml file in the current directory.",
		ArgsUsage:   "[function-subdirectory]",
		Action:      a.init,
		Flags:       initFlags(a),
	}
}

func (a *initFnCmd) init(c *cli.Context) error {
	var err error
	var dir string

	dir = common.GetWd()
	if a.wd != "" {
		dir = a.wd
	}

	if a.triggerType != "" {
		var fn modelsV2.Fn
		function.FnWithFlags(c, &fn)
		a.bindFn(&fn)

		return a.initV2(c, fn)
	}

	var rt models.Route
	route.WithFlags(c, &rt)
	a.bindRoute(&rt)

	runtime := c.String("runtime")
	initImage := c.String("init-image")

	if runtime != "" && initImage != "" {
		return fmt.Errorf("You can't supply --runtime with --init-image")
	}

	runtimeSpecified := runtime != ""
	if runtimeSpecified {
		// go no further if the specified runtime is not supported
		if runtime != common.FuncfileDockerRuntime && langs.GetLangHelper(runtime) == nil {
			return fmt.Errorf("Init does not support the '%s' runtime", runtime)
		}
	}

	path := c.Args().First()
	if path != "" {
		fmt.Printf("Creating function at: /%s\n", path)
		dir = filepath.Join(dir, path)

		// check if dir exists, if it does, then we can't create function
		if common.Exists(dir) {
			if !a.force {
				return fmt.Errorf("directory %s already exists, cannot init function", dir)
			}
		} else {
			err = os.MkdirAll(dir, 0755)
			if err != nil {
				return err
			}
		}
	}

	err = os.Chdir(dir)
	if err != nil {
		return err
	}

	defer os.Chdir(dir) // todo: wrap this so we can log the error if changing back fails

	if !a.force {
		_, ff, err := common.LoadFuncfile(dir)
		if _, ok := err.(*common.NotFoundError); !ok && err != nil {
			return err
		}
		if ff != nil {
			return errors.New("Function file already exists, aborting")
		}
	}
	err = a.BuildFuncFile(c, dir) // TODO: Return LangHelper here, then don't need to refind the helper in generateBoilerplate() below
	if err != nil {
		return err
	}


	if (initImage != ""){

		fmt.Println("Building from init-image: " + initImage)

		// Run the initImage
		var c1ErrB, c2ErrB bytes.Buffer

		c1 := exec.Command("docker", "run", "-e", "FN_FUNCTION_NAME="+a.ff.Name, initImage)
		c1.Stderr = &c1ErrB

		c2 := exec.Command("tar", "-x")
		c2.Stderr = &c2ErrB
		c2.Stdin, _ = c1.StdoutPipe()
		c2.Stdout = os.Stdout

		_ = c2.Start()
		c1_err := c1.Run()
		c2_err := c2.Wait()

		if c1_err != nil {
			fmt.Println(c1ErrB.String())
			return errors.New("Error running init-image")
		}

		if c2_err != nil {
			fmt.Println(c2ErrB.String())
			return errors.New("Error un-tarring output from init-image")
		}

		// Merge the func.yaml from the initImage with a.ff
		//     write out the new func file
		var initFf, err = common.ParseFuncfile("func.init.yaml")
		if err != nil {
			return errors.New("init-image did not produce a valid func.yaml fragment")
		}

		initFf.Name = a.ff.Name
		initFf.Version = a.ff.Version

		if err := common.EncodeFuncfileYAML("func.yaml", initFf); err != nil {
			return err
		}

	} else {

		// TODO: why don't we treat "docker" runtime as just another language helper? Then can get rid of several Docker
		// specific if/else's like this one.
		if runtimeSpecified && runtime != common.FuncfileDockerRuntime {
			err := a.generateBoilerplate(dir, runtime)
			if err != nil {
				return err
			}
		}

		if err := common.EncodeFuncfileYAML("func.yaml", a.ff); err != nil {
			return err
		}

	}



	fmt.Println("func.yaml created.")
	return nil
}

func (a *initFnCmd) initV2(c *cli.Context, fn modelsV2.Fn) error {
	var err error
	var dir string

	dir = common.GetWd()
	if a.wd != "" {
		dir = a.wd
	}

	a.ffV20180707.Name = c.Args().First()

	if a.triggerType == "http" {
		trig := make([]common.Trigger, 1)
		trig[0] = common.Trigger{
			a.ffV20180707.Name + "-trigger",
			a.triggerType,
			"/" + a.ffV20180707.Name + "-trigger",
		}
		a.ffV20180707.Triggers = trig
	}

	runtime := c.String("runtime")

	runtimeSpecified := runtime != ""

	a.ffV20180707.Schema_version = common.LatestYamlVersion
	if runtimeSpecified {
		// go no further if the specified runtime is not supported
		if runtime != common.FuncfileDockerRuntime && langs.GetLangHelper(runtime) == nil {
			return fmt.Errorf("Init does not support the '%s' runtime", runtime)
		}
	}

	path := c.Args().First()
	if path != "" {
		fmt.Printf("Creating function at: /%s\n", path)
		dir = filepath.Join(dir, path)

		// check if dir exists, if it does, then we can't create function
		if common.Exists(dir) {
			if !a.force {
				return fmt.Errorf("directory %s already exists, cannot init function", dir)
			}
		} else {
			err = os.MkdirAll(dir, 0755)
			if err != nil {
				return err
			}
		}
	}

	err = os.Chdir(dir)
	if err != nil {
		return err
	}

	defer os.Chdir(dir) // todo: wrap this so we can log the error if changing back fails

	if !a.force {
		_, ff, err := common.LoadFuncfile(dir)
		if _, ok := err.(*common.NotFoundError); !ok && err != nil {
			return err
		}
		if ff != nil {
			return errors.New("Function file already exists, aborting")
		}
	}
	err = a.BuildFuncFileV20180707(c, dir) // TODO: Return LangHelper here, then don't need to refind the helper in generateBoilerplate() below
	if err != nil {
		return err
	}

	a.ffV20180707.Schema_version = common.LatestYamlVersion

	// TODO: why don't we treat "docker" runtime as just another language helper? Then can get rid of several Docker
	// specific if/else's like this one.
	if runtimeSpecified && runtime != common.FuncfileDockerRuntime {
		err := a.generateBoilerplate(dir, runtime)
		if err != nil {
			return err
		}
	}

	if err := common.EncodeFuncFileV20180707YAML("func.yaml", a.ffV20180707); err != nil {
		return err
	}

	fmt.Println("func.yaml created.")
	return nil

}

func (a *initFnCmd) generateBoilerplate(path, runtime string) error {
	helper := langs.GetLangHelper(runtime)
	if helper != nil && helper.HasBoilerplate() {
		if err := helper.GenerateBoilerplate(path); err != nil {
			if err == langs.ErrBoilerplateExists {
				return nil
			}
			return err
		}
		fmt.Println("Function boilerplate generated.")
	}
	return nil
}

func (a *initFnCmd) bindRoute(fn *models.Route) {
	ff := a.ff
	if fn.Format != "" {
		ff.Format = fn.Format
	}
	if fn.Memory > 0 {
		ff.Memory = fn.Memory
	}
	if fn.Timeout != nil {
		ff.Timeout = fn.Timeout
	}
	if fn.IDLETimeout != nil {
		ff.IDLETimeout = fn.IDLETimeout
	}
}

func (a *initFnCmd) bindFn(fn *modelsV2.Fn) {
	ff := a.ffV20180707
	if fn.Format != "" {
		ff.Format = fn.Format
	}
	if fn.Mem > 0 {
		ff.Memory = fn.Mem
	}
	if fn.Timeout != nil {
		ff.Timeout = fn.Timeout
	}
	if fn.IDLETimeout != nil {
		ff.IDLE_timeout = fn.IDLETimeout
	}
}

// ValidateFuncName checks if the func name is valid, the name can't contain a colon and
// must be all lowercase
func ValidateFuncName(name string) error {
	if strings.Contains(name, ":") {
		return errors.New("Function name cannot contain a colon")
	}
	if strings.ToLower(name) != name {
		return errors.New("Function name must be lowercase")
	}
	return nil
}

func (a *initFnCmd) BuildFuncFile(c *cli.Context, path string) error {
	var err error

	if c.String("name") != "" {
		a.ff.Name = strings.ToLower(c.String("name"))
	}

	if a.ff.Name == "" {
		// then defaults to current directory for name, the name must be lowercase
		a.ff.Name = strings.ToLower(filepath.Base(path))
	}

	a.ff.Version = c.String("version")

	if err = ValidateFuncName(a.ff.Name); err != nil {
		return err
	}

	if (c.String("init-image") != ""){
		// Building from an image only requires us to have
		// Name and Version generated here.
		return nil
	}

	//if Dockerfile present, use 'docker' as 'runtime'
	if common.Exists("Dockerfile") {
		fmt.Println("Dockerfile found. Using runtime 'docker'.")
		a.ff.Runtime = common.FuncfileDockerRuntime
		return nil
	}

	runtime := c.String("runtime")
	if runtime == common.FuncfileDockerRuntime {
		return errors.New("Function file runtime is 'docker', but no Dockerfile exists")
	}

	var helper langs.LangHelper
	if runtime == "" {
		helper, err = detectRuntime(path)
		if err != nil {
			return err
		}
		fmt.Printf("Found %v function, assuming %v runtime.\n", helper.Runtime(), helper.Runtime())
		//need to default this to default format to be backwards compatible. Might want to just not allow this anymore, fail here.
		if a.ff.Format == "" {
			a.ff.Format = "default"
		}
	} else {
		fmt.Println("Runtime:", runtime)
		helper = langs.GetLangHelper(runtime)
	}
	if helper == nil {
		fmt.Printf("Init does not support the %s runtime, you'll have to create your own Dockerfile for this function.\n", a.ff.Runtime)
	} else {
		if c.String("entrypoint") == "" {
			a.ff.Entrypoint, err = helper.Entrypoint()
			if err != nil {
				return err
			}
		}

		if runtime == "" {
			runtime = helper.Runtime()
		}

		a.ff.Runtime = runtime

		if c.String("format") == "" {
			a.ff.Format = helper.DefaultFormat()
		}

		if c.String("cmd") == "" {
			cmd, err := helper.Cmd()
			if err != nil {
				return err
			}
			a.ff.Cmd = cmd
		}

		if helper.FixImagesOnInit() {
			if a.ff.BuildImage == "" {
				buildImage, err := helper.BuildFromImage()
				if err != nil {
					return err
				}
				a.ff.BuildImage = buildImage
			}
			if helper.IsMultiStage() {
				if a.ff.RunImage == "" {
					runImage, err := helper.RunFromImage()
					if err != nil {
						return err
					}
					a.ff.RunImage = runImage
				}
			}
		}
	}

	if a.ff.Entrypoint == "" && a.ff.Cmd == "" {
		return fmt.Errorf("Could not detect entrypoint or cmd for %v, use --entrypoint and/or --cmd to set them explicitly", runtime)
	}

	return nil
}

func (a *initFnCmd) BuildFuncFileV20180707(c *cli.Context, path string) error {
	var err error

	if c.String("name") != "" {
		a.ffV20180707.Name = strings.ToLower(c.String("name"))
	}

	if a.ffV20180707.Name == "" {
		// then defaults to current directory for name, the name must be lowercase
		a.ffV20180707.Name = strings.ToLower(filepath.Base(path))
	}

	a.ffV20180707.Version = c.String("version")
	if err = ValidateFuncName(a.ffV20180707.Name); err != nil {
		return err
	}

	//if Dockerfile present, use 'docker' as 'runtime'
	if common.Exists("Dockerfile") {
		fmt.Println("Dockerfile found. Using runtime 'docker'.")
		a.ff.Runtime = common.FuncfileDockerRuntime
		return nil
	}
	runtime := c.String("runtime")
	if runtime == common.FuncfileDockerRuntime {
		return errors.New("Function file runtime is 'docker', but no Dockerfile exists")
	}

	var helper langs.LangHelper
	if runtime == "" {
		helper, err = detectRuntime(path)
		if err != nil {
			return err
		}
		fmt.Printf("Found %v function, assuming %v runtime.\n", helper.Runtime(), helper.Runtime())
		//need to default this to default format to be backwards compatible. Might want to just not allow this anymore, fail here.
		if c.String("format") == "" {
			a.ffV20180707.Format = "default"
		}
	} else {
		fmt.Println("Runtime:", runtime)
		helper = langs.GetLangHelper(runtime)
	}
	if helper == nil {
		fmt.Printf("Init does not support the %s runtime, you'll have to create your own Dockerfile for this function.\n", runtime)
	} else {
		if c.String("entrypoint") == "" {
			a.ffV20180707.Entrypoint, err = helper.Entrypoint()
			if err != nil {
				return err
			}
		}

		if runtime == "" {
			a.ffV20180707.Runtime = helper.Runtime()
		}

		a.ffV20180707.Runtime = runtime

		if c.String("format") == "" {
			a.ffV20180707.Format = helper.DefaultFormat()
		}

		if c.String("cmd") == "" {
			cmd, err := helper.Cmd()
			if err != nil {
				return err
			}
			a.ffV20180707.Cmd = cmd
		}

		if helper.FixImagesOnInit() {
			if a.ffV20180707.Build_image == "" {
				buildImage, err := helper.BuildFromImage()
				if err != nil {
					return err
				}
				a.ffV20180707.Build_image = buildImage
			}
			if helper.IsMultiStage() {
				if a.ffV20180707.Run_image == "" {
					runImage, err := helper.RunFromImage()
					if err != nil {
						return err
					}
					a.ffV20180707.Run_image = runImage
				}
			}
		}
	}

	if a.ffV20180707.Entrypoint == "" && a.ffV20180707.Cmd == "" {
		return fmt.Errorf("Could not detect entrypoint or cmd for %v, use --entrypoint and/or --cmd to set them explicitly", a.ffV20180707.Runtime)
	}

	return nil
}

func detectRuntime(path string) (langs.LangHelper, error) {
	for _, h := range langs.Helpers() {
		filenames := []string{}
		for _, ext := range h.Extensions() {
			filenames = append(filenames,
				filepath.Join(path, fmt.Sprintf("func%s", ext)),
				filepath.Join(path, fmt.Sprintf("Func%s", ext)),
				filepath.Join(path, fmt.Sprintf("src/main%s", ext)), // rust
			)
		}
		for _, filename := range filenames {
			if common.Exists(filename) {
				return h, nil
			}
		}
	}
	return nil, fmt.Errorf("No supported files found to guess runtime, please set runtime explicitly with --runtime flag")
}
