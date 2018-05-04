package testharness

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func assertFileExists(t *testing.T, destPath string) {
	stat, err := os.Stat(destPath)
	if err != nil {
		t.Fatalf("expecting path %s to exist", destPath)
	}

	if stat.Size() == 0 {
		t.Fatalf("expecting path %s to contain data ", destPath)
	}

}

func assertContents(t *testing.T, destPath, content string) {
	data, err := ioutil.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Error reading file %s : %s", destPath, err)
	}

	if string(data) != content {
		t.Fatalf("expecting %s to contain `%s` but was `%s`", destPath, content, string(data))
	}

}

func TestCopyContext(t *testing.T) {

	ctx := Create(t)
	defer ctx.Cleanup()
	ctx.CopyFiles(map[string]string{
		"testdir/testfiles": "tf",
		"harness.go":        "harness.go",
		"harness_test.go":   "foo/cli_test.go",
	})

	assertContents(t, path.Join(ctx.testDir, "tf/test.txt"), "hello world")
	assertFileExists(t, path.Join(ctx.testDir, "harness.go"))
	assertFileExists(t, path.Join(ctx.testDir, "foo/cli_test.go"))
}

func TestFileManipulation(t *testing.T) {

	ctx := Create(t)
	defer ctx.Cleanup()

	ctx.WithFile("fileA.txt", "Foo", 0644)
	assertContents(t, path.Join(ctx.testDir, "fileA.txt"), "Foo")

	ctx.MkDir("td")
	ctx.WithFile("td/file.txt", "Foo", 0644)
	assertContents(t, path.Join(ctx.testDir, "td/file.txt"), "Foo")

	contents := ctx.GetFile("td/file.txt")
	if contents != "Foo" {
		t.Errorf("Failed to get file contents , expected Foo, got %s", contents)
	}

	ctx.MkDir("testDir")
	ctx.Cd("testDir")
	ctx.WithFile("fileB.txt", "Bar", 0644)

	assertContents(t, path.Join(ctx.testDir, "testDir/fileB.txt"), "Bar")

	contents = ctx.GetFile("fileB.txt")
	if contents != "Bar" {
		t.Errorf("Failed to get file contents , expected Bar, got %s", contents)
	}

	ctx.Cd("")
	ctx.WithFile("baseFile", "value1", 0644)

	ctx.FileAppend("baseFile", "value2")
	assertContents(t, path.Join(ctx.testDir, "baseFile"), "value1value2")

}

func TestDirOps(t *testing.T) {

	ctx := Create(t)
	defer ctx.Cleanup()
	ctx.MkDir("foo")
	ctx.Cd("foo")
	ctx.Cd("../")
	ctx.MkDir("bar")
	ctx.Cd("bar")
	ctx.MkDir("baz")
	ctx.Cd("baz")
	ctx.WithFile("bob.txt", "some text", 0644)
	ctx.Cd("")

	ctx.WithFile("root.txt", "some text", 0644)
	assertFileExists(t, path.Join(ctx.testDir, "foo"))
	assertFileExists(t, path.Join(ctx.testDir, "bar"))
	assertFileExists(t, path.Join(ctx.testDir, "bar/baz"))
	assertFileExists(t, path.Join(ctx.testDir, "bar/baz/bob.txt"))
	assertFileExists(t, path.Join(ctx.testDir, "root.txt"))

}
