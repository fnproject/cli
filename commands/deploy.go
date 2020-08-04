package commands

import (
	"errors"
	"fmt"
	"github.com/fnproject/cli/adapter"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fnproject/cli/client"
	"github.com/fnproject/cli/common"
	function "github.com/fnproject/cli/objects/fn"
	"github.com/fnproject/cli/objects/trigger"
	v2Client "github.com/fnproject/fn_go/clientv2"
	models "github.com/fnproject/fn_go/modelsv2"
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
			providerAdapter, err := client.CurrentProviderAdapter()
			if err != nil {
				return err
			}
			cmd.clientV2 = provider.APIClientv2()
			cmd.apiClientAdapter = providerAdapter.APIClient()
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
	clientV2      *v2Client.Fn
	apiClientAdapter adapter.APIClient

	appName   string
	createApp bool
	wd        string
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
			Destination: &common.CommandVerbose,
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

	// appfApp is used to create/update app, with app file additions if provided
	appfApp := adapter.App{
		Name: appName,
	}
	if appf != nil {
		// set other fields from app file
		appfApp.Config = appf.Config
		appfApp.Annotations = appf.Annotations
		if appf.SyslogURL != "" {
			// TODO consistent with some other fields (config), unsetting in app.yaml doesn't unset on server. undecided policy for all fields
			appfApp.SyslogURL = &appf.SyslogURL
		}
	}

	// find and create/update app if required
	app, err := p.apiClientAdapter.AppClient().GetApp(appName)
	if _, ok := err.(adapter.AppNameNotFoundError); ok && p.createApp {
		app, err = p.apiClientAdapter.AppClient().CreateApp(&appfApp)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else if appf != nil {
		// app exists, but we need to update it if we have an app file
		appfApp.ID = app.ID
		app, err = p.apiClientAdapter.AppClient().UpdateApp(&appfApp)
		if err != nil {
			return fmt.Errorf("Failed to update app config: %v", err)
		}
	}

	if app == nil {
		panic("app should not be nil here") // tests should catch... better than panic later
	}

	// deploy functions
	if p.all {
		return p.deployAll(c, app)
	}
	return p.deploySingle(c, app)
}

// deploySingle deploys a single function, either the current directory or if in the context
// of an app and user provides relative path as the first arg, it will deploy that function.
func (p *deploycmd) deploySingle(c *cli.Context, app *adapter.App) error {
	var dir string
	wd := common.GetWd()

	if c.String("working-dir") != "" {
		dir = c.String("working-dir")
	} else {
		// if we're in the context of an app, first arg is path to the function
		path := c.Args().First()
		if path != "" {
			fmt.Printf("Deploying function at: ./%s\n", path)
		}
		dir = filepath.Join(wd, path)
	}

	err := os.Chdir(dir)
	if err != nil {
		return err
	}
	defer os.Chdir(wd)

	fpath, ff, err := common.FindAndParseFuncFileV20180708(dir)
	if err != nil {
		return err
	}
	return p.deployFuncV20180708(c, app, fpath, ff)
}

// deployAll deploys all functions in an app.
func (p *deploycmd) deployAll(c *cli.Context, app *adapter.App) error {
	var dir string
	wd := common.GetWd()

	if c.String("working-dir") != "" {
		dir = c.String("working-dir")
	} else {
		// if we're in the context of an app, first arg is path to the function
		path := c.Args().First()
		if path != "" {
			fmt.Printf("Deploying function at: ./%s\n", path)
		}
		dir = filepath.Join(wd, path)
	}

	var funcFound bool
	err := common.WalkFuncsV20180708(dir, func(path string, ff *common.FuncFileV20180708, err error) error {
		if err != nil { // probably some issue with funcfile parsing, can decide to handle this differently if we'd like
			return err
		}
		dir := filepath.Dir(path)
		if dir != wd {
			// change dirs
			err = os.Chdir(dir)
			if err != nil {
				return err
			}
		}
		p2 := strings.TrimPrefix(dir, wd)
		if ff.Name == "" {
			ff.Name = strings.Replace(p2, "/", "-", -1)
			if strings.HasPrefix(ff.Name, "-") {
				ff.Name = ff.Name[1:]
			}
		}

		err = p.deployFuncV20180708(c, app, path, ff)
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

func (p *deploycmd) deployFuncV20180708(c *cli.Context, app *adapter.App, funcfilePath string, funcfile *common.FuncFileV20180708) error {
	if funcfile.Name == "" {
		funcfile.Name = filepath.Base(filepath.Dir(funcfilePath)) // todo: should probably make a copy of ff before changing it
	}
	fmt.Printf("Deploying %s to app: %s\n", funcfile.Name, app.Name)

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
	_, err = common.BuildFuncV20180708(common.IsVerbose(), funcfilePath, funcfile, buildArgs, p.noCache)
	if err != nil {
		return err
	}

	if !p.local {
		if err := common.DockerPushV20180708(funcfile); err != nil {
			return err
		}
	}

	return p.updateFunction(c, app.ID, funcfile)
}

func (p *deploycmd) updateFunction(c *cli.Context, appID string, ff *common.FuncFileV20180708) error {
	fmt.Printf("Updating function %s using image %s...\n", ff.Name, ff.ImageNameV20180708())

	fn := &adapter.Fn{}
	if err := function.WithFuncFileV20180708(ff, fn); err != nil {
		return fmt.Errorf("Error getting function with funcfile: %s", err)
	}

	fnRes, err := p.apiClientAdapter.FnClient().GetFn(appID, ff.Name)
	if _, ok := err.(adapter.FunctionNameNotFoundError); ok {
		fn.Name = ff.Name
		fn.AppID = appID
		fn, err = p.apiClientAdapter.FnClient().CreateFn(fn)
		if err != nil {
			return err
		}
	} else if err != nil {
		// probably service is down or something...
		return err
	} else {
		fn.ID = fnRes.ID
		fn.AppID = appID
		err = function.PutFn(p.apiClientAdapter.FnClient(), fn.ID, fn)
		if err != nil {
			return err
		}
	}

	if len(ff.Triggers) != 0 {
		for _, t := range ff.Triggers {
			trig := &models.Trigger{
				AppID:  appID,
				FnID:   fn.ID,
				Name:   t.Name,
				Source: t.Source,
				Type:   t.Type,
			}

			trigs, err := trigger.GetTriggerByName(p.clientV2, appID, fn.ID, t.Name)
			if _, ok := err.(trigger.NameNotFoundError); ok {
				err = trigger.CreateTrigger(p.clientV2, trig)
				if err != nil {
					return err
				}
			} else if err != nil {
				return err
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
