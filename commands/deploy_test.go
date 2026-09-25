package commands

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/fnproject/cli/config"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

func TestClarifyInvalidCodeOnlyRuntimeName(t *testing.T) {
	t.Run("replaces the service runtime validation message", func(t *testing.T) {
		err := clarifyInvalidCodeOnlyRuntimeName(errors.New("Invalid java"), "java")
		want := "Invalid runtime name 'java'. To check supported runtimes, run `fn list runtimes`."
		if err == nil || err.Error() != want {
			t.Fatalf("error = %v, want %q", err, want)
		}
	})

	t.Run("preserves other service errors", func(t *testing.T) {
		original := errors.New("Invalid handler")
		if got := clarifyInvalidCodeOnlyRuntimeName(original, "java"); got != original {
			t.Fatalf("error = %v, want original error", got)
		}
	})
}

func TestNormalizeRuntimeConfigTypeForDeploy(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "function-update hyphen", input: "function-update", want: "FUNCTION_UPDATE"},
		{name: "function_update underscore", input: "function_update", want: "FUNCTION_UPDATE"},
		{name: "manual", input: "manual", want: "MANUAL"},
		{name: "already upper", input: "FUNCTION_UPDATE", want: "FUNCTION_UPDATE"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeRuntimeConfigTypeForDeploy(tc.input)
			if got != tc.want {
				t.Fatalf("normalizeRuntimeConfigTypeForDeploy(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestResolveCodeOnlyDeployTargetFromContext(t *testing.T) {
	t.Run("configured context should return bucket namespace and configured true", func(t *testing.T) {
		oldHome := os.Getenv("HOME")
		defer func() {
			_ = os.Setenv("HOME", oldHome)
			homedir.Reset()
		}()

		tmpHome := t.TempDir()
		if err := os.Setenv("HOME", tmpHome); err != nil {
			t.Fatalf("failed to set HOME: %v", err)
		}
		homedir.Reset()

		contextsDir := filepath.Join(tmpHome, ".fn", "contexts")
		if err := os.MkdirAll(contextsDir, 0755); err != nil {
			t.Fatalf("failed to create contexts dir: %v", err)
		}

		contextName := "testctx"
		contextPath := filepath.Join(contextsDir, contextName+".yaml")
		content := []byte("provider: oracle\nobject_storage_bucket_name: code-only-test-files\nnamespace: oraclefunctionsdevelopm\n")
		if err := os.WriteFile(contextPath, content, 0644); err != nil {
			t.Fatalf("failed to write context file: %v", err)
		}

		oldContext := viper.GetString(config.CurrentContext)
		defer viper.Set(config.CurrentContext, oldContext)
		viper.Set(config.CurrentContext, contextName)

		bucket, namespace, configured, err := resolveCodeOnlyDeployTargetFromContext()
		if err != nil {
			t.Fatalf("resolveCodeOnlyDeployTargetFromContext returned error: %v", err)
		}
		if bucket != "code-only-test-files" {
			t.Fatalf("bucket = %q, want %q", bucket, "code-only-test-files")
		}
		if namespace != "oraclefunctionsdevelopm" {
			t.Fatalf("namespace = %q, want %q", namespace, "oraclefunctionsdevelopm")
		}
		if !configured {
			t.Fatal("configured = false, want true")
		}
	})

	t.Run("missing bucket or namespace should return configured false", func(t *testing.T) {
		oldHome := os.Getenv("HOME")
		defer func() {
			_ = os.Setenv("HOME", oldHome)
			homedir.Reset()
		}()

		tmpHome := t.TempDir()
		if err := os.Setenv("HOME", tmpHome); err != nil {
			t.Fatalf("failed to set HOME: %v", err)
		}
		homedir.Reset()

		contextsDir := filepath.Join(tmpHome, ".fn", "contexts")
		if err := os.MkdirAll(contextsDir, 0755); err != nil {
			t.Fatalf("failed to create contexts dir: %v", err)
		}

		contextName := "emptyctx"
		contextPath := filepath.Join(contextsDir, contextName+".yaml")
		content := []byte("provider: oracle\n")
		if err := os.WriteFile(contextPath, content, 0644); err != nil {
			t.Fatalf("failed to write context file: %v", err)
		}

		oldContext := viper.GetString(config.CurrentContext)
		defer viper.Set(config.CurrentContext, oldContext)
		viper.Set(config.CurrentContext, contextName)

		bucket, namespace, configured, err := resolveCodeOnlyDeployTargetFromContext()
		if err != nil {
			t.Fatalf("resolveCodeOnlyDeployTargetFromContext returned error: %v", err)
		}
		if bucket != "" || namespace != "" {
			t.Fatalf("expected empty bucket/namespace, got %q / %q", bucket, namespace)
		}
		if configured {
			t.Fatal("configured = true, want false")
		}
	})
}
