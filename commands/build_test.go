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

func useFakeMaven(t *testing.T, script string) {
	t.Helper()
	binDir := t.TempDir()
	mvnPath := filepath.Join(binDir, "mvn")
	if err := os.WriteFile(mvnPath, []byte("#!/bin/sh\n"+script+"\n"), 0755); err != nil {
		t.Fatalf("write fake Maven executable: %v", err)
	}
	t.Setenv("PATH", binDir+":/usr/bin:/bin")
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
