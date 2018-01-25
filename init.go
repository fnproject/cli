package main

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
   and no Dockerfile exists,  print an error message then exit
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

	"github.com/fnproject/cli/langs"
	"github.com/fnproject/fn_go/models"
	"github.com/urfave/cli"
)

type initFnCmd struct {
	force bool
	ff    *funcfile
}

func initFlags(a *initFnCmd) []cli.Flag {
	fgs := []cli.Flag{
		cli.StringFlag{
			Name:        "name",
			Usage:       "name of the function. Defaults to directory name.",
			Destination: &a.ff.Name,
		},
		cli.BoolFlag{
			Name:        "force",
			Usage:       "overwrite existing func.yaml",
			Destination: &a.force,
		},
		cli.StringFlag{
			Name:        "runtime",
			Usage:       "choose an existing runtime - " + langsList(),
			Destination: &a.ff.Runtime,
		},
		cli.StringFlag{
			Name:        "entrypoint",
			Usage:       "entrypoint is the command to run to start this function - equivalent to Dockerfile ENTRYPOINT.",
			Destination: &a.ff.Entrypoint,
		},
		cli.StringFlag{
			Name:        "cmd",
			Usage:       "command to run to start this function - equivalent to Dockerfile CMD.",
			Destination: &a.ff.Entrypoint,
		},
		cli.StringFlag{
			Name:        "version",
			Usage:       "set initial function version",
			Destination: &a.ff.Version,
			Value:       initialVersion,
		},
	}

	return append(fgs, routeFlags...)
}

func langsList() string {
	allLangs := []string{}
	for _, h := range langs.Helpers() {
		allLangs = append(allLangs, h.LangStrings()...)
	}
	return strings.Join(allLangs, ", ")
}

func initFn() cli.Command {
	a := &initFnCmd{ff: &funcfile{}}

	return cli.Command{
		Name:        "init",
		Usage:       "create a local func.yaml file",
		Description: "Creates a func.yaml file in the current directory.",
		ArgsUsage:   "[FUNCTION_NAME]",
		Action:      a.init,
		Flags:       initFlags(a),
	}
}

func (a *initFnCmd) init(c *cli.Context) error {
	wd := getWd()

	var rt models.Route
	routeWithFlags(c, &rt)
	a.bindRoute(&rt)

	runtimeSpecified := a.ff.Runtime != ""
	if runtimeSpecified {
		// go no further if the specified runtime is not supported
		if a.ff.Runtime != funcfileDockerRuntime && langs.GetLangHelper(a.ff.Runtime) == nil {
			return fmt.Errorf("Init does not support the '%s' runtime.", a.ff.Runtime)
		}
	}

	var err error
	path := c.Args().First()
	if path != "" {
		fmt.Printf("Creating function at: /%s\n", path)
		dir := filepath.Join(wd, path)
		// check if dir exists, if it does, then we can't create function
		if exists(dir) {
			if !a.force {
				return fmt.Errorf("directory %s already exists, cannot init function", dir)
			}
		} else {
			err = os.MkdirAll(dir, 0755)
			if err != nil {
				return err
			}
		}
		err = os.Chdir(dir)
		if err != nil {
			return err
		}
		defer os.Chdir(wd) // todo: wrap this so we can log the error if changing back fails
	}

	if !a.force {
		_, ff, err := loadFuncfile()
		if _, ok := err.(*notFoundError); !ok && err != nil {
			return err
		}
		if ff != nil {
			return errors.New("Function file already exists, aborting.")
		}
	}

	err = a.buildFuncFile(c) // TODO: Return LangHelper here, then don't need to refind the helper in generateBoilerplate() below
	if err != nil {
		return err
	}

	// TODO: why don't we treat "docker" runtime as just another language helper? Then can get rid of several Docker
	// specific if/else's like this one.
	if runtimeSpecified && a.ff.Runtime != funcfileDockerRuntime {
		err := a.generateBoilerplate()
		if err != nil {
			return err
		}
	}

	if err := encodeFuncfileYAML("func.yaml", a.ff); err != nil {
		return err
	}
	fmt.Println("func.yaml created.")
	return nil
}

func (a *initFnCmd) generateBoilerplate() error {
	helper := langs.GetLangHelper(a.ff.Runtime)
	if helper != nil && helper.HasBoilerplate() {
		if err := helper.GenerateBoilerplate(); err != nil {
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
	if rt.Timeout != nil {
		ff.Timeout = rt.Timeout
	}
	if rt.IDLETimeout != nil {
		ff.IDLETimeout = rt.IDLETimeout
	}
}

func (a *initFnCmd) buildFuncFile(c *cli.Context) error {
	wd := getWd()
	var err error

	if a.ff.Name == "" {
		// then defaults to current directory for name, we'll just leave it out of func.yaml
		// a.Name = filepath.Base(pwd)
	} else if strings.Contains(a.ff.Name, ":") {
		return errors.New("function name cannot contain a colon")
	}

	//if Dockerfile present, use 'docker' as 'runtime'
	if exists("Dockerfile") {
		fmt.Println("Dockerfile found. Using runtime 'docker'.")
		a.ff.Runtime = funcfileDockerRuntime
		return nil
	}
	if a.ff.Runtime == funcfileDockerRuntime {
		return errors.New("function file runtime is 'docker', but no Dockerfile exists")
	}

	var helper langs.LangHelper
	if a.ff.Runtime == "" {
		helper, err = detectRuntime(wd)
		if err != nil {
			return err
		}
		fmt.Printf("Found %v function, assuming %v runtime.\n", helper.Runtime(), helper.Runtime())
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
		return fmt.Errorf("could not detect entrypoint or cmd for %v, use --entrypoint and/or --cmd to set them explicitly", a.ff.Runtime)
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
			if exists(filename) {
				return h, nil
			}
		}
	}
	return nil, fmt.Errorf("no supported files found to guess runtime, please set runtime explicitly with --runtime flag")
}
