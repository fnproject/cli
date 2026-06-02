package common

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadCLIJSONInput(t *testing.T) {
	var parsed struct {
		Name string `json:"name"`
	}
	if err := LoadCLIJSONInput(`{"name":"hello"}`, &parsed); err != nil {
		t.Fatalf("LoadCLIJSONInput(raw) error = %v", err)
	}
	if parsed.Name != "hello" {
		t.Fatalf("expected parsed name hello, got %q", parsed.Name)
	}
	tmp := t.TempDir()
	path := filepath.Join(tmp, "input.json")
	if err := os.WriteFile(path, []byte(`{"name":"world"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	parsed = struct {
		Name string `json:"name"`
	}{}
	if err := LoadCLIJSONInput("file://"+path, &parsed); err != nil {
		t.Fatalf("LoadCLIJSONInput(file) error = %v", err)
	}
	if parsed.Name != "world" {
		t.Fatalf("expected parsed name world, got %q", parsed.Name)
	}
}
