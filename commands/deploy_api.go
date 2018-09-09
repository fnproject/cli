package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fnproject/cli/common"
	fnclient "github.com/fnproject/fn_go/client"
	v2Client "github.com/fnproject/fn_go/clientv2"
)

func DeploySingle(clientV1 *fnclient.Fn, clientV2 *v2Client.Fn, deployConfig *DeployConfig, dir, appName string, appf *common.AppFile) error {
	wd := common.GetWd()

	err := os.Chdir(dir)
	if err != nil {
		return err
	}
	defer os.Chdir(dir)

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
			err = UpdateAppConfig(clientV1, appf)
			if err != nil {
				return fmt.Errorf("Failed to update app config: %v", err)
			}
		}

		return DeployFuncV20180708(clientV1, clientV2, deployConfig, appName, fpath, ff)
	default:
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
			err = UpdateAppConfig(clientV1, appf)
			if err != nil {
				return fmt.Errorf("Failed to update app config: %v", err)
			}
		}

		return DeployFunc(clientV1, deployConfig, appName, fpath, ff)
	}
}

func DeployAll(client *fnclient.Fn, deployConfig *DeployConfig, appDir, appName string, appf *common.AppFile) error {
	if appf != nil {
		err := UpdateAppConfig(client, appf)
		if err != nil {
			return fmt.Errorf("failed to update app config: %v", err)
		}
	}
	var dir string
	wd := common.GetWd()

	if appDir != "" {
		dir = appDir
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

		err = DeployFunc(client, deployConfig, appName, path, ff)
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

	return nil
}

// DeployFunc performs several actions to deploy to a functions server.
// Parse func.yaml file, bump version, build image, push to registry, and
// finally it will update function's route. Optionally,
// the route can be overriden inside the func.yaml file.
func DeployFunc(client *fnclient.Fn, deployConfig *DeployConfig, appName, funcfilePath string, funcfile *common.FuncFile) error {
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

	var err error
	if !deployConfig.NoBump {
		funcfile2, err := common.BumpIt(funcfilePath, common.Patch)
		if err != nil {
			return err
		}
		funcfile.Version = funcfile2.Version
		// TODO: this whole funcfile handling needs some love, way too confusing. Only bump makes permanent changes to it.
	}

	// p.noCache
	_, err = common.BuildFunc(deployConfig.BuildArgs, deployConfig.Verbose, deployConfig.NoCache, funcfilePath, funcfile)
	if err != nil {
		return err
	}

	// p.local
	if !deployConfig.IsLocal {
		if err := common.DockerPush(funcfile); err != nil {
			return err
		}
	}

	return updateRoute(client, appName, funcfile)
}

func DeployFuncV20180708(clientV1 *fnclient.Fn, clientV2 *v2Client.Fn, deployConfig *DeployConfig, appName, funcfilePath string, funcfile *common.FuncFileV20180708) error {
	if funcfile.Name == "" {
		funcfile.Name = filepath.Base(filepath.Dir(funcfilePath)) // todo: should probably make a copy of ff before changing it
	}

	var err error
	if !deployConfig.NoBump {
		funcfile2, err := common.BumpItV20180708(funcfilePath, common.Patch)
		if err != nil {
			return err
		}
		funcfile.Version = funcfile2.Version
		// TODO: this whole funcfile handling needs some love, way too confusing. Only bump makes permanent changes to it.
	}

	_, err = common.BuildFuncV20180708(deployConfig.BuildArgs, deployConfig.Verbose,
		deployConfig.NoCache, funcfilePath, funcfile)
	if err != nil {
		return err
	}

	if !deployConfig.IsLocal {
		if err := common.DockerPushV20180708(funcfile); err != nil {
			return err
		}
	}

	return updateFunction(clientV1, clientV2, appName, funcfile)
}
