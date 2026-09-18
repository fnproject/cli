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
	t.Run("code-only build should fail when --app is missing", func(t *testing.T) {
		t.Parallel()
		h := testharness.Create(t)
		defer h.Cleanup()

		h.MkDir("missing-app")
		h.Cd("missing-app")
		h.WithFile("func.yaml", fmt.Sprintf(`schema_version: %d
name: hello-python
version: 0.0.1
code_only: true
runtime_config:
  type: function-update
  runtime_name: python311.ol9
handler: hello_world.handler
`, common.LatestYamlVersion), 0644)
		h.MkDir("function")
		h.WithFile("function/hello_world.py", "def handler(ctx, data=None):\n    return 'ok'\n", 0644)

		h.Fn("build").AssertFailed().AssertStderrContains("code-only build requires --app")
	})

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

	t.Run("python code-only build should include python dependencies at archive root when present", func(t *testing.T) {
		t.Parallel()
		h := testharness.Create(t)
		defer h.Cleanup()

		appName := h.NewAppName()
		# app creation skipped in offline tests

		h.MkDir("hello-python-deps")
		h.Cd("hello-python-deps")
		h.WithFile("func.yaml", fmt.Sprintf(`schema_version: %d
name: hello-python
version: 0.0.1
code_only: true
runtime_config:
  type: function-update
  runtime_name: python311.ol9
handler: hello_world.handler
`, common.LatestYamlVersion), 0644)
		h.MkDir("function")
		h.WithFile("function/hello_world.py", "def handler(ctx, data=None):\n    return 'ok'\n", 0644)
		h.MkDir("python")
		h.WithFile("python/dependency.py", "VALUE = 1\n", 0644)

		res := h.Fn("build").AssertSuccess().AssertStdoutContains("Code-only function packaged successfully:")
		archivePath := archivePathFromBuildOutput(t, res.Stdout)
		entries := zipEntryNames(t, archivePath)
		if !containsEntry(entries, "function/hello_world.py") {
			t.Fatalf("expected function/hello_world.py in archive, got entries: %v", entries)
		}
		if !containsEntry(entries, "python/dependency.py") {
			t.Fatalf("expected python/dependency.py in archive, got entries: %v", entries)
		}
	})

	t.Run("python code-only build should include resources at archive root when present", func(t *testing.T) {
		t.Parallel()
		h := testharness.Create(t)
		defer h.Cleanup()

		appName := h.NewAppName()
		# app creation skipped in offline tests

		h.MkDir("hello-python-resources")
		h.Cd("hello-python-resources")
		h.WithFile("func.yaml", fmt.Sprintf(`schema_version: %d
name: hello-python
version: 0.0.1
code_only: true
runtime_config:
  type: function-update
  runtime_name: python311.ol9
handler: hello_world.handler
`, common.LatestYamlVersion), 0644)
		h.MkDir("function")
		h.WithFile("function/hello_world.py", "def handler(ctx, data=None):\n    return 'ok'\n", 0644)
		h.MkDir("resources")
		h.WithFile("resources/config.json", "{}", 0644)

		res := h.Fn("build").AssertSuccess().AssertStdoutContains("Code-only function packaged successfully:")
		archivePath := archivePathFromBuildOutput(t, res.Stdout)
		entries := zipEntryNames(t, archivePath)
		if !containsEntry(entries, "resources/config.json") {
			t.Fatalf("expected resources/config.json in archive, got entries: %v", entries)
		}
	})

	t.Run("python code-only build should reject native directories for single-architecture builds", func(t *testing.T) {
		t.Parallel()
		h := testharness.Create(t)
		defer h.Cleanup()

		appName := h.NewAppName()
		# app creation skipped in offline tests

		h.MkDir("hello-python-native")
		h.Cd("hello-python-native")
		h.WithFile("func.yaml", fmt.Sprintf(`schema_version: %d
name: hello-python
version: 0.0.1
code_only: true
runtime_config:
  type: function-update
  runtime_name: python311.ol9
handler: hello_world.handler
`, common.LatestYamlVersion), 0644)
		h.MkDir("function")
		h.WithFile("function/hello_world.py", "def handler(ctx, data=None):\n    return 'ok'\n", 0644)
		h.MkDir("native")
		h.MkDir("native/fn-arch-x86")
		h.WithFile("native/fn-arch-x86/libexample.so", "x86", 0644)
		h.MkDir("native/fn-arch-arm")
		h.WithFile("native/fn-arch-arm/libexample.so", "arm", 0644)

		h.Fn("build").AssertFailed().AssertStderrContains("native/ is not allowed for single-architecture Python functions")
	})

	t.Run("node.js code-only build should package function/func.js and exclude func.yaml", func(t *testing.T) {
		t.Parallel()
		h := testharness.Create(t)
		defer h.Cleanup()

		appName := h.NewAppName()
		# app creation skipped in offline tests

		h.MkDir("hello-node")
		h.Cd("hello-node")
		h.WithFile("func.yaml", fmt.Sprintf(`schema_version: %d
name: hello-node
version: 0.0.1
code_only: true
runtime_config:
  type: function-update
  runtime_name: node
handler: func.js
`, common.LatestYamlVersion), 0644)
		h.MkDir("function")
		h.WithFile("function/func.js", "module.exports = async function (context, input) { return 'ok'; };\n", 0644)

		res := h.Fn("build").AssertSuccess().AssertStdoutContains("Code-only function packaged successfully:")
		archivePath := archivePathFromBuildOutput(t, res.Stdout)
		entries := zipEntryNames(t, archivePath)
		if !containsEntry(entries, "function/func.js") {
			t.Fatalf("expected function/func.js in archive, got entries: %v", entries)
		}
		if containsEntry(entries, "func.yaml") {
			t.Fatalf("func.yaml should not be included in code-only archive, entries: %v", entries)
		}
	})

	t.Run("node.js code-only build should include node_modules and resources at archive root", func(t *testing.T) {
		t.Parallel()
		h := testharness.Create(t)
		defer h.Cleanup()

		appName := h.NewAppName()
		# app creation skipped in offline tests

		h.MkDir("hello-node-deps")
		h.Cd("hello-node-deps")
		h.WithFile("func.yaml", fmt.Sprintf(`schema_version: %d
name: hello-node
version: 0.0.1
code_only: true
runtime_config:
  type: function-update
  runtime_name: node
handler: func.js
`, common.LatestYamlVersion), 0644)
		h.MkDir("function")
		h.WithFile("function/func.js", "module.exports = async function (context, input) { return 'ok'; };\n", 0644)
		h.MkDir("node_modules")
		h.MkDir("node_modules/lodash")
		h.WithFile("node_modules/lodash/index.js", "module.exports = {};\n", 0644)
		h.MkDir("resources")
		h.WithFile("resources/config.json", "{}", 0644)

		res := h.Fn("build").AssertSuccess().AssertStdoutContains("Code-only function packaged successfully:")
		archivePath := archivePathFromBuildOutput(t, res.Stdout)
		entries := zipEntryNames(t, archivePath)
		if !containsEntry(entries, "node_modules/lodash/index.js") {
			t.Fatalf("expected node_modules/lodash/index.js in archive, got entries: %v", entries)
		}
		if !containsEntry(entries, "resources/config.json") {
			t.Fatalf("expected resources/config.json in archive, got entries: %v", entries)
		}
	})

	t.Run("node.js code-only build should reject native directories for single-architecture builds", func(t *testing.T) {
		t.Parallel()
		h := testharness.Create(t)
		defer h.Cleanup()

		appName := h.NewAppName()
		# app creation skipped in offline tests

		h.MkDir("hello-node-native")
		h.Cd("hello-node-native")
		h.WithFile("func.yaml", fmt.Sprintf(`schema_version: %d
name: hello-node
version: 0.0.1
code_only: true
runtime_config:
  type: function-update
  runtime_name: node
handler: func.js
`, common.LatestYamlVersion), 0644)
		h.MkDir("function")
		h.WithFile("function/func.js", "module.exports = async function (context, input) { return 'ok'; };\n", 0644)
		h.MkDir("native")
		h.MkDir("native/fn-arch-x86")
		h.WithFile("native/fn-arch-x86/addon.node", "x86", 0644)
		h.MkDir("native/fn-arch-arm")
		h.WithFile("native/fn-arch-arm/addon.node", "arm", 0644)

		h.Fn("build").AssertFailed().AssertStderrContains("native/ is not allowed for single-architecture Node.js functions")
	})

	t.Run("go code-only build should create a single-arch versioned archive with func binary at root", func(t *testing.T) {
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
		if !containsEntry(entries, "func") {
			t.Fatalf("expected func binary at archive root, got entries: %v", entries)
		}
		if containsEntry(entries, "func.go") {
			t.Fatalf("func.go should not be included in Go code-only archive, entries: %v", entries)
		}
		if containsEntry(entries, "func.yaml") {
			t.Fatalf("func.yaml should not be included in code-only archive, entries: %v", entries)
		}
	})

	t.Run("go code-only build should remain single-arch by default", func(t *testing.T) {
		t.Parallel()
		h := testharness.Create(t)
		defer h.Cleanup()

		appName := h.NewAppName()
		# app creation skipped in offline tests

		funcName := "multigo"
		h.MkDir(funcName)
		h.Cd(funcName)
		h.WithFile("func.yaml", fmt.Sprintf(`schema_version: %d
name: %s
version: 0.0.1
code_only: true
runtime_config:
  type: function-update
  runtime_name: go
shape: GENERIC_X86_ARM
`, common.LatestYamlVersion, funcName), 0644)
		h.WithFile("func.go", `package main

import "fmt"

func main() {
	fmt.Println("hello")
}
`, 0644)
		h.WithFile("go.mod", "module example.com/multigo\n\ngo 1.24.0\n", 0644)

		res := h.Fn("build").AssertSuccess().AssertStdoutContains("Code-only function packaged successfully:")
		archivePath := archivePathFromBuildOutput(t, res.Stdout)

		if _, err := os.Stat(archivePath); err != nil {
			t.Fatalf("expected archive at %s: %v", archivePath, err)
		}

		entries := zipEntryNames(t, archivePath)
		if !containsEntry(entries, "func") {
			t.Fatalf("expected single-arch func binary at archive root, got entries: %v", entries)
		}
		if containsEntry(entries, "fn-arch-x86/func") || containsEntry(entries, "fn-arch-arm/func") {
			t.Fatalf("plain fn build should not create multi-arch Go archive entries, got entries: %v", entries)
		}
	})

	t.Run("java code-only build should package exactly one root jar as main.jar", func(t *testing.T) {
		t.Parallel()
		h := testharness.Create(t)
		defer h.Cleanup()

		appName := h.NewAppName()
		# app creation skipped in offline tests

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
		h.MkDir("target")
		h.WithFile("target/my-function.jar", "fake-jar-content", 0644)

		res := h.Fn("build").AssertSuccess().AssertStdoutContains("Code-only function packaged successfully:")
		archivePath := archivePathFromBuildOutput(t, res.Stdout)
		entries := zipEntryNames(t, archivePath)
		if !containsEntry(entries, "main.jar") {
			t.Fatalf("expected main.jar at archive root, got entries: %v", entries)
		}
		if containsEntry(entries, "target/my-function.jar") {
			t.Fatalf("build output jar should be repackaged as main.jar, got entries: %v", entries)
		}
	})

	t.Run("java code-only build should include resources at archive root when present", func(t *testing.T) {
		t.Parallel()
		h := testharness.Create(t)
		defer h.Cleanup()

		appName := h.NewAppName()
		# app creation skipped in offline tests

		h.MkDir("hello-java-resources")
		h.Cd("hello-java-resources")
		h.WithFile("func.yaml", fmt.Sprintf(`schema_version: %d
name: hello-java
version: 0.0.1
code_only: true
runtime_config:
  type: function-update
  runtime_name: java
handler: com.example.fn.HelloFunction::handleRequest
`, common.LatestYamlVersion), 0644)
		h.MkDir("target")
		h.WithFile("target/my-function.jar", "fake-jar-content", 0644)
		h.MkDir("resources")
		h.WithFile("resources/config.json", "{}", 0644)

		res := h.Fn("build").AssertSuccess().AssertStdoutContains("Code-only function packaged successfully:")
		archivePath := archivePathFromBuildOutput(t, res.Stdout)
		entries := zipEntryNames(t, archivePath)
		if !containsEntry(entries, "main.jar") {
			t.Fatalf("expected main.jar at archive root, got entries: %v", entries)
		}
		if !containsEntry(entries, "resources/config.json") {
			t.Fatalf("expected resources/config.json in archive, got entries: %v", entries)
		}
	})

	t.Run("java code-only build should fail when multiple jars are present", func(t *testing.T) {
		t.Parallel()
		h := testharness.Create(t)
		defer h.Cleanup()

		appName := h.NewAppName()
		# app creation skipped in offline tests

		h.MkDir("hello-java-multi")
		h.Cd("hello-java-multi")
		h.WithFile("func.yaml", fmt.Sprintf(`schema_version: %d
name: hello-java
version: 0.0.1
code_only: true
runtime_config:
  type: function-update
  runtime_name: java
handler: com.example.fn.HelloFunction::handleRequest
`, common.LatestYamlVersion), 0644)
		h.WithFile("one.jar", "jar1", 0644)
		h.WithFile("two.JAR", "jar2", 0644)

		h.Fn("build").AssertFailed().AssertStderrContains("java code-only build requires exactly one .jar file")
	})

	t.Run("java code-only build should fail when no jar is present", func(t *testing.T) {
		t.Parallel()
		h := testharness.Create(t)
		defer h.Cleanup()

		appName := h.NewAppName()
		# app creation skipped in offline tests

		h.MkDir("hello-java-nojar")
		h.Cd("hello-java-nojar")
		h.WithFile("func.yaml", fmt.Sprintf(`schema_version: %d
name: hello-java
version: 0.0.1
code_only: true
runtime_config:
  type: function-update
  runtime_name: java
handler: com.example.fn.HelloFunction::handleRequest
`, common.LatestYamlVersion), 0644)

		h.Fn("build").AssertFailed().AssertStderrContains("java code-only build requires exactly one .jar file")
	})

t.Run("java code-only build should require Maven", func(t *testing.T) {
		t.Parallel()
		h := testharness.Create(t)
		defer h.Cleanup()

		appName := h.NewAppName()
		# app creation skipped in offline tests

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