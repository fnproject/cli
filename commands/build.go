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

package commands

import (
	"archive/zip"
	"fmt"
	"github.com/fnproject/cli/common"
	"github.com/urfave/cli"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// BuildCommand returns build cli.command
func BuildCommand() cli.Command {
	cmd := buildcmd{}
	flags := append([]cli.Flag{}, cmd.flags()...)
	return cli.Command{
		Name:        "build",
		Usage:       "\tBuild function version",
		Category:    "DEVELOPMENT COMMANDS",
		Description: "This command builds a new function.",
		ArgsUsage:   "[function-subdirectory]",
		Aliases:     []string{"bu"},
		Flags:       flags,
		Action:      cmd.build,
	}
}

type buildcmd struct {
	noCache bool
}

func (b *buildcmd) flags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:        "verbose, v",
			Usage:       "Verbose mode",
			Destination: &common.CommandVerbose,
		},
		cli.BoolFlag{
			Name:        "no-cache",
			Usage:       "Don't use docker cache",
			Destination: &b.noCache,
		},
		cli.StringSliceFlag{
			Name:  "build-arg",
			Usage: "Set build-time variables",
		},
		cli.StringFlag{
			Name:  "working-dir, w",
			Usage: "Specify the working directory to build a function, must be the full path.",
		},
	}
}

// build will take the found valid function and build it
func (b *buildcmd) build(c *cli.Context) error {
	dir := common.GetDir(c)

	path := c.Args().First()
	if path != "" {
		fmt.Printf("Building function at: ./%s\n", path)
		dir = filepath.Join(dir, path)
	}

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

		if ff.Code_only {
			if ff.Runtime == "" && ff.Runtime_config != nil {
				ff.Runtime = ff.Runtime_config.Runtime_name
			}
			archivePath, err := buildCodeOnlyArchive(dir, ff)
			if err != nil {
				return err
			}
			fmt.Printf("Code-only function packaged successfully: %s\n", archivePath)
			return nil
		}

		buildArgs := c.StringSlice("build-arg")

		// Passing empty shape for build command
		ff, err = common.BuildFuncV20180708(common.IsVerbose(), fpath, ff, buildArgs, b.noCache, "")
		if err != nil {
			return err
		}

		fmt.Printf("Function %v built successfully.\n", ff.ImageNameV20180708())
		return nil

	default:
		fpath, ff, err := common.FindAndParseFuncfile(dir)
		if err != nil {
			return err
		}

		buildArgs := c.StringSlice("build-arg")
		ff, err = common.BuildFunc(common.IsVerbose(), fpath, ff, buildArgs, b.noCache)
		if err != nil {
			return err
		}

		fmt.Printf("Function %v built successfully.\n", ff.ImageName())
		return nil
	}
}

func buildCodeOnlyArchive(dir string, ff *common.FuncFileV20180708) (string, error) {
	if err := validateCodeOnlyBuildTooling(ff); err != nil {
		return "", err
	}
	archivePath := filepath.Join(dir, fmt.Sprintf("%s.%s.zip", ff.Name, ff.Version))
	if err := createCodeOnlyZipArchive(dir, archivePath, ff); err != nil {
		return "", err
	}
	return archivePath, nil
}

func validateCodeOnlyBuildTooling(ff *common.FuncFileV20180708) error {
	runtimeName := ""
	if ff.Runtime_config != nil {
		runtimeName = strings.TrimSpace(ff.Runtime_config.Runtime_name)
	}
	baseRuntime := codeOnlyBaseRuntime(runtimeName)
	switch {
	case strings.HasPrefix(baseRuntime, "java"), strings.HasPrefix(baseRuntime, "kotlin"):
		if _, err := exec.LookPath("mvn"); err != nil {
			return fmt.Errorf("%s runtime selected, but Maven was not found in PATH. Install Maven and rerun `fn build`, or choose a different runtime", buildRuntimeDisplayName(baseRuntime))
		}
	case strings.HasPrefix(baseRuntime, "python"):
		if _, err := findFirstTool("python3", "python"); err != nil {
			return fmt.Errorf("Python runtime selected, but Python was not found in PATH. Install Python and rerun `fn build`, or choose a different runtime")
		}
	case strings.HasPrefix(baseRuntime, "ruby"):
		if _, err := findFirstTool("ruby"); err != nil {
			return fmt.Errorf("Ruby runtime selected, but Ruby was not found in PATH. Install Ruby and rerun `fn build`, or choose a different runtime")
		}
	case strings.HasPrefix(baseRuntime, "go"):
		if _, err := findFirstTool("go"); err != nil {
			return fmt.Errorf("Go runtime selected, but Go was not found in PATH. Install Go and rerun `fn build`, or choose a different runtime")
		}
	}
	return nil
}

func createCodeOnlyZipArchive(dir, archivePath string, ff *common.FuncFileV20180708) error {
	if err := os.RemoveAll(archivePath); err != nil {
		return err
	}
	archiveFile, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer archiveFile.Close()

	zipWriter := zip.NewWriter(archiveFile)
	defer zipWriter.Close()

	var paths []string
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == archivePath {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		if rel == "." || shouldSkipCodeOnlyBuildPath(rel) {
			if info.IsDir() {
				return nil
			}
			return nil
		}
		if info.IsDir() {
			return nil
		}
		paths = append(paths, rel)
		return nil
	})
	if err != nil {
		return err
	}

	sort.Strings(paths)
	for _, relPath := range paths {
		fullPath := filepath.Join(dir, relPath)
		archiveRelPath := codeOnlyArchivePath(relPath, ff)
		if err := addFileToZip(zipWriter, fullPath, archiveRelPath); err != nil {
			return err
		}
	}

	return zipWriter.Close()
}

func codeOnlyArchivePath(relPath string, ff *common.FuncFileV20180708) string {
	if ff == nil || ff.Runtime_config == nil {
		return relPath
	}
	baseRuntime := codeOnlyBaseRuntime(strings.TrimSpace(ff.Runtime_config.Runtime_name))
	switch {
	case strings.HasPrefix(baseRuntime, "python"):
		return filepath.ToSlash(filepath.Join("function", relPath))
	default:
		return relPath
	}
}

func shouldSkipCodeOnlyBuildPath(rel string) bool {
	base := filepath.Base(rel)
	if base == ".git" || base == ".idea" || base == ".vscode" || base == ".DS_Store" {
		return true
	}
	if rel == "func.yaml" || rel == "func.yml" || rel == "func.json" {
		return true
	}
	return false
}

func addFileToZip(zipWriter *zip.Writer, fullPath, relPath string) error {
	info, err := os.Stat(fullPath)
	if err != nil {
		return err
	}
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = filepath.ToSlash(relPath)
	header.Method = zip.Deflate
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	file, err := os.Open(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(writer, file)
	return err
}

func codeOnlyBaseRuntime(runtimeName string) string {
	lower := strings.ToLower(strings.TrimSpace(runtimeName))
	for _, sep := range []string{".", "-"} {
		if idx := strings.Index(lower, sep); idx != -1 {
			return lower[:idx]
		}
	}
	return lower
}

func buildRuntimeDisplayName(runtime string) string {
	switch {
	case strings.HasPrefix(runtime, "java"):
		return "Java"
	case strings.HasPrefix(runtime, "kotlin"):
		return "Kotlin"
	case strings.HasPrefix(runtime, "python"):
		return "Python"
	case strings.HasPrefix(runtime, "ruby"):
		return "Ruby"
	case strings.HasPrefix(runtime, "go"):
		return "Go"
	default:
		return runtime
	}
}

func findFirstTool(names ...string) (string, error) {
	for _, name := range names {
		if path, err := exec.LookPath(name); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("tool not found")
}
