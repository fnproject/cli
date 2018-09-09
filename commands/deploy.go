package commands

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/fnproject/cli/client"
	"github.com/fnproject/cli/common"
	fnclient "github.com/fnproject/fn_go/client"
	clientApps "github.com/fnproject/fn_go/client/apps"
	v2Client "github.com/fnproject/fn_go/clientv2"
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
		Usage:   "\tDeploys a function to the functions server (bumps, build, pushes and updates route).",
		Aliases: []string{"dp"},
		Before: func(cxt *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			cmd.client = provider.APIClient()
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
	appName  string
	client   *fnclient.Fn
	clientV2 *v2Client.Fn

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

	buildArgs := c.StringSlice("build-arg")
	verbose := c.GlobalBool("verbose")
	noBump := p.noBump
	isLocal := p.local
	noCache := p.noCache
	appDir := c.String("dir")

	if p.all {
		return DeployAll(
			p.client, buildArgs,
			verbose, noBump,
			isLocal, noCache,
			appDir, appName, appf)
	}
	return p.deploySingle(c, appName, appf)
}

// deploySingle deploys a single function, either the current directory or if in the context
// of an app and user provides relative path as the first arg, it will deploy that function.
func (p *deploycmd) deploySingle(c *cli.Context, appName string, appf *common.AppFile) error {
	var dir string
	wd := common.GetWd()

	fpath := c.Args().First()
	workingDir := c.String("working-dir")

	if workingDir != "" {
		dir = workingDir
	} else {
		// if we're in the context of an app, first arg is path to the function
		if fpath != "" {
			fmt.Printf("Deploying function at: /%s\n", fpath)
		}
		dir = filepath.Join(wd, fpath)
	}

	buildArgs := c.StringSlice("build-arg")
	verbose := c.GlobalBool("verbose")
	noBump := p.noBump
	isLocal := p.local
	noCache := p.noCache

	return DeploySingle(p.client, p.clientV2, buildArgs, verbose,
		noBump, isLocal, noCache, dir, appName, appf)
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

func UpdateAppConfig(client *fnclient.Fn, appf *common.AppFile) error {
	param := clientApps.NewPatchAppsAppParams()
	param.App = appf.Name
	param.Body = &models.AppWrapper{
		App: &models.App{
			Config:      appf.Config,
			Annotations: appf.Annotations,
		},
	}

	_, err := client.Apps.PatchAppsApp(param)
	if err != nil {
		postParams := clientApps.NewPostAppsParams() //XXX switch to put when v2.0 Fn
		postParams.Body = &models.AppWrapper{
			App: &models.App{
				Name:        appf.Name,
				Config:      appf.Config,
				Annotations: appf.Annotations,
			},
		}

		_, err = client.Apps.PostApps(postParams)
		if err != nil {
			return err
		}
	}
	return nil
}
