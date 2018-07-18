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
	route "github.com/fnproject/cli/objects/route"
	fnclient "github.com/fnproject/fn_go/client"
	clientApps "github.com/fnproject/fn_go/client/apps"
	"github.com/fnproject/fn_go/models"
	"github.com/urfave/cli"
)

// DeployCommand returns deploy cli.command
func DeployCommand() cli.Command {
	cmd := deploycmd{}
	var flags []cli.Flag
	flags = append(flags, cmd.flags()...)
	return cli.Command{
		Name:    "deploy",
		Usage:   "\tDeploys a function to the functions server. (bumps, build, pushes and updates route)",
		Aliases: []string{"dp"},
		Before: func(cxt *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			cmd.client = provider.APIClient()
			return nil
		},
		Category:    "DEVELOPMENT COMMANDS",
		Description: "This command deploys one or all (--all) functions to the function server.",
		Flags:       flags,
		Action:      cmd.deploy,
	}
}

type deploycmd struct {
	appName string
	client  *fnclient.Fn

	wd       string
	verbose  bool
	local    bool
	noCache  bool
	registry string
	all      bool
	noBump   bool
}

func (p *deploycmd) flags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:        "app",
			Usage:       "App name to deploy to",
			Destination: &p.appName,
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

	fmt.Println("dir: ", dir)

	err := os.Chdir(dir)
	if err != nil {
		return err
	}
	defer os.Chdir(wd)

	fpath, ff, err := common.FindAndParseFuncfile(dir)
	if err != nil {
		return err
	}
	if appf != nil {
		if dir == wd {
			setRootFuncInfo(ff, appf.Name)
		}
	}

	if appf != nil {
		err = p.updateAppConfig(appf)
		if err != nil {
			return fmt.Errorf("Failed to update app config: %v", err)
		}
	}

	err = p.deployFunc(c, appName, wd, fpath, ff)
	if err != nil {
		return err
	}

	return nil
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
	err := common.WalkFuncs(dir, func(path string, ff *common.FuncFile, err error) error {
		if err != nil { // probably some issue with funcfile parsing, can decide to handle this differently if we'd like
			return err
		}
		dir := filepath.Dir(path)
		if dir == wd {
			setRootFuncInfo(ff, appName)
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
			if ff.Path == "" {
				ff.Path = p2
			}
		}

		err = p.deployFunc(c, appName, wd, path, ff)
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

// deployFunc performs several actions to deploy to a functions server.
// Parse func.yaml file, bump version, build image, push to registry, and
// finally it will update function's route. Optionally,
// the route can be overriden inside the func.yaml file.
func (p *deploycmd) deployFunc(c *cli.Context, appName, baseDir, funcfilePath string, funcfile *common.FuncFile) error {
	if appName == "" {
		return errors.New("App name must be provided, try `--app APP_NAME`")
	}
	dir := filepath.Dir(funcfilePath)
	// get name from directory if it's not defined
	if funcfile.Name == "" {
		funcfile.Name = filepath.Base(filepath.Dir(funcfilePath)) // todo: should probably make a copy of ff before changing it
	}
	if funcfile.Path == "" {
		if dir == "." {
			funcfile.Path = "/"
		} else {
			funcfile.Path = "/" + filepath.Base(dir)
		}

	}
	fmt.Printf("Deploying %s to app: %s at path: %s\n", funcfile.Name, appName, funcfile.Path)

	var err error
	if !p.noBump {
		funcfile2, err := common.BumpIt(funcfilePath, common.Patch)
		if err != nil {
			return err
		}
		funcfile.Version = funcfile2.Version
		// TODO: this whole funcfile handling needs some love, way too confusing. Only bump makes permanent changes to it.
	}

	buildArgs := c.StringSlice("build-arg")
	_, err = common.BuildFunc(c, funcfilePath, funcfile, buildArgs, p.noCache)
	if err != nil {
		return err
	}

	if !p.local {
		if err := common.DockerPush(funcfile); err != nil {
			return err
		}
	}

	return p.updateRoute(c, appName, funcfile)
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

func (p *deploycmd) updateRoute(c *cli.Context, appName string, ff *common.FuncFile) error {
	fmt.Printf("Updating route %s using image %s...\n", ff.Path, ff.ImageName())
	rt := &models.Route{}
	if err := route.WithFuncFile(ff, rt); err != nil {
		return fmt.Errorf("Error getting route with funcfile: %s", err)
	}
	return route.PutRoute(p.client, appName, ff.Path, rt)
}

func expandEnvConfig(configs map[string]string) map[string]string {
	for k, v := range configs {
		configs[k] = os.ExpandEnv(v)
	}
	return configs
}

func (p *deploycmd) updateAppConfig(appf *common.AppFile) error {
	param := clientApps.NewPatchAppsAppParams()
	param.App = appf.Name
	param.Body = &models.AppWrapper{
		App: &models.App{
			Config:      appf.Config,
			Annotations: appf.Annotations,
		},
	}

	_, err := p.client.Apps.PatchAppsApp(param)
	if err != nil {
		postParams := clientApps.NewPostAppsParams() //XXX switch to put when v2.0 Fn
		postParams.Body = &models.AppWrapper{
			App: &models.App{
				Name:        appf.Name,
				Config:      appf.Config,
				Annotations: appf.Annotations,
			},
		}

		_, err = p.client.Apps.PostApps(postParams)
		if err != nil {
			return err
		}
	}
	return nil
}
