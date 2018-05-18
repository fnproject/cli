package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fnproject/cli/common"
)

// WalkFuncsFunc good name huh?
type walkFuncsFunc func(path string, ff *common.FuncFile, err error) error

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

		if !common.IsFuncFile(path, info) {
			return nil
		}

		// TODO: test/try this again to speed up deploys.
		if false && !isstale(path) {
			return nil
		}
		// Then we found a func file, so let's deploy it:
		ff, err := common.ParseFuncfile(path)
		// if err != nil {
		// return err
		// }
		return walkFn(path, ff, err)
	})
}
