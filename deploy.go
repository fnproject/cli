package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	client "github.com/fnproject/cli/client"
	functions "github.com/funcy/functions_go"
	"github.com/funcy/functions_go/models"
	"github.com/urfave/cli"
)

func deploy() cli.Command {
	cmd := deploycmd{
		RoutesApi: functions.NewRoutesApi(),
	}
	var flags []cli.Flag
	flags = append(flags, cmd.flags()...)
	return cli.Command{
		Name:   "deploy",
		Usage:  "deploys a function to the functions server. (bumps, build, pushes and updates route)",
		Flags:  flags,
		Action: cmd.deploy,
	}
}

type deploycmd struct {
	appName string
	*functions.RoutesApi

	wd       string
	verbose  bool
	local    bool
	noCache  bool
	registry string
	all      bool
}

func (cmd *deploycmd) Registry() string {
	return cmd.registry
}

func (p *deploycmd) flags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:        "app",
			Usage:       "app name to deploy to",
			Destination: &p.appName,
		},
		cli.BoolFlag{
			Name:        "verbose, v",
			Usage:       "verbose mode",
			Destination: &p.verbose,
		},
		cli.BoolFlag{
			Name:        "no-cache",
			Usage:       "Don't use Docker cache for the build",
			Destination: &p.noCache,
		},
		cli.BoolFlag{
			Name:        "local, skip-push", // todo: deprecate skip-push
			Usage:       "does not push Docker built images onto Docker Hub - useful for local development.",
			Destination: &p.local,
		},
		cli.StringFlag{
			Name:        "registry",
			Usage:       "Sets the Docker owner for images and optionally the registry. This will be prefixed to your function name for pushing to Docker registries. eg: `--registry username` will set your Docker Hub owner. `--registry registry.hub.docker.com/username` will set the registry and owner.",
			Destination: &p.registry,
		},
		cli.BoolFlag{
			Name:        "all",
			Usage:       "if in root directory containing `app.yaml`, this will deploy all functions",
			Destination: &p.all,
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
	setRegistryEnv(p)

	appName := ""

	appf, err := loadAppfile()
	if err != nil {
		if _, ok := err.(*notFoundError); ok {
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
		return errors.New("app name must be provided, try `--app APP_NAME`.")
	}

	if !p.all {
		return p.deploySingle(c, appName, appf)
	}

	return p.deployAll(c, appName, appf)
}

// deploySingle deploys a single function, either the current directory or if in the context
// of an app and user provides relative path as the first arg, it will deploy that function.
func (p *deploycmd) deploySingle(c *cli.Context, appName string, appf *appfile) error {
	wd := getWd()

	dir := wd
	// if we're in the context of an app, first arg is path to the function
	path := c.Args().First()
	if path != "" {
		fmt.Printf("Deploying function at: /%s\n", path)
		dir = filepath.Join(wd, path)
		err := os.Chdir(dir)
		if err != nil {
			return err
		}
		defer os.Chdir(wd) // todo: wrap this so we can log the error if changing back fails
	}

	fpath, ff, err := findAndParseFuncfile(dir)
	if err != nil {
		return err
	}
	if appf != nil && ff.Path == "" && dir == wd {
		ff.Path = "/"
	}
	return p.deployFunc(c, appName, wd, fpath, ff)
}

// deployAll deploys all functions in an app.
func (p *deploycmd) deployAll(c *cli.Context, appName string, appf *appfile) error {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln("Couldn't get working directory:", err)
	}

	var funcFound bool
	err = filepath.Walk(wd, func(path string, info os.FileInfo, err error) error {
		if path != wd && info.IsDir() {
			return nil
		}

		if !isFuncfile(path, info) {
			return nil
		}

		// TODO: test/try this again to speed up deploys.
		if false && !isstale(path) {
			return nil
		}
		// Then we found a func file, so let's deploy it:
		ff, err := parseFuncfile(path)
		if err != nil {
			return err
		}
		dir := filepath.Dir(path)
		if ff.Path == "" && dir == wd {
			// then in root dir, so this will be deployed at /
			ff.Path = "/"
		} else {
			// change dirs
			err = os.Chdir(dir)
			if err != nil {
				return err
			}
		}
		err = p.deployFunc(c, appName, wd, path, ff)
		if err != nil {
			return fmt.Errorf("deploy error on %s: %v", path, err)
		}

		now := time.Now()
		os.Chtimes(path, now, now)
		funcFound = true
		return err
	})
	if err != nil {
		return err
	}

	if !funcFound {
		return errors.New("no functions found to deploy")
	}
	return nil
}

// deployFunc performs several actions to deploy to a functions server.
// Parse func.yaml file, bump version, build image, push to registry, and
// finally it will update function's route. Optionally,
// the route can be overriden inside the func.yaml file.
func (p *deploycmd) deployFunc(c *cli.Context, appName, baseDir, funcfilePath string, funcfile *funcfile) error {
	if appName == "" {
		return errors.New("app name must be provided, try `--app APP_NAME`.")
	}
	dir := filepath.Dir(funcfilePath)
	if funcfile.Path == "" {
		if dir == "." {
			funcfile.Path = "/"
		} else {
			funcfile.Path = "/" + filepath.Base(dir)
		}

	}
	fmt.Printf("Deploying %s to app: %s at path: %s\n", funcfile.Name, appName, funcfile.Path)

	funcfile2, err := bumpIt(funcfilePath, Patch)
	if err != nil {
		return err
	}
	funcfile.Version = funcfile2.Version
	// TODO: this whole funcfile handling needs some love, way too confusing. Only bump makes permanent changes to it.

	_, err = buildfunc(funcfilePath, funcfile, p.noCache)
	if err != nil {
		return err
	}

	if !p.local {
		if err := dockerPush(funcfile); err != nil {
			return err
		}
	}

	return p.updateRoute(c, appName, funcfile)
}

func (p *deploycmd) updateRoute(c *cli.Context, appName string, ff *funcfile) error {
	fmt.Printf("Updating route %s using image %s...\n", ff.Path, ff.ImageName())
	if err := resetBasePath(p.Configuration); err != nil {
		return fmt.Errorf("error setting endpoint: %v", err)
	}

	routesCmd := routesCmd{client: client.APIClient()}
	rt := &models.Route{}
	if err := routeWithFuncFile(ff, rt); err != nil {
		return fmt.Errorf("error getting route with funcfile: %s", err)
	}
	return routesCmd.putRoute(c, appName, ff.Path, rt)
}

func expandEnvConfig(configs map[string]string) map[string]string {
	for k, v := range configs {
		configs[k] = os.ExpandEnv(v)
	}
	return configs
}

// Theory of operation: this takes an optimistic approach to detect whether a
// package must be rebuild/bump/deployed. It loads for all files mtime's and
// compare with functions.json own mtime. If any file is younger than
// functions.json, it triggers a rebuild.
// The problem with this approach is that depending on the OS running it, the
// time granularity of these timestamps might lead to false negatives - that is
// a package that is stale but it is not recompiled. A more elegant solution
// could be applied here, like https://golang.org/src/cmd/go/pkg.go#L1111
func isstale(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return true
	}

	fnmtime := fi.ModTime()
	dir := filepath.Dir(path)
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if info.ModTime().After(fnmtime) {
			return errors.New("found stale package")
		}
		return nil
	})

	return err != nil
}
