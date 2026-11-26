package commands

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fnproject/cli/common"
)

func useFakeTool(t *testing.T, name, script string) {
	useFakeTools(t, map[string]string{name: script})
}

func useFakeTools(t *testing.T, scripts map[string]string) {
	t.Helper()
	binDir := t.TempDir()
	for name, script := range scripts {
		toolPath := filepath.Join(binDir, name)
		if err := os.WriteFile(toolPath, []byte("#!/bin/sh\n"+script+"\n"), 0755); err != nil {
			t.Fatalf("write fake %s executable: %v", name, err)
		}
	}
	t.Setenv("PATH", binDir+":/usr/bin:/bin")
}

func useFakeMaven(t *testing.T, script string) {
	useFakeTool(t, "mvn", script)
}

func TestBuildCodeOnlyArchiveRunsMavenPackageForJava(t *testing.T) {
	dir := t.TempDir()
	useFakeMaven(t, `
[ "$1" = "package" ] || exit 1
printf '%s' "$1" > maven-argument
mkdir -p target
printf 'jar-from-maven' > target/function.jar`)

	archivePath, err := buildCodeOnlyArchive(dir, &common.FuncFileV20180708{
		Name:    "example",
		Version: "0.0.1",
		Runtime: "java",
	}, "GENERIC_X86")
	if err != nil {
		t.Fatalf("buildCodeOnlyArchive returned error: %v", err)
	}

	argument, err := os.ReadFile(filepath.Join(dir, "maven-argument"))
	if err != nil {
		t.Fatalf("read Maven argument: %v", err)
	}
	if string(argument) != "package" {
		t.Fatalf("Maven argument = %q, want package", argument)
	}

	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		t.Fatalf("open archive: %v", err)
	}
	defer reader.Close()
	if len(reader.File) != 1 || reader.File[0].Name != "main.jar" {
		t.Fatalf("archive entries = %v, want only main.jar", reader.File)
	}
}

func TestBuildCodeOnlyArchiveReturnsMavenPackageFailure(t *testing.T) {
	dir := t.TempDir()
	useFakeMaven(t, "exit 42")

	_, err := buildCodeOnlyArchive(dir, &common.FuncFileV20180708{
		Name:    "example",
		Version: "0.0.1",
		Runtime: "java",
	}, "GENERIC_X86")
	if err == nil || !strings.Contains(err.Error(), "java code-only build failed while running `mvn package`") {
		t.Fatalf("error = %v, want Maven package failure", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "example.0.0.1.zip")); !os.IsNotExist(err) {
		t.Fatalf("archive exists after Maven failure; stat error = %v", err)
	}
}

func TestBuildCodeOnlyArchiveInstallsPythonRequirements(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "function"), 0755); err != nil {
		t.Fatalf("create function directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "function", "hello_world.py"), []byte("def handler(): pass"), 0600); err != nil {
		t.Fatalf("write function: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "requirements.txt"), []byte("example==1.0"), 0600); err != nil {
		t.Fatalf("write requirements: %v", err)
	}
	useFakeTool(t, "python3", `
[ "$1" = "-m" ] && [ "$2" = "pip" ] && [ "$3" = "install" ] || exit 1
mkdir -p python
printf 'installed' > python/dependency.py`)

	archivePath, err := buildCodeOnlyArchive(dir, &common.FuncFileV20180708{
		Name:    "example",
		Version: "0.0.1",
		Runtime: "python312.ol9",
	}, "GENERIC_X86")
	if err != nil {
		t.Fatalf("buildCodeOnlyArchive returned error: %v", err)
	}

	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		t.Fatalf("open archive: %v", err)
	}
	defer reader.Close()
	entries := make([]string, 0, len(reader.File))
	for _, file := range reader.File {
		entries = append(entries, file.Name)
	}
	if !containsString(entries, "python/dependency.py") {
		t.Fatalf("archive entries = %v, want installed Python dependency", entries)
	}
}

func TestBuildCodeOnlyArchiveInstallsNodeDependencies(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "function"), 0755); err != nil {
		t.Fatalf("create function directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "function", "func.js"), []byte("module.exports = () => {};"), 0600); err != nil {
		t.Fatalf("write function: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"dependencies":{}}`), 0600); err != nil {
		t.Fatalf("write package.json: %v", err)
	}
	useFakeTools(t, map[string]string{
		"node": "exit 0",
		"npm": `
[ "$1" = "install" ] && [ "$2" = "--omit=dev" ] || exit 1
mkdir -p node_modules/example
printf 'installed' > node_modules/example/index.js`,
	})

	archivePath, err := buildCodeOnlyArchive(dir, &common.FuncFileV20180708{
		Name:    "example",
		Version: "0.0.1",
		Runtime: "node24.ol9",
	}, "GENERIC_X86")
	if err != nil {
		t.Fatalf("buildCodeOnlyArchive returned error: %v", err)
	}

	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		t.Fatalf("open archive: %v", err)
	}
	defer reader.Close()
	entries := make([]string, 0, len(reader.File))
	for _, file := range reader.File {
		entries = append(entries, file.Name)
	}
	if !containsString(entries, "node_modules/example/index.js") {
		t.Fatalf("archive entries = %v, want installed Node dependency", entries)
	}
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func TestCreateCodeOnlyZipArchiveDoesNotLeaveEmptyArchiveOnFailure(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "example.0.0.1.zip")
	existingArchive := []byte("existing archive contents")
	if err := os.WriteFile(archivePath, existingArchive, 0600); err != nil {
		t.Fatalf("write existing archive: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "pom.xml"), []byte("<project/>"), 0600); err != nil {
		t.Fatalf("write pom.xml: %v", err)
	}

	err := createCodeOnlyZipArchive(dir, archivePath, &common.FuncFileV20180708{
		Name:    "example",
		Version: "0.0.1",
		Runtime: "java21.ol9",
	}, "GENERIC_X86")
	if err == nil {
		t.Fatal("createCodeOnlyZipArchive succeeded, want missing JAR error")
	}

	got, err := os.ReadFile(archivePath)
	if err != nil {
		t.Fatalf("read existing archive after failure: %v", err)
	}
	if !bytes.Equal(got, existingArchive) {
		t.Fatalf("archive contents = %q, want existing contents %q", got, existingArchive)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read directory: %v", err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "."+filepath.Base(archivePath)+"-") {
			t.Fatalf("temporary archive %q was not removed", entry.Name())
		}
	}
}

func TestCreateCodeOnlyZipArchiveDoesNotCreateArchiveOnFailure(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "example.0.0.1.zip")
	if err := os.WriteFile(filepath.Join(dir, "pom.xml"), []byte("<project/>"), 0600); err != nil {
		t.Fatalf("write pom.xml: %v", err)
	}

	err := createCodeOnlyZipArchive(dir, archivePath, &common.FuncFileV20180708{
		Name:    "example",
		Version: "0.0.1",
		Runtime: "java21.ol9",
	}, "GENERIC_X86")
	if err == nil {
		t.Fatal("createCodeOnlyZipArchive succeeded, want missing JAR error")
	}
	if _, err := os.Stat(archivePath); !os.IsNotExist(err) {
		t.Fatalf("archive exists after failed packaging; stat error = %v", err)
	}
}
