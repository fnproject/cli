package test

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fnproject/cli/common"
	"github.com/fnproject/cli/testharness"
)

func archivePathFromBuildOutput(t *testing.T, stdout string) string {
	t.Helper()
	const prefix = "Code-only function packaged successfully: "
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, prefix))
		}
	}
	t.Fatalf("could not find archive path in build output: %q", stdout)
	return ""
}

func zipEntryNames(t *testing.T, archivePath string) []string {
	t.Helper()
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		t.Fatalf("failed to open zip archive %s: %v", archivePath, err)
	}
	defer reader.Close()

	entries := make([]string, 0, len(reader.File))
	for _, f := range reader.File {
		entries = append(entries, f.Name)
	}
	return entries
}

func containsEntry(entries []string, target string) bool {
	for _, entry := range entries {
		if entry == target {
			return true
		}
	}
	return false
}

func TestCodeOnlyBuild(t *testing.T) {
	t.Run("python code-only build should create a versioned archive with function root and exclude func.yaml", func(t *testing.T) {
		t.Parallel()
		h := testharness.Create(t)
		defer h.Cleanup()

		appName := h.NewAppName()
		funcName := h.NewFuncName(appName)
		dirName := funcName + "_dir"
		h.Fn("init", "--code-only", "--runtime-name", "python311.ol9", "--runtime-config-type", "function-update", "--name", funcName, dirName).AssertSuccess()

		h.Cd(dirName)
		res := h.Fn("build").AssertSuccess().AssertStdoutContains("Code-only function packaged successfully:")
		archivePath := archivePathFromBuildOutput(t, res.Stdout)

		if _, err := os.Stat(archivePath); err != nil {
			t.Fatalf("expected archive at %s: %v", archivePath, err)
		}
		expectedArchiveName := fmt.Sprintf("%s.0.0.1.zip", funcName)
		if filepath.Base(archivePath) != expectedArchiveName {
			t.Fatalf("archive name was %q, expected %q", filepath.Base(archivePath), expectedArchiveName)
		}

		entries := zipEntryNames(t, archivePath)
		if !containsEntry(entries, "function/hello_world.py") {
			t.Fatalf("expected function/hello_world.py in archive, got entries: %v", entries)
		}
		if containsEntry(entries, "func.yaml") {
			t.Fatalf("func.yaml should not be included in code-only archive, entries: %v", entries)
		}
	})

	t.Run("go code-only build should create a versioned archive and exclude func.yaml", func(t *testing.T) {
		t.Parallel()
		h := testharness.Create(t)
		defer h.Cleanup()

		appName := h.NewAppName()
		funcName := h.NewFuncName(appName)
		dirName := funcName + "_dir"
		h.Fn("init", "--code-only", "--runtime", "go", "--runtime-config-type", "function-update", "--name", funcName, dirName).AssertSuccess()

		h.Cd(dirName)
		res := h.Fn("build").AssertSuccess().AssertStdoutContains("Code-only function packaged successfully:")
		archivePath := archivePathFromBuildOutput(t, res.Stdout)

		if _, err := os.Stat(archivePath); err != nil {
			t.Fatalf("expected archive at %s: %v", archivePath, err)
		}
		expectedArchiveName := fmt.Sprintf("%s.0.0.1.zip", funcName)
		if filepath.Base(archivePath) != expectedArchiveName {
			t.Fatalf("archive name was %q, expected %q", filepath.Base(archivePath), expectedArchiveName)
		}

		entries := zipEntryNames(t, archivePath)
		if !containsEntry(entries, "func.go") {
			t.Fatalf("expected func.go in archive, got entries: %v", entries)
		}
		if containsEntry(entries, "func.yaml") {
			t.Fatalf("func.yaml should not be included in code-only archive, entries: %v", entries)
		}
	})

	t.Run("java code-only build should require Maven", func(t *testing.T) {
		t.Parallel()
		h := testharness.Create(t)
		defer h.Cleanup()

		h.MkDir("hello-java")
		h.Cd("hello-java")
		h.WithFile("func.yaml", fmt.Sprintf(`schema_version: %d
name: hello-java
version: 0.0.1
code_only: true
runtime_config:
  type: function-update
  runtime_name: java
handler: com.example.fn.HelloFunction::handleRequest
`, common.LatestYamlVersion), 0644)
		h.WithEnv("PATH", "/usr/bin:/bin")

		h.Fn("build").AssertFailed().AssertStderrContains("Maven was not found in PATH")
	})
}