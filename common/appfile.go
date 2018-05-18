package common

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

type AppFile struct {
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	// TODO: Config here is not yet used
	Config map[string]string `yaml:"config,omitempty" json:"config,omitempty"`
}

func findAppfile(path string) (string, error) {
	for _, fn := range validAppfileNames {
		fullfn := filepath.Join(path, fn)
		if Exists(fullfn) {
			return fullfn, nil
		}
	}
	return "", NewNotFoundError("could not find app file")
}

func LoadAppfile() (*AppFile, error) {
	fn, err := findAppfile(".")
	if err != nil {
		return nil, err
	}
	return parseAppfile(fn)
}

func parseAppfile(path string) (*AppFile, error) {
	ext := filepath.Ext(path)
	switch ext {
	case ".json":
		return decodeAppfileJSON(path)
	case ".yaml", ".yml":
		return decodeAppfileYAML(path)
	}
	return nil, errUnexpectedFileFormat
}

func decodeAppfileJSON(path string) (*AppFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open %s for parsing. Error: %v", path, err)
	}
	ff := &AppFile{}
	err = json.NewDecoder(f).Decode(ff)
	return ff, err
}

func decodeAppfileYAML(path string) (*AppFile, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not open %s for parsing. Error: %v", path, err)
	}
	ff := &AppFile{}
	err = yaml.Unmarshal(b, ff)
	return ff, err
}
