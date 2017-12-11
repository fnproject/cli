package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// WalkFuncsFunc good name huh?
type walkFuncsFunc func(path string, ff *funcfile, err error) error

// walkFuncs is similar to filepath.Walk except only returns func.yaml's (so on per function)
func walkFuncs(root string, walkFn walkFuncsFunc) error {
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

		if !isFuncfile(path, info) {
			return nil
		}

		// TODO: test/try this again to speed up deploys.
		if false && !isstale(path) {
			return nil
		}
		// Then we found a func file, so let's deploy it:
		ff, err := parseFuncfile(path)
		// if err != nil {
		// return err
		// }
		return walkFn(path, ff, err)
	})
}
