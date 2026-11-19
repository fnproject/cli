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

	errUnexpectedFileFormat = errors.New("Unexpected file format for function file")
)

// AppFile defines the internal structure of a app.yaml/json/yml
type AppFile struct {
	Name        string                 `yaml:"name,omitempty" json:"name,omitempty"`
	Config      map[string]string      `yaml:"config,omitempty" json:"config,omitempty"`
	Annotations map[string]interface{} `yaml:"annotations,omitempty" json:"annotations,omitempty"`
	SyslogURL   string                 `yaml:"syslog_url,omitempty" json:"syslog_url,omitempty"`
}

func findAppfile(path string) (string, error) {
	for _, fn := range validAppfileNames {
		fullfn := filepath.Join(path, fn)
		if Exists(fullfn) {
			return fullfn, nil
		}
	}
	return "", NewNotFoundError("Could not find app file")
}

// LoadAppfile returns a parsed appfile.
func LoadAppfile(path string) (*AppFile, error) {
	fn, err := findAppfile(path)
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
		return nil, fmt.Errorf("Could not open %s for parsing. Error: %v", path, err)
	}
	ff := &AppFile{}
	err = json.NewDecoder(f).Decode(ff)
	return ff, err
}

func decodeAppfileYAML(path string) (*AppFile, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Could not open %s for parsing. Error: %v", path, err)
	}
	ff := &AppFile{}
	err = yaml.Unmarshal(b, ff)
	if err != nil {
		return ff, err
	}
	// yaml.v2 decodes nested maps as map[interface{}]interface{}, which the
	// API clients cannot use (e.g. oracle.com/oci/definedTags), so convert
	// them to map[string]interface{} like app.json would give.
	for k, v := range ff.Annotations {
		ff.Annotations[k] = normalizeYAMLValue(v)
	}
	return ff, nil
}

func normalizeYAMLValue(v interface{}) interface{} {
	switch t := v.(type) {
	case map[interface{}]interface{}:
		m := make(map[string]interface{}, len(t))
		for k, v2 := range t {
			m[fmt.Sprint(k)] = normalizeYAMLValue(v2)
		}
		return m
	case []interface{}:
		out := make([]interface{}, len(t))
		for i, item := range t {
			out[i] = normalizeYAMLValue(item)
		}
		return out
	default:
		return v
	}
}
