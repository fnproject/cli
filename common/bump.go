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

package common

import (
	"fmt"
	"strings"

	"github.com/coreos/go-semver/semver"
	bumper "github.com/giantswarm/semver-bump/bump"
	"github.com/giantswarm/semver-bump/storage"
	"github.com/urfave/cli"
)

// VType represents the version type
type VType int

//
const (
	Patch VType = iota
	Minor
	Major
)

var (
	// InitialVersion - inital fn version.
	InitialVersion = "0.0.1"
)

// BumpCommand command to build function version.
func BumpCommand() cli.Command {
	cmd := bumpcmd{}
	flags := append([]cli.Flag{}, cmd.flags()...)
	return cli.Command{
		Name:        "bump",
		Usage:       "\tBump function version",
		Aliases:     []string{"bm"},
		Category:    "DEVELOPMENT COMMANDS",
		Description: "This command bumps the version of the func.yaml.",
		Flags:       flags,
		Action:      cmd.bump,
	}
}

type bumpcmd struct {
	major bool
	minor bool
}

func (b *bumpcmd) flags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:        "major",
			Usage:       "bumps major version",
			Destination: &b.major,
		},
		cli.BoolFlag{
			Name:        "minor",
			Usage:       "bumps minor version",
			Destination: &b.minor,
		},
		cli.BoolFlag{
			Name:        "verbose, v",
			Usage:       "verbose mode",
			Destination: &CommandVerbose,
		},
		cli.StringFlag{
			Name:  "working-dir,w",
			Usage: "Specify the working directory to bump a function, must be the full path.",
		},
	}
}

// bump will take the found valid function and bump its version
func (b *bumpcmd) bump(c *cli.Context) error {
	var t VType
	var dir string

	if b.major {
		t = Major
	} else if b.minor {
		t = Minor
	} else {
		t = Patch
	}

	dir = GetDir(c)

	ff, err := ReadInFuncFile()
	version := GetFuncYamlVersion(ff)
	if version == LatestYamlVersion {
		_, err = bumpItWdV20180708(dir, t)
	} else {
		_, err = bumpItWd(dir, t)
	}

	return err
}

func bumpItWd(wd string, vtype VType) (*FuncFile, error) {
	fn, err := FindFuncfile(wd)
	if err != nil {
		return nil, err
	}
	return BumpIt(fn, vtype)
}

// BumpIt returns updated funcfile
func BumpIt(fpath string, vtype VType) (*FuncFile, error) {
	// fmt.Println("Bumping version in func file at: ", fpath)
	funcfile, err := ParseFuncfile(fpath)
	if err != nil {
		return nil, err
	}

	funcfile, err = bumpVersion(funcfile, vtype)
	if err != nil {
		return nil, err
	}

	if err := storeFuncfile(fpath, funcfile); err != nil {
		return nil, err
	}
	fmt.Println("Bumped to version", funcfile.Version)
	return funcfile, nil
}

func bumpVersion(funcfile *FuncFile, t VType) (*FuncFile, error) {
	funcfile.Name = cleanImageName(funcfile.Name)
	if funcfile.Version == "" {
		funcfile.Version = InitialVersion
		return funcfile, nil
	}

	s, err := storage.NewVersionStorage("local", funcfile.Version)
	if err != nil {
		return nil, err
	}

	version := bumper.NewSemverBumper(s, "")
	var newver *semver.Version
	if t == Major {
		newver, err = version.BumpMajorVersion("", "")
	} else if t == Minor {
		newver, err = version.BumpMinorVersion("", "")
	} else {
		newver, err = version.BumpPatchVersion("", "")
	}
	if err != nil {
		return nil, err
	}

	funcfile.Version = newver.String()
	return funcfile, nil
}

// cleanImageName is intended to remove any trailing tag from the image name
// since the version field conveys this information. More cleanup could be done
// here in future if necessary.
func cleanImageName(name string) string {
	slashParts := strings.Split(name, "/")
	l := len(slashParts) - 1
	if i := strings.Index(slashParts[l], ":"); i > -1 {
		slashParts[l] = slashParts[l][:i]
	}
	return strings.Join(slashParts, "/")
}

// --------- FuncFileV20180708 -------------

func bumpItWdV20180708(wd string, vtype VType) (*FuncFileV20180708, error) {
	fn, err := FindFuncfile(wd)
	if err != nil {
		return nil, err
	}
	return BumpItV20180708(fn, vtype)
}

// BumpIt returns updated funcfile
func BumpItV20180708(fpath string, vtype VType) (*FuncFileV20180708, error) {
	// fmt.Println("Bumping version in func file at: ", fpath)
	funcfile, err := ParseFuncFileV20180708(fpath)
	if err != nil {
		return nil, err
	}

	funcfile, err = bumpVersionV20180708(funcfile, vtype)
	if err != nil {
		return nil, err
	}

	if err := storeFuncFileV20180708(fpath, funcfile); err != nil {
		return nil, err
	}
	fmt.Println("Bumped to version", funcfile.Version)
	return funcfile, nil
}

func bumpVersionV20180708(funcfile *FuncFileV20180708, t VType) (*FuncFileV20180708, error) {
	funcfile.Name = cleanImageName(funcfile.Name)
	if funcfile.Version == "" {
		funcfile.Version = InitialVersion
		return funcfile, nil
	}

	s, err := storage.NewVersionStorage("local", funcfile.Version)
	if err != nil {
		return nil, err
	}

	version := bumper.NewSemverBumper(s, "")
	var newver *semver.Version
	if t == Major {
		newver, err = version.BumpMajorVersion("", "")
	} else if t == Minor {
		newver, err = version.BumpMinorVersion("", "")
	} else {
		newver, err = version.BumpPatchVersion("", "")
	}
	if err != nil {
		return nil, err
	}

	funcfile.Version = newver.String()
	return funcfile, nil
}
