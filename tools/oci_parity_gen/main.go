package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fnproject/cli/internal/ociparity"
)

func main() {
	var specPath string
	flag.StringVar(&specPath, "spec", os.Getenv("SPEC"), "Path to OCI Functions API spec")
	flag.Parse()
	if specPath == "" {
		fmt.Fprintln(os.Stderr, "missing spec path; use SPEC=/absolute/path/to/functions-api-spec.yaml make generate-oci-parity")
		os.Exit(1)
	}
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	root := filepath.Clean(filepath.Join(wd))
	if err := ociparity.WriteGeneratedFiles(root, specPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
