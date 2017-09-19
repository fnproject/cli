package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

var (
	validAppfileNames = [...]string{
		"app.yaml",
		"app.yml",
		"app.json",
	}

	errUnexpectedFileFormat = errors.New("unexpected file format for function file")
)

type appfile struct {
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	// TODO: Config here is not yet used
	Config map[string]string `yaml:"config,omitempty" json:"config,omitempty"`
}

func findAppfile(path string) (string, error) {
	for _, fn := range validAppfileNames {
		fullfn := filepath.Join(path, fn)
		if exists(fullfn) {
			return fullfn, nil
		}
	}
	return "", newNotFoundError("could not find app file")
}

func loadAppfile() (*appfile, error) {
	fn, err := findAppfile(".")
	if err != nil {
		return nil, err
	}
	return parseAppfile(fn)
}

func parseAppfile(path string) (*appfile, error) {
	ext := filepath.Ext(path)
	switch ext {
	case ".json":
		return decodeAppfileJSON(path)
	case ".yaml", ".yml":
		return decodeAppfileYAML(path)
	}
	return nil, errUnexpectedFileFormat
}

func decodeAppfileJSON(path string) (*appfile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open %s for parsing. Error: %v", path, err)
	}
	ff := &appfile{}
	err = json.NewDecoder(f).Decode(ff)
	return ff, err
}

func decodeAppfileYAML(path string) (*appfile, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not open %s for parsing. Error: %v", path, err)
	}
	ff := &appfile{}
	err = yaml.Unmarshal(b, ff)
	return ff, err
}
