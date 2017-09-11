package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/coreos/go-semver/semver"
	bumper "github.com/giantswarm/semver-bump/bump"
	"github.com/giantswarm/semver-bump/storage"
	"github.com/urfave/cli"
)

type VType int

const (
	Patch VType = iota
	Minor
	Major
)

var (
	initialVersion = "0.0.1"
)

func bump() cli.Command {
	cmd := bumpcmd{}
	flags := append([]cli.Flag{}, cmd.flags()...)
	return cli.Command{
		Name:   "bump",
		Usage:  "bump function version",
		Flags:  flags,
		Action: cmd.bump,
	}
}

type bumpcmd struct {
	verbose bool
	major   bool
	minor   bool
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
			Destination: &b.verbose,
		},
	}
}

// bump will take the found valid function and bump its version
func (b *bumpcmd) bump(c *cli.Context) error {

	path, err := os.Getwd()
	if err != nil {
		return err
	}
	fn, err := findFuncfile(path)
	if err != nil {
		return err
	}

	fmt.Println("bumping version in func file at: ", fn)

	funcfile, err := parseFuncfile(fn)
	if err != nil {
		return err
	}

	var t VType
	if b.major {
		t = Major
	} else if b.minor {
		t = Minor
	} else {
		t = Patch
	}
	funcfile, err = bumpVersion(*funcfile, t)
	if err != nil {
		return err
	}

	if err := storeFuncfile(fn, funcfile); err != nil {
		return err
	}

	fmt.Println("Bumped to version", funcfile.Version)
	return nil
}

func bumpVersion(funcfile funcfile, t VType) (*funcfile, error) {
	funcfile.Name = cleanImageName(funcfile.Name)
	if funcfile.Version == "" {
		funcfile.Version = initialVersion
		return &funcfile, nil
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
	return &funcfile, nil
}

func cleanImageName(name string) string {
	if i := strings.Index(name, ":"); i != -1 {
		name = name[:i]
	}

	return name
}
