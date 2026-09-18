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
	client "github.com/fnproject/cli/client"
	"github.com/fnproject/cli/common"
	apps "github.com/fnproject/cli/objects/app"
	"github.com/urfave/cli"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
		cli.StringFlag{
			Name:  "app",
			Usage: "App name used to resolve target application shape for code-only builds.",
		},
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
			appName := strings.TrimSpace(c.String("app"))
			if appName == "" {
				return fmt.Errorf("code-only build requires --app so the target application shape can be used for packaging")
			}
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			app, err := apps.GetAppByName(provider.APIClientv2(), appName)
			if err != nil {
				return err
			}
			shape := app.Shape
			if shape == "" {
				shape = common.DefaultAppShape
			}
			archivePath, err := buildCodeOnlyArchive(dir, ff, shape)
			if err != nil {
				return err
			}
			fmt.Printf("Code-only function packaged successfully: %s\n", archivePath)
			return nil
		}

		buildArgs := c.StringSlice("build-arg")

		// Passing empty shape for build command
		ff, err = common.BuildFuncV20180708(common.IsVerbose(), fpath, ff, buildArgs, b.noCache, "", false)
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

func buildCodeOnlyArchive(dir string, ff *common.FuncFileV20180708, shape string) (string, error) {
	if err := validateCodeOnlyBuildTooling(dir, ff); err != nil {
		return "", err
	}
	archivePath := filepath.Join(dir, fmt.Sprintf("%s.%s.zip", ff.Name, ff.Version))
	if err := createCodeOnlyZipArchive(dir, archivePath, ff, shape); err != nil {
		return "", err
	}
	return archivePath, nil
}

func validateCodeOnlyBuildTooling(dir string, ff *common.FuncFileV20180708) error {
	baseRuntime := detectCodeOnlyBaseRuntime(dir, ff)
	switch {
	case strings.HasPrefix(baseRuntime, "java"):
		if _, err := exec.LookPath("mvn"); err != nil {
			return fmt.Errorf("%s runtime selected, but Maven was not found in PATH. Install Maven and rerun `fn build`, or choose a different runtime", buildRuntimeDisplayName(baseRuntime))
		}
	case strings.HasPrefix(baseRuntime, "python"):
		if _, err := findFirstTool("python3", "python"); err != nil {
			return fmt.Errorf("Python runtime selected, but Python was not found in PATH. Install Python and rerun `fn build`, or choose a different runtime")
		}
	case strings.HasPrefix(baseRuntime, "go"):
		if _, err := findFirstTool("go"); err != nil {
			return fmt.Errorf("Go runtime selected, but Go was not found in PATH. Install Go and rerun `fn build`, or choose a different runtime")
		}
	case strings.HasPrefix(baseRuntime, "node"):
		if _, err := findFirstTool("node"); err != nil {
			return fmt.Errorf("Node.js runtime selected, but Node.js was not found in PATH. Install Node.js and rerun `fn build`, or choose a different runtime")
		}
	}
	return nil
}

func createCodeOnlyZipArchive(dir, archivePath string, ff *common.FuncFileV20180708, shape string) error {
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
	if err := addCodeOnlyArchiveContents(zipWriter, dir, paths, ff, shape); err != nil {
		return err
	}

	return zipWriter.Close()
}


func addCodeOnlyArchiveContents(zipWriter *zip.Writer, dir string, paths []string, ff *common.FuncFileV20180708, shape string) error {
	baseRuntime := detectCodeOnlyBaseRuntime(dir, ff)
	if strings.HasPrefix(baseRuntime, "go") {
		return addGoCodeOnlyArchiveContents(zipWriter, dir, ff, shape)
	}
	if strings.HasPrefix(baseRuntime, "java") {
		return addJavaCodeOnlyArchiveContents(zipWriter, dir, ff, shape)
	}
	if strings.HasPrefix(baseRuntime, "python") {
		return addPythonCodeOnlyArchiveContents(zipWriter, dir, ff, shape)
	}
	if strings.HasPrefix(baseRuntime, "node") {
		return addNodeCodeOnlyArchiveContents(zipWriter, dir, ff, shape)
	}
	for _, relPath := range paths {
		fullPath := filepath.Join(dir, relPath)
		archiveRelPath := codeOnlyArchivePath(relPath, ff)
		if err := addFileToZip(zipWriter, fullPath, archiveRelPath); err != nil {
			return err
		}
	}
	return nil
}

func detectCodeOnlyBaseRuntime(dir string, ff *common.FuncFileV20180708) string {
	isKnown := func(base string) bool {
		return base == "go" || base == "java" || base == "python" || base == "node"
	}
	if ff != nil {
		if strings.TrimSpace(ff.Runtime) != "" {
			if base := codeOnlyBaseRuntime(strings.TrimSpace(ff.Runtime)); base != "" {
				if isKnown(base) {
					return base
				}
			}
		}
		if ff.Runtime_config != nil && strings.TrimSpace(ff.Runtime_config.Runtime_name) != "" {
			if base := codeOnlyBaseRuntime(strings.TrimSpace(ff.Runtime_config.Runtime_name)); base != "" {
				if isKnown(base) {
					return base
				}
			}
		}
	}

	if common.Exists(filepath.Join(dir, "go.mod")) {
		return "go"
	}
	if common.Exists(filepath.Join(dir, "pom.xml")) || common.Exists(filepath.Join(dir, "build.gradle")) || common.Exists(filepath.Join(dir, "build.gradle.kts")) {
		return "java"
	}
	if common.Exists(filepath.Join(dir, "function", "func.js")) || common.Exists(filepath.Join(dir, "package.json")) {
		return "node"
	}
	if common.Exists(filepath.Join(dir, "function", "hello_world.py")) {
		return "python"
	}
	return ""
}


func addNodeCodeOnlyArchiveContents(zipWriter *zip.Writer, dir string, ff *common.FuncFileV20180708, shape string) error {
	functionDir := filepath.Join(dir, "function")
	if !common.Exists(functionDir) {
		return fmt.Errorf("node.js code-only build requires a function/ directory at the archive root")
	}
	packageJSON := filepath.Join(dir, "package.json")
	nodeModulesDir := filepath.Join(dir, "node_modules")
	if common.Exists(packageJSON) && !common.Exists(nodeModulesDir) {
		npmBin, err := findFirstTool("npm")
		if err != nil {
			return fmt.Errorf("node.js code-only build requires npm to install @fnproject/fdk dependencies")
		}
		npmInstall := exec.Command(npmBin, "install", "--omit=dev")
		npmInstall.Dir = dir
		npmInstall.Stdout = os.Stdout
		npmInstall.Stderr = os.Stderr
		if err := npmInstall.Run(); err != nil {
			return err
		}
	}
	if err := addDirectoryToZip(zipWriter, functionDir, "function"); err != nil {
		return err
	}
	if common.Exists(nodeModulesDir) {
		if err := addDirectoryToZip(zipWriter, nodeModulesDir, "node_modules"); err != nil {
			return err
		}
	}
	if common.Exists(packageJSON) {
		if err := addFileToZip(zipWriter, packageJSON, "package.json"); err != nil {
			return err
		}
	}
	resourcesDir := filepath.Join(dir, "resources")
	if common.Exists(resourcesDir) {
		if err := addDirectoryToZip(zipWriter, resourcesDir, "resources"); err != nil {
			return err
		}
	}
	nativeDir := filepath.Join(dir, "native")
	if common.Exists(nativeDir) {
		if err := addNodeNativeArchiveContents(zipWriter, nativeDir, shape); err != nil {
			return err
		}
	}
	return nil
}

func addNodeNativeArchiveContents(zipWriter *zip.Writer, nativeDir, shape string) error {
	entries, err := os.ReadDir(nativeDir)
	if err != nil {
		return err
	}
	valid := map[string]bool{"fn-arch-x86": true, "fn-arch-arm": true}
	found := map[string]bool{}
	for _, entry := range entries {
		if !entry.IsDir() {
			return fmt.Errorf("native/ must not contain files directly at its root")
		}
		if !valid[entry.Name()] {
			return fmt.Errorf("native/ contains unsupported architecture directory %s", entry.Name())
		}
		found[entry.Name()] = true
	}
	arches := codeOnlyGoTargetArchitectures(shape)
	if len(arches) == 1 {
		return fmt.Errorf("native/ is not allowed for single-architecture Node.js functions")
	}
	if !found["fn-arch-x86"] || !found["fn-arch-arm"] {
		return fmt.Errorf("native/ must contain both fn-arch-x86 and fn-arch-arm for a multi-architecture application")
	}
	for name := range found {
		archDir := filepath.Join(nativeDir, name)
		if err := addDirectoryToZip(zipWriter, archDir, filepath.ToSlash(filepath.Join("native", name))); err != nil {
			return err
		}
	}
	return nil
}

func addPythonCodeOnlyArchiveContents(zipWriter *zip.Writer, dir string, ff *common.FuncFileV20180708, shape string) error {
	functionDir := filepath.Join(dir, "function")
	if !common.Exists(functionDir) {
		return fmt.Errorf("python code-only build requires a function/ directory at the archive root")
	}
	if err := addDirectoryToZip(zipWriter, functionDir, "function"); err != nil {
		return err
	}
	pythonDir := filepath.Join(dir, "python")
	if common.Exists(pythonDir) {
		if err := addDirectoryToZip(zipWriter, pythonDir, "python"); err != nil {
			return err
		}
	}
	resourcesDir := filepath.Join(dir, "resources")
	if common.Exists(resourcesDir) {
		if err := addDirectoryToZip(zipWriter, resourcesDir, "resources"); err != nil {
			return err
		}
	}
	nativeDir := filepath.Join(dir, "native")
	if common.Exists(nativeDir) {
		if err := addPythonNativeArchiveContents(zipWriter, nativeDir, shape); err != nil {
			return err
		}
	}
	return nil
}

func addPythonNativeArchiveContents(zipWriter *zip.Writer, nativeDir, shape string) error {
	entries, err := os.ReadDir(nativeDir)
	if err != nil {
		return err
	}
	valid := map[string]bool{"fn-arch-x86": true, "fn-arch-arm": true}
	found := map[string]bool{}
	for _, entry := range entries {
		if !entry.IsDir() {
			return fmt.Errorf("native/ must not contain files directly at its root")
		}
		if !valid[entry.Name()] {
			return fmt.Errorf("native/ contains unsupported architecture directory %s", entry.Name())
		}
		found[entry.Name()] = true
	}
	arches := codeOnlyGoTargetArchitectures(shape)
	if len(arches) == 1 {
		return fmt.Errorf("native/ is not allowed for single-architecture Python functions")
	}
	if !found["fn-arch-x86"] || !found["fn-arch-arm"] {
		return fmt.Errorf("native/ must contain both fn-arch-x86 and fn-arch-arm for a multi-architecture application")
	}
	for name := range found {
		archDir := filepath.Join(nativeDir, name)
		if err := addDirectoryToZip(zipWriter, archDir, filepath.ToSlash(filepath.Join("native", name))); err != nil {
			return err
		}
	}
	return nil
}

func addJavaCodeOnlyArchiveContents(zipWriter *zip.Writer, dir string, ff *common.FuncFileV20180708, shape string) error {
	jarPath, err := locateJavaCodeOnlyJar(dir)
	if err != nil {
		return err
	}
	if err := addFileToZip(zipWriter, jarPath, "main.jar"); err != nil {
		return err
	}
	resourcesDir := filepath.Join(dir, "resources")
	if common.Exists(resourcesDir) {
		if err := addDirectoryToZip(zipWriter, resourcesDir, "resources"); err != nil {
			return err
		}
	}
	nativeDir := filepath.Join(dir, "native")
	if common.Exists(nativeDir) {
		if err := addJavaNativeArchiveContents(zipWriter, nativeDir, shape); err != nil {
			return err
		}
	}
	return nil
}

func locateJavaCodeOnlyJar(dir string) (string, error) {
	var jars []string
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.EqualFold(filepath.Ext(name), ".jar") {
			jars = append(jars, filepath.Join(dir, name))
		}
	}
	if len(jars) == 0 {
		for _, pattern := range []string{"target/*.jar", "build/libs/*.jar"} {
			matches, _ := filepath.Glob(filepath.Join(dir, pattern))
			for _, match := range matches {
				if strings.EqualFold(filepath.Ext(match), ".jar") {
					jars = append(jars, match)
				}
			}
		}
	}
	unique := make([]string, 0, len(jars))
	seen := map[string]struct{}{}
	for _, jar := range jars {
		if _, ok := seen[jar]; ok {
			continue
		}
		seen[jar] = struct{}{}
		unique = append(unique, jar)
	}
	jars = unique
	if len(jars) == 0 {
		return "", fmt.Errorf("java code-only build requires exactly one .jar file at the archive root or a single build output under target/ or build/libs")
	}
	if len(jars) > 1 {
		return "", fmt.Errorf("java code-only build requires exactly one .jar file, found %d", len(jars))
	}
	return jars[0], nil
}

func addJavaNativeArchiveContents(zipWriter *zip.Writer, nativeDir, shape string) error {
	entries, err := os.ReadDir(nativeDir)
	if err != nil {
		return err
	}
	valid := map[string]bool{"fn-arch-x86": true, "fn-arch-arm": true}
	found := map[string]bool{}
	for _, entry := range entries {
		if !entry.IsDir() {
			return fmt.Errorf("native/ must not contain files directly at its root")
		}
		if !valid[entry.Name()] {
			return fmt.Errorf("native/ contains unsupported architecture directory %s", entry.Name())
		}
		found[entry.Name()] = true
	}
	arches := codeOnlyGoTargetArchitectures(shape)
	if len(arches) == 1 {
		required := goCodeOnlyArchiveSegment(arches[0])
		if !found[required] {
			return fmt.Errorf("native/ must contain %s for the target application shape", required)
		}
		for name := range found {
			if name != required {
				return fmt.Errorf("native/ must contain only %s for the target application shape", required)
			}
		}
	} else {
		if !found["fn-arch-x86"] || !found["fn-arch-arm"] {
			return fmt.Errorf("native/ must contain both fn-arch-x86 and fn-arch-arm for a multi-architecture application")
		}
	}
	for name := range found {
		archDir := filepath.Join(nativeDir, name)
		if err := addDirectoryToZip(zipWriter, archDir, filepath.ToSlash(filepath.Join("native", name))); err != nil {
			return err
		}
	}
	return nil
}

func addGoCodeOnlyArchiveContents(zipWriter *zip.Writer, dir string, ff *common.FuncFileV20180708, shape string) error {
	architectures := codeOnlyGoTargetArchitectures(shape)
	if len(architectures) == 1 {
		binaryPath := filepath.Join(dir, "func")
		if err := buildGoCodeOnlyBinary(dir, binaryPath, architectures[0]); err != nil {
			return err
		}
		if err := addFileToZip(zipWriter, binaryPath, "func"); err != nil {
			return err
		}
		resourcesDir := filepath.Join(dir, "resources")
		if common.Exists(resourcesDir) {
			if err := addDirectoryToZip(zipWriter, resourcesDir, "resources"); err != nil {
				return err
			}
		}
		return nil
	}
	for _, arch := range architectures {
		segment := goCodeOnlyArchiveSegment(arch)
		archDir := filepath.Join(dir, segment)
		if err := os.MkdirAll(archDir, 0755); err != nil {
			return err
		}
		binaryPath := filepath.Join(archDir, "func")
		if err := buildGoCodeOnlyBinary(dir, binaryPath, arch); err != nil {
			return err
		}
		if err := addFileToZip(zipWriter, binaryPath, filepath.ToSlash(filepath.Join(segment, "func"))); err != nil {
			return err
		}
		resourcesDir := filepath.Join(dir, "resources")
		if common.Exists(resourcesDir) {
			if err := addDirectoryToZip(zipWriter, resourcesDir, filepath.ToSlash(filepath.Join(segment, "resources"))); err != nil {
				return err
			}
		}
	}
	return nil
}

func codeOnlyGoTargetArchitectures(shape string) []string {
	if shape == "" {
		switch runtime.GOARCH {
		case "arm64":
			return []string{"arm64"}
		default:
			return []string{"amd64"}
		}
	}
	if archs, ok := common.TargetPlatformMap[shape]; ok && len(archs) > 0 {
		parts := strings.Split(archs[0], "_")
		return parts
	}
	return []string{"amd64"}
}

func goCodeOnlyArchiveSegment(arch string) string {
	if arch == "arm64" {
		return "fn-arch-arm"
	}
	return "fn-arch-x86"
}

func buildGoCodeOnlyBinary(dir, outputPath, arch string) error {
	goBin := resolveGoBinary()
	env := withEnvOverrides(os.Environ(), map[string]string{
		"GOOS":       "linux",
		"GOARCH":     arch,
		"GOFLAGS":    "-mod=mod",
		"GOTOOLCHAIN": "go1.24.0+auto",
	})

	modDownload := exec.Command(goBin, "mod", "download")
	modDownload.Dir = dir
	modDownload.Env = env
	modDownload.Stdout = os.Stdout
	modDownload.Stderr = os.Stderr
	if err := modDownload.Run(); err != nil {
		return err
	}

	cmd := exec.Command(goBin, "build", "-trimpath", "-ldflags", "-s -w", "-o", outputPath, ".")
	cmd.Dir = dir
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func withEnvOverrides(base []string, overrides map[string]string) []string {
	filtered := make([]string, 0, len(base)+len(overrides))
	for _, kv := range base {
		i := strings.Index(kv, "=")
		if i <= 0 {
			filtered = append(filtered, kv)
			continue
		}
		k := kv[:i]
		if _, drop := overrides[k]; drop {
			continue
		}
		filtered = append(filtered, kv)
	}
	for k, v := range overrides {
		filtered = append(filtered, k+"="+v)
	}
	return filtered
}

func resolveGoBinary() string {
	if override := strings.TrimSpace(os.Getenv("FN_GO_BIN")); override != "" {
		return override
	}
	preferred := "/usr/local/go/bin/go"
	if st, err := os.Stat(preferred); err == nil && !st.IsDir() {
		return preferred
	}
	return "go"
}

func addDirectoryToZip(zipWriter *zip.Writer, sourceDir, archivePrefix string) error {
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		return addFileToZip(zipWriter, path, filepath.ToSlash(filepath.Join(archivePrefix, rel)))
	})
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
	if strings.EqualFold(filepath.Ext(base), ".zip") {
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
	if lower == "ol9" || lower == "ol8" {
		// Managed Go runtimes may be represented by generic OS names.
		return "go"
	}
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
	case strings.HasPrefix(runtime, "python"):
		return "Python"
	case strings.HasPrefix(runtime, "go"):
		return "Go"
	case strings.HasPrefix(runtime, "node"):
		return "Node.js"
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
