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
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// WalkFuncsFunc good name huh?
type walkFuncsFunc func(path string, ff *FuncFile, err error) error
type walkFuncsFuncV20180708 func(path string, ff *FuncFileV20180708, err error) error

// WalkFuncs is similar to filepath.Walk except only returns func.yaml's (so on per function)
func WalkFuncs(root string, walkFn walkFuncsFunc) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// logging this so we can figure out any common issues
			fmt.Println("error walking filesystem:", err)
			return err
		}
		// was `path != wd` necessary?
		if info.IsDir() {
			return nil
		}

		if !IsFuncFile(path, info) {
			return nil
		}

		// TODO: test/try this again to speed up deploys.
		if false && !isstale(path) {
			return nil
		}
		// Then we found a func file, so let's deploy it:
		ff, err := ParseFuncfile(path)
		// if err != nil {
		// return err
		// }
		return walkFn(path, ff, err)
	})
}

// WalkFuncsV20180708 is similar to filepath.Walk except only returns func.yaml's (so on per function)
func WalkFuncsV20180708(root string, walkFn walkFuncsFuncV20180708) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// logging this so we can figure out any common issues
			fmt.Println("error walking filesystem:", err)
			return err
		}
		// was `path != wd` necessary?
		if info.IsDir() {
			return nil
		}

		if !IsFuncFile(path, info) {
			return nil
		}

		// TODO: test/try this again to speed up deploys.
		if false && !isstale(path) {
			return nil
		}
		// Then we found a func file, so let's deploy it:
		ff, err := ParseFuncFileV20180708(path)
		// if err != nil {
		// return err
		// }
		return walkFn(path, ff, err)
	})
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
