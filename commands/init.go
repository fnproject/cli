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
	"sort"
	"strings"

	"github.com/fnproject/cli/common"
	"github.com/fnproject/cli/langs"
	function "github.com/fnproject/cli/objects/fn"
	modelsV2 "github.com/fnproject/fn_go/modelsv2"
	"github.com/urfave/cli"
)

type initFnCmd struct {
	force       bool
	triggerType string
	wd          string
	ff          *common.FuncFileV20180708
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
			Usage:       "Specify the trigger type - permitted values are 'http'.",
			Destination: &a.triggerType,
		},
		cli.Uint64Flag{
			Name:  "memory,m",
			Usage: "Memory in MiB",
		},
		cli.StringSliceFlag{
			Name:  "config,c",
			Usage: "Function configuration",
		},
		cli.IntFlag{
			Name:  "timeout",
			Usage: "Function timeout (eg. 30)",
		},
		cli.IntFlag{
			Name:  "idle-timeout",
			Usage: "Function idle timeout (eg. 30)",
		},
		cli.StringSliceFlag{
			Name:  "annotation",
			Usage: "Function annotation (can be specified multiple times)",
		},
	}

	return fgs
}

func langsList() string {
	allLangs := []string{}
	for _, h := range langs.Helpers() {
		allLangs = append(allLangs, h.LangStrings()...)
	}
	sort.Strings(allLangs)
	// remove duplicates
	var allUnique []string
	for i, s := range allLangs {
		if i > 0 && s == allLangs[i-1] {
			continue
		}
		if deprecatedPythonRuntime(s) {
			continue
		}
		allUnique = append(allUnique, s)
	}
	return strings.Join(allUnique, ", ")
}

func deprecatedPythonRuntime(runtime string) bool {
	return runtime == "python3.8.5" || runtime == "python3.7.1" || runtime == "python3.9" || runtime == "python3.8"
}

// InitCommand returns init cli.command
func InitCommand() cli.Command {
	a := &initFnCmd{ff: &common.FuncFileV20180708{}}

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
	var fn modelsV2.Fn

	dir = common.GetWd()
	if a.wd != "" {
		dir = a.wd
	}

	function.WithFlags(c, &fn)
	a.bindFn(&fn)

	runtime := c.String("runtime")
	initImage := c.String("init-image")

	if runtime != "" && initImage != "" {
		return fmt.Errorf("You can't supply --runtime with --init-image")
	}

	runtimeSpecified := runtime != ""

	a.ff.Schema_version = common.LatestYamlVersion
	if runtimeSpecified {
		// go no further if the specified runtime is not supported
		if runtime != common.FuncfileDockerRuntime && langs.GetLangHelper(runtime) == nil {
			return fmt.Errorf("Init does not support the '%s' runtime", runtime)
		}
		if deprecatedPythonRuntime(runtime) {
			return fmt.Errorf("Runtime %s is no more supported for new apps. Please use python or %s runtime for new apps.", runtime, runtime[:strings.LastIndex(runtime, ".")])
		}
	}

	path := c.Args().First()
	if path != "" {
		fmt.Printf("Creating function at: ./%s\n", path)
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

	if c.String("name") != "" {
		a.ff.Name = strings.ToLower(c.String("name"))
	}

	if a.ff.Name == "" {
		// then defaults to current directory for name, the name must be lowercase
		a.ff.Name = strings.ToLower(filepath.Base(dir))
	}

	if a.triggerType != "" {
		a.triggerType = strings.ToLower(a.triggerType)
		ok := validateTriggerType(a.triggerType)
		if !ok {
			return fmt.Errorf("init does not support the trigger type '%s'.\n Permitted values are 'http'.", a.triggerType)
		}

		// TODO when we allow multiple trigger definitions in a func file, we need
		// to allow naming triggers in a func file as well as use the type of
		// trigger to deduplicate the trigger names

		trig := make([]common.Trigger, 1)
		trig[0] = common.Trigger{
			Name:   a.ff.Name,
			Type:   a.triggerType,
			Source: "/" + a.ff.Name,
		}

		a.ff.Triggers = trig

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
	err = a.BuildFuncFileV20180708(c, dir) // TODO: Return LangHelper here, then don't need to refind the helper in generateBoilerplate() below
	if err != nil {
		return err
	}

	a.ff.Schema_version = common.LatestYamlVersion

	if initImage != "" {
		err := a.doInitImage(initImage, c)
		if err != nil {
			return err
		}
	} else {
		// TODO: why don't we treat "docker" runtime as just another language helper?
		// Then can get rid of several Docker specific if/else's like this one.
		if runtimeSpecified && runtime != common.FuncfileDockerRuntime {
			err := a.generateBoilerplate(dir, runtime)
			if err != nil {
				return err
			}
		}
	}

	if err := common.EncodeFuncFileV20180708YAML("func.yaml", a.ff); err != nil {
		return err
	}

	fmt.Println("func.yaml created.")
	return nil
}

func (a *initFnCmd) doInitImage(initImage string, c *cli.Context) error {
	err := common.RunInitImage(initImage, a.ff.Name)
	if err != nil {
		return err
	}
	err = common.MergeFuncFileInitYAML("func.init.yaml", a.ff)
	if err != nil {
		return err
	}
	// Then CLI args can override some init-image options (TODO: remove this with #383)
	if c.String("cmd") != "" {
		a.ff.Cmd = c.String("cmd")
	}
	if c.String("entrypoint") != "" {
		a.ff.Entrypoint = c.String("entrypoint")
	}
	_ = os.Remove("func.init.yaml")
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

func (a *initFnCmd) bindFn(fn *modelsV2.Fn) {
	ff := a.ff
	if fn.Memory > 0 {
		ff.Memory = fn.Memory
	}
	if fn.Timeout != nil {
		ff.Timeout = fn.Timeout
	}
	if fn.IdleTimeout != nil {
		ff.IDLE_timeout = fn.IdleTimeout
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

func (a *initFnCmd) BuildFuncFileV20180708(c *cli.Context, path string) error {
	var err error

	a.ff.Version = c.String("version")
	if err = ValidateFuncName(a.ff.Name); err != nil {
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

	if c.String("init-image") != "" {
		return nil
	}

	var helper langs.LangHelper
	if runtime == "" {
		helper, err = detectRuntime(path)
		if err != nil {
			return err
		}
		fmt.Printf("Found %v function, assuming %v runtime.\n", helper.Runtime(), helper.Runtime())
	} else {
		helper = langs.GetLangHelper(runtime)
	}
	if helper == nil {
		fmt.Printf("Init does not support the %s runtime, you'll have to create your own Dockerfile for this function.\n", runtime)
	} else {
		if c.String("entrypoint") == "" {
			a.ff.Entrypoint, err = helper.Entrypoint()
			if err != nil {
				return err
			}

		} else {
			a.ff.Entrypoint = c.String("entrypoint")
		}

		if runtime == "" {
			runtime = helper.Runtime()
		}

		a.ff.Runtime = runtime

		if c.Uint64("memory") == 0 {
			a.ff.Memory = helper.CustomMemory()
		}

		if c.String("cmd") == "" {
			cmd, err := helper.Cmd()
			if err != nil {
				return err
			}
			a.ff.Cmd = cmd
		} else {
			a.ff.Cmd = c.String("cmd")
		}
		if helper.FixImagesOnInit() {
			if a.ff.Build_image == "" {
				buildImage, err := helper.BuildFromImage()
				if err != nil {
					return err
				}
				a.ff.Build_image = buildImage
			}
			if helper.IsMultiStage() {
				if a.ff.Run_image == "" {
					runImage, err := helper.RunFromImage()
					if err != nil {
						return err
					}
					a.ff.Run_image = runImage
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

func validateTriggerType(triggerType string) bool {
	switch triggerType {
	case "http":
		return true
	default:
		return false
	}
}
