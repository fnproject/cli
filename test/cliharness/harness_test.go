package cliharness

import (
	"testing"
	"os"
	"path"
)

func assertFileExists(t *testing.T, paths ... string) {
	destPath := path.Join(paths...)
	stat, err := os.Stat(destPath)
	if err != nil {
		t.Fatalf("expecting path %s to exist", destPath)
	}

	if stat.Size() == 0 {
		t.Fatalf("expecting path %s to contain data ", destPath)
	}

}

func TestCopyContext(t *testing.T) {

	ctx := Create(t)
	defer ctx.Cleanup()
	ctx.CopyFiles(map[string]string{
		"testdir/testfiles":        "tf",
		"harness.go":      "harness.go",
		"harness_test.go": "foo/cli_test.go",
	})

	assertFileExists(t, ctx.testDir, "tf/test.txt")
	assertFileExists(t, ctx.testDir, "harness.go")
	assertFileExists(t, ctx.testDir, "foo/cli_test.go")
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
	ctx.WithFile("bob.txt","some text")

	assertFileExists(t, ctx.testDir, "foo")
	assertFileExists(t, ctx.testDir, "bar")
	assertFileExists(t, ctx.testDir, "bar/baz")
	assertFileExists(t, ctx.testDir, "bar/baz/bob.txt")


}
