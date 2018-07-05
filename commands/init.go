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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fnproject/cli/common"
	"github.com/fnproject/cli/langs"
	"github.com/fnproject/cli/objects/route"
	"github.com/fnproject/fn_go/models"
	"github.com/urfave/cli"
)

type initFnCmd struct {
	force bool
	ff    *common.FuncFile
}

func initFlags(a *initFnCmd) []cli.Flag {
	fgs := []cli.Flag{
		cli.StringFlag{
			Name:        "name",
			Usage:       "Name of the function. Defaults to directory name in lowercase.",
			Destination: &a.ff.Name,
		},
		cli.BoolFlag{
			Name:        "force",
			Usage:       "Overwrite existing func.yaml",
			Destination: &a.force,
		},
		cli.StringFlag{
			Name:        "runtime",
			Usage:       "Choose an existing runtime - " + langsList(),
			Destination: &a.ff.Runtime,
		},
		cli.StringFlag{
			Name:        "entrypoint",
			Usage:       "Entrypoint is the command to run to start this function - equivalent to Dockerfile ENTRYPOINT.",
			Destination: &a.ff.Entrypoint,
		},
		cli.StringFlag{
			Name:        "cmd",
			Usage:       "Command to run to start this function - equivalent to Dockerfile CMD.",
			Destination: &a.ff.Entrypoint,
		},
		cli.StringFlag{
			Name:        "version",
			Usage:       "Set initial function version",
			Destination: &a.ff.Version,
			Value:       common.InitialVersion,
		},
		cli.StringFlag{
			Name:  "dir",
			Usage: "specify the working directory to init a function",
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
	a := &initFnCmd{ff: &common.FuncFile{}}

	return cli.Command{
		Name:        "init",
		Usage:       "Create a local func.yaml file",
		Category:    "DEVELOPMENT COMMANDS",
		Aliases:     []string{"in"},
		Description: "Creates a func.yaml file in the current directory.",
		ArgsUsage:   "[FUNCTION_NAME]",
		Action:      a.init,
		Flags:       initFlags(a),
	}
}

func (a *initFnCmd) init(c *cli.Context) error {
	var err error
	var dir string

	dir = common.GetWd()

	if c.String("dir") != "" {
		dir = c.String("dir")
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

	var rt models.Route
	route.WithFlags(c, &rt)
	a.bindRoute(&rt)

	runtimeSpecified := a.ff.Runtime != ""
	if runtimeSpecified {
		// go no further if the specified runtime is not supported
		if a.ff.Runtime != common.FuncfileDockerRuntime && langs.GetLangHelper(a.ff.Runtime) == nil {
			return fmt.Errorf("Init does not support the '%s' runtime", a.ff.Runtime)
		}
	}

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

	// TODO: why don't we treat "docker" runtime as just another language helper? Then can get rid of several Docker
	// specific if/else's like this one.
	if runtimeSpecified && a.ff.Runtime != common.FuncfileDockerRuntime {
		err := a.generateBoilerplate(dir)
		if err != nil {
			return err
		}
	}

	if err := common.EncodeFuncfileYAML("func.yaml", a.ff); err != nil {
		return err
	}
	fmt.Println("func.yaml created.")
	return nil
}

func (a *initFnCmd) generateBoilerplate(path string) error {
	helper := langs.GetLangHelper(a.ff.Runtime)
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

func (a *initFnCmd) bindRoute(rt *models.Route) {
	ff := a.ff
	if rt.Format != "" {
		ff.Format = rt.Format
	}
	if rt.Type != "" {
		ff.Type = rt.Type
	}
	if rt.Memory > 0 {
		ff.Memory = rt.Memory
	}
	if rt.Cpus != "" {
		ff.Cpus = rt.Cpus
	}
	if rt.Timeout != nil {
		ff.Timeout = rt.Timeout
	}
	if rt.IDLETimeout != nil {
		ff.IDLETimeout = rt.IDLETimeout
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

	if a.ff.Name == "" {
		// then defaults to current directory for name, the name must be lowercase
		a.ff.Name = strings.ToLower(filepath.Base(path))
	}

	if err = ValidateFuncName(a.ff.Name); err != nil {
		return err
	}

	//if Dockerfile present, use 'docker' as 'runtime'
	if common.Exists("Dockerfile") {
		fmt.Println("Dockerfile found. Using runtime 'docker'.")
		a.ff.Runtime = common.FuncfileDockerRuntime
		return nil
	}
	if a.ff.Runtime == common.FuncfileDockerRuntime {
		return errors.New("Function file runtime is 'docker', but no Dockerfile exists")
	}

	var helper langs.LangHelper
	if a.ff.Runtime == "" {
		helper, err = detectRuntime(path)
		if err != nil {
			return err
		}
		fmt.Printf("Found %v function, assuming %v runtime.\n", helper.Runtime(), helper.Runtime())
		// need to default this to default format to be backwards compatible. Might want to just not allow this anymore, fail here.
		if a.ff.Format == "" {
			a.ff.Format = "default"
		}
	} else {
		fmt.Println("Runtime:", a.ff.Runtime)
		helper = langs.GetLangHelper(a.ff.Runtime)
	}
	if helper == nil {
		fmt.Printf("Init does not support the %s runtime, you'll have to create your own Dockerfile for this function.\n", a.ff.Runtime)
	} else {
		if a.ff.Entrypoint == "" {
			a.ff.Entrypoint, err = helper.Entrypoint()
			if err != nil {
				return err
			}
		}

		if a.ff.Runtime == "" {
			a.ff.Runtime = helper.Runtime()
		}

		if a.ff.Format == "" {
			a.ff.Format = helper.DefaultFormat()
		}

		if a.ff.Cmd == "" {
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
		return fmt.Errorf("Could not detect entrypoint or cmd for %v, use --entrypoint and/or --cmd to set them explicitly", a.ff.Runtime)
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
