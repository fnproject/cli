package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fnproject/cli/common"
)

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
