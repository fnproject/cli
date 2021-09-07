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

package config

import (
	"os"
	"path"
	"testing"
)

type testCase struct {
	name     string
	file     string
	contents string
	expected *ContextFile
}

func TestContextFile(t *testing.T) {
	folder := path.Join(os.TempDir(), "fn-tests")
	err := os.Mkdir(folder, 0755)
	if err != nil {
		t.Fatalf("failed to create test folder %s: %v", folder, err)
	}
	defer cleanup(folder)

	tests, err := prepareTestFiles(folder)
	if err != nil {
		t.Fatalf("failed to prepare test files in %s", folder)
	}

	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			actual, err := NewContextFile(tst.file)
			if err != nil {
				t.Fatalf("failed to create a context file: %v", err)
			}

			if actual.ContextProvider != tst.expected.ContextProvider {
				t.Fatalf("ContextProvider: expected '%s', but got '%s'", tst.expected.ContextProvider, actual.ContextProvider)
			}

			if actual.EnvFnAPIURL != tst.expected.EnvFnAPIURL {
				t.Fatalf("EnvFnAPIURL: expected '%s', but got '%s'", tst.expected.EnvFnAPIURL, actual.EnvFnAPIURL)
			}

			if actual.EnvFnRegistry != tst.expected.EnvFnRegistry {
				t.Fatalf("EnvFnRegistry: expected '%s', but got '%s'", tst.expected.EnvFnRegistry, actual.EnvFnRegistry)
			}
		})
	}
}

// prepareTestFiles creates YAML files in a temporary test folder
func prepareTestFiles(folder string) ([]testCase, error) {
	var tests = []testCase{
		{
			name: "Simple context file",
			file: path.Join(folder, "default.yaml"),
			contents: `
api-url: http://localhost:8080
provider: default
registry: "someregistry"`,
			expected: &ContextFile{
				ContextProvider: "default",
				EnvFnAPIURL:     "http://localhost:8080",
				EnvFnRegistry:   "someregistry",
			},
		},
	}

	for _, tf := range tests {
		f, err := os.Create(tf.file)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		f.WriteString(tf.contents)
	}
	return tests, nil
}

// cleanup removes the temporary folder
func cleanup(folder string) error {
	return os.RemoveAll(folder)
}
