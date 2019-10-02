package test

import (
	"strings"
	"testing"

	"github.com/fnproject/cli/testharness"
)

func TestResourceCopy(t *testing.T) {

	const testFuncSource = `package main
	import (
		"context"
		"io"
		"os" 
		fdk "github.com/fnproject/fdk-go"
	)
	func main() {
		fdk.Handle(fdk.HandlerFunc(myHandler))
	}
	func myHandler(ctx context.Context, in io.Reader, out io.Writer) {
		testFile,err := os.Open("/testfile.txt")
		if err != nil {
			panic(err.Error()) 
		}
		defer testFile.Close()
		io.Copy(out,testFile) 
	}
	`

	const funcYaml = 
`schema_version: 20180708
name: filefn
version: 0.0.1
runtime: go
entrypoint: ./func
copy_resources:
   - from: testfile.txt
     to: /
`

	const gomod = `module func`
	const testFile = `hello world!`

	t.Parallel()
	h := testharness.Create(t)
	defer h.Cleanup()

	appName := h.NewAppName()

	h.WithFile("func.yaml", funcYaml, 0644)
	h.WithFile("func.go", testFuncSource, 0644)
	h.WithFile("testfile.txt", testFile, 0644)
	h.WithFile("go.mod", gomod, 0644)

	h.Fn("deploy", "--app", appName, "--local", "--create-app").AssertSuccess()

	res := h.Fn("invoke", appName, "filefn")

	res.AssertSuccess()
	if strings.TrimSpace(res.Stdout) != testFile {
		t.Fatalf("Expected result to be '%s', got '%s'", testFile, res.Stdout)
	}
}
