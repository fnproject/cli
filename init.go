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

	"sort"
	"strings"


	"github.com/fnproject/cli/langs"
	"github.com/fnproject/fn_go/models"
	"github.com/urfave/cli"
)

var (
	fileExtToRuntime = map[string]string{
		".go":   "go",
		".js":   "node",
		".rb":   "ruby",
		".py":   "python",
		".php":  "php",
		".rs":   "rust",
		".cs":   "dotnet",
		".fs":   "dotnet",
		".java": "java",
	}

	fnInitRuntimes []string
)

func init() {
	for _, rt := range fileExtToRuntime {
		fnInitRuntimes = append(fnInitRuntimes, rt)
	}
	fnInitRuntimes=removeDuplicates(fnInitRuntimes)
}

type initFnCmd struct {
	force bool
	funcfile
}

func initFlags(a *initFnCmd) []cli.Flag {
	fgs := []cli.Flag{
		cli.StringFlag{
			Name:        "name",
			Usage:       "name of the function. Defaults to directory name.",
			Destination: &a.Name,
		},
		cli.BoolFlag{
			Name:        "force",
			Usage:       "overwrite existing func.yaml",
			Destination: &a.force,
		},
		cli.StringFlag{
			Name:        "runtime",
			Usage:       "choose an existing runtime - " + strings.Join(fnInitRuntimes, ", "),
			Destination: &a.Runtime,
		},
		cli.StringFlag{
			Name:        "entrypoint",
			Usage:       "entrypoint is the command to run to start this function - equivalent to Dockerfile ENTRYPOINT.",
			Destination: &a.Entrypoint,
		},
		cli.StringFlag{
			Name:        "cmd",
			Usage:       "command to run to start this function - equivalent to Dockerfile CMD.",
			Destination: &a.Entrypoint,
		},
		cli.StringFlag{
			Name:        "version",
			Usage:       "set initial function version",
			Destination: &a.Version,
			Value:       initialVersion,
		},
	}

	return append(fgs, routeFlags...)
}

func initFn() cli.Command {
	a := &initFnCmd{}

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
			err := os.MkdirAll(dir, 0755)
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

	var rt models.Route
	routeWithFlags(c, &rt)
	a.bindRoute(&rt)

	if !a.force {
		_, ff, err := loadFuncfile()
		if _, ok := err.(*notFoundError); !ok && err != nil {
			return err
		}
		if ff != nil {
			return errors.New("Function file already exists, aborting.")
		}
	}

	err = a.buildFuncFile(c)
	if err != nil {
		return err
	}

	runtimeSpecified := a.Runtime != ""

	if runtimeSpecified && a.Runtime != funcfileDockerRuntime {
		err := a.generateBoilerplate()
		if err != nil {
			return err
		}
	}

	if err := encodeFuncfileYAML("func.yaml", &a.funcfile); err != nil {
		return err
	}
	fmt.Println("func.yaml created.")
	return nil
}

func (a *initFnCmd) generateBoilerplate() error {
	helper := langs.GetLangHelper(a.Runtime)
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
	ff := &a.funcfile
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

	if a.Name == "" {
		// then defaults to current directory for name, we'll just leave it out of func.yaml
		// a.Name = filepath.Base(pwd)
	} else if strings.Contains(a.Name, ":") {
		return errors.New("function name cannot contain a colon")
	}

	//if Dockerfile present, use 'docker' as 'runtime'
	if exists("Dockerfile") {
		fmt.Println("Dockerfile found. Using runtime 'docker'.")
		a.Runtime = funcfileDockerRuntime
		return nil
	}
	if a.Runtime == funcfileDockerRuntime {
		return errors.New("function file runtime is 'docker', but no Dockerfile exists")
	}

	var rt string
	if a.Runtime == "" {
		rt, err = detectRuntime(wd)
		if err != nil {
			return err
		}
		a.Runtime = rt
		fmt.Printf("Found %v function, assuming %v runtime.\n", rt, rt)
	} else {
		fmt.Println("Runtime:", a.Runtime)
	}
	helper := langs.GetLangHelper(a.Runtime)
	if helper == nil {
		fmt.Printf("Init does not support the %s runtime, you'll have to create your own Dockerfile for this function", a.Runtime)
	} else {
		if a.Entrypoint == "" {
			a.Entrypoint, err = helper.Entrypoint()
			if err != nil {
				return err
			}
		}

		if a.Format == "" {
			a.Format = helper.DefaultFormat()
		}

		if a.Cmd == "" {
			cmd, err := helper.Cmd()
			if err != nil {
				return err
			}
			a.Cmd = cmd
		}

		if helper.FixImagesOnInit() {
			if a.BuildImage == "" {
				buildImage, err := helper.BuildFromImage()
				if err != nil {
					return err
				}
				a.BuildImage = buildImage
			}
			if helper.IsMultiStage() {
				if a.RunImage == "" {
					runImage, err := helper.RunFromImage()
					if err != nil {
						return err
					}
					a.RunImage = runImage
				}
			}
		}
	}

	if a.Entrypoint == "" && a.Cmd == "" {
		return fmt.Errorf("could not detect entrypoint or cmd for %v, use --entrypoint and/or --cmd to set them explicitly", a.Runtime)
	}

	return nil
}

func detectRuntime(path string) (runtime string, err error) {
	for ext, runtime := range fileExtToRuntime {
		filenames := []string{
			filepath.Join(path, fmt.Sprintf("func%s", ext)),
			filepath.Join(path, fmt.Sprintf("Func%s", ext)),
			filepath.Join(path, fmt.Sprintf("src/main%s", ext)), // rust
		}
		for _, filename := range filenames {
			if exists(filename) {
				return runtime, nil
			}
		}
	}
	return "", fmt.Errorf("no supported files found to guess runtime, please set runtime explicitly with --runtime flag")
}

func removeDuplicates(elements []string) []string {
    // Use map to record duplicates as we find them.
    encountered := map[string]bool{}
    result := []string{}

    for v := range elements {
        if encountered[elements[v]] == true {
            // Do not add duplicate.
        } else {
            encountered[elements[v]] = true
            // Append to result slice.
            result = append(result, elements[v])
        }
    }
	// Return the new after sort  slice.
	sort.Strings(result)
    return result
}