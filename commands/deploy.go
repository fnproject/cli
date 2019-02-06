package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	client "github.com/fnproject/cli/client"
	common "github.com/fnproject/cli/common"
	apps "github.com/fnproject/cli/objects/app"
	function "github.com/fnproject/cli/objects/fn"
	trigger "github.com/fnproject/cli/objects/trigger"
	v2Client "github.com/fnproject/fn_go/clientv2"
	clientApps "github.com/fnproject/fn_go/clientv2/apps"
	modelsV2 "github.com/fnproject/fn_go/modelsv2"
	"github.com/urfave/cli"
)

// DeployCommand returns deploy cli.command
func DeployCommand() cli.Command {
	cmd := deploycmd{}
	var flags []cli.Flag
	flags = append(flags, cmd.flags()...)
	return cli.Command{
		Name:    "deploy",
		Usage:   "\tDeploys a function to the functions server (bumps, build, pushes and updates functions and/or triggers).",
		Aliases: []string{"dp"},
		Before: func(cxt *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			cmd.clientV2 = provider.APIClientv2()
			return nil
		},
		Category:    "DEVELOPMENT COMMANDS",
		Description: "This command deploys one or all (--all) functions to the function server.",
		ArgsUsage:   "[function-subdirectory]",
		Flags:       flags,
		Action:      cmd.deploy,
	}
}

type deploycmd struct {
	clientV2 *v2Client.Fn

	appName   string
	createApp bool
	wd        string
	verbose   bool
	local     bool
	noCache   bool
	registry  string
	all       bool
	noBump    bool
}

func (p *deploycmd) flags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:        "app",
			Usage:       "App name to deploy to",
			Destination: &p.appName,
		},
		cli.BoolFlag{
			Name:        "create-app",
			Usage:       "Enable automatic creation of app if it doesn't exist during deploy",
			Destination: &p.createApp,
		},
		cli.BoolFlag{
			Name:        "verbose, v",
			Usage:       "Verbose mode",
			Destination: &p.verbose,
		},
		cli.BoolFlag{
			Name:        "no-cache",
			Usage:       "Don't use Docker cache for the build",
			Destination: &p.noCache,
		},
		cli.BoolFlag{
			Name:        "local, skip-push", // todo: deprecate skip-push
			Usage:       "Do not push Docker built images onto Docker Hub - useful for local development.",
			Destination: &p.local,
		},
		cli.StringFlag{
			Name:        "registry",
			Usage:       "Set the Docker owner for images and optionally the registry. This will be prefixed to your function name for pushing to Docker registries.\r  eg: `--registry username` will set your Docker Hub owner. `--registry registry.hub.docker.com/username` will set the registry and owner. ",
			Destination: &p.registry,
		},
		cli.BoolFlag{
			Name:        "all",
			Usage:       "If in root directory containing `app.yaml`, this will deploy all functions",
			Destination: &p.all,
		},
		cli.BoolFlag{
			Name:        "no-bump",
			Usage:       "Do not bump the version, assuming external version management",
			Destination: &p.noBump,
		},
		cli.StringSliceFlag{
			Name:  "build-arg",
			Usage: "Set build time variables",
		},
		cli.StringFlag{
			Name:  "working-dir,w",
			Usage: "Specify the working directory to deploy a function, must be the full path.",
		},
	}
}

// deploy deploys a function or a set of functions for an app
// By default this will deploy a single function, either the function in the current directory
// or if an arg is passed in, a function in the path representing that arg, relative to the
// current working directory.
//
// If user passes in --all flag, it will deploy all functions in an app. An app must have an `app.yaml`
// file in it's root directory. The functions will be deployed based on the directory structure
// on the file system (can be overridden using the `path` arg in each `func.yaml`. The index/root function
// is the one that lives in the same directory as the app.yaml.
func (p *deploycmd) deploy(c *cli.Context) error {
	appName := ""
	dir := common.GetDir(c)

	appf, err := common.LoadAppfile(dir)

	if err != nil {
		if _, ok := err.(*common.NotFoundError); ok {
			if p.all {
				return err
			}
			// otherwise, it's ok
		} else {
			return err
		}
	} else {
		appName = appf.Name
	}
	if p.appName != "" {
		// flag overrides all
		appName = p.appName
	}

	if appName == "" {
		return errors.New("App name must be provided, try `--app APP_NAME`")
	}

	if p.all {
		return p.deployAll(c, appName, appf)
	}
	return p.deploySingle(c, appName, appf)
}

// deploySingle deploys a single function, either the current directory or if in the context
// of an app and user provides relative path as the first arg, it will deploy that function.
func (p *deploycmd) deploySingle(c *cli.Context, appName string, appf *common.AppFile) error {
	var dir string
	wd := common.GetWd()

	if c.String("working-dir") != "" {
		dir = c.String("working-dir")
	} else {
		// if we're in the context of an app, first arg is path to the function
		path := c.Args().First()
		if path != "" {
			fmt.Printf("Deploying function at: /%s\n", path)
		}
		dir = filepath.Join(wd, path)
	}

	err := os.Chdir(dir)
	if err != nil {
		return err
	}
	defer os.Chdir(wd)

	ffV, err := common.ReadInFuncFile()
	if err != nil {
		return err
	}

	switch common.GetFuncYamlVersion(ffV) {
	case common.LatestYamlVersion:
		fpath, ff, err := common.FindAndParseFuncFileV20180708(dir)
		if err != nil {
			return err
		}
		if appf != nil {
			if dir == wd {
				setFuncInfoV20180708(ff, appf.Name)
			}
		}

		if appf != nil {
			err = p.updateAppConfig(appf)
			if err != nil {
				return fmt.Errorf("Failed to update app config: %v", err)
			}
		}

		return p.deployFuncV20180708(c, appName, wd, fpath, ff)
	default:
		return fmt.Errorf("routes are no longer supported, please use the migrate command to update your metadata")
	}
}

// deployAll deploys all functions in an app.
func (p *deploycmd) deployAll(c *cli.Context, appName string, appf *common.AppFile) error {
	if appf != nil {
		err := p.updateAppConfig(appf)
		if err != nil {
			return fmt.Errorf("Failed to update app config: %v", err)
		}
	}

	var dir string
	wd := common.GetWd()

	if c.String("dir") != "" {
		dir = c.String("dir")
	} else {
		dir = wd
	}

	var funcFound bool
	err := common.WalkFuncsV20180708(dir, func(path string, ff *common.FuncFileV20180708, err error) error {
		if err != nil { // probably some issue with funcfile parsing, can decide to handle this differently if we'd like
			return err
		}
		dir := filepath.Dir(path)
		if dir == wd {
			setFuncInfoV20180708(ff, appName)
		} else {
			// change dirs
			err = os.Chdir(dir)
			if err != nil {
				return err
			}
			p2 := strings.TrimPrefix(dir, wd)
			if ff.Name == "" {
				ff.Name = strings.Replace(p2, "/", "-", -1)
				if strings.HasPrefix(ff.Name, "-") {
					ff.Name = ff.Name[1:]
				}
				// todo: should we prefix appname too?
			}
		}

		err = p.deployFuncV20180708(c, appName, wd, path, ff)
		if err != nil {
			return fmt.Errorf("deploy error on %s: %v", path, err)
		}

		now := time.Now()
		os.Chtimes(path, now, now)
		funcFound = true
		return nil
	})
	if err != nil {
		return err
	}

	if !funcFound {
		return errors.New("No functions found to deploy")
	}

	return nil
}

func (p *deploycmd) deployFuncV20180708(c *cli.Context, appName, baseDir, funcfilePath string, funcfile *common.FuncFileV20180708) error {
	if appName == "" {
		return errors.New("App name must be provided, try `--app APP_NAME`")
	}

	if funcfile.Name == "" {
		funcfile.Name = filepath.Base(filepath.Dir(funcfilePath)) // todo: should probably make a copy of ff before changing it
	}
	fmt.Printf("Deploying %s to app: %s\n", funcfile.Name, appName)

	var err error
	if !p.noBump {
		funcfile2, err := common.BumpItV20180708(funcfilePath, common.Patch)
		if err != nil {
			return err
		}
		funcfile.Version = funcfile2.Version
		// TODO: this whole funcfile handling needs some love, way too confusing. Only bump makes permanent changes to it.
	}

	buildArgs := c.StringSlice("build-arg")
	_, err = common.BuildFuncV20180708(c.GlobalBool("verbose"), funcfilePath, funcfile, buildArgs, p.noCache)
	if err != nil {
		return err
	}

	if !p.local {
		if err := common.DockerPushV20180708(funcfile); err != nil {
			return err
		}
	}

	return p.updateFunction(c, appName, funcfile)
}

func setRootFuncInfo(ff *common.FuncFile, appName string) {
	if ff.Name == "" {
		fmt.Println("Setting name")
		ff.Name = fmt.Sprintf("%s-root", appName)
	}
	if ff.Path == "" {
		// then in root dir, so this will be deployed at /
		ff.Path = "/"
	}
}

func setFuncInfoV20180708(ff *common.FuncFileV20180708, appName string) {
	if ff.Name == "" {
		fmt.Println("Setting name")
		ff.Name = fmt.Sprintf("%s-root", appName)
	}
}

func (p *deploycmd) updateFunction(c *cli.Context, appName string, ff *common.FuncFileV20180708) error {
	fmt.Printf("Updating function %s using image %s...\n", ff.Name, ff.ImageNameV20180708())

	fn := &modelsV2.Fn{}
	if err := function.WithFuncFileV20180708(ff, fn); err != nil {
		return fmt.Errorf("Error getting function with funcfile: %s", err)
	}

	app, err := apps.GetAppByName(p.clientV2, appName)
	if err != nil {
		if p.createApp {
			app = &modelsV2.App{
				Name: appName,
			}

			err = apps.CreateApp(p.clientV2, app)
			if err != nil {
				return err
			}
			app, err = apps.GetAppByName(p.clientV2, appName)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	fnRes, err := function.GetFnByName(p.clientV2, app.ID, ff.Name)
	if err != nil {
		fn.Name = ff.Name
		err := function.CreateFn(p.clientV2, appName, fn)
		if err != nil {
			return err
		}
	} else {
		fn.ID = fnRes.ID
		err = function.PutFn(p.clientV2, fn.ID, fn)
		if err != nil {
			return err
		}
	}

	if fnRes == nil {
		fn, err = function.GetFnByName(p.clientV2, app.ID, ff.Name)
		if err != nil {
			return err
		}
	}

	if len(ff.Triggers) != 0 {
		for _, t := range ff.Triggers {
			trig := &modelsV2.Trigger{
				AppID:  app.ID,
				FnID:   fn.ID,
				Name:   t.Name,
				Source: t.Source,
				Type:   t.Type,
			}

			trigs, err := trigger.GetTriggerByName(p.clientV2, app.ID, fn.ID, t.Name)
			if err != nil {
				err = trigger.CreateTrigger(p.clientV2, trig)
				if err != nil {
					return err
				}
			} else {
				trig.ID = trigs.ID
				err = trigger.PutTrigger(p.clientV2, trig)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (p *deploycmd) updateAppConfig(appf *common.AppFile) error {
	app, err := apps.GetAppByName(p.clientV2, appf.Name)
	if err != nil {
		switch err.(type) {
		case apps.NameNotFoundError:
			param := clientApps.NewCreateAppParams()
			param.Body = &modelsV2.App{
				Name:        appf.Name,
				Config:      appf.Config,
				Annotations: appf.Annotations,
			}
			if _, err = p.clientV2.Apps.CreateApp(param); err != nil {
				return err
			}
			return nil
		default:
			return err
		}
	}
	param := clientApps.NewUpdateAppParams()
	param.AppID = app.ID
	param.Body = &modelsV2.App{
		Config:      appf.Config,
		Annotations: appf.Annotations,
	}

	if _, err = p.clientV2.Apps.UpdateApp(param); err != nil {
		return err
	}
	return nil

}
