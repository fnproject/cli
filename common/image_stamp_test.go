package common

import (
	"path/filepath"
	"testing"
)

func TestImageStampFuncFileV20180708UsesExpectedRuntimeVersion(t *testing.T) {
	t.Setenv("FN_JAVA_FDK_VERSION", "1.2.3")

	tests := []struct {
		name      string
		runtime   string
		wantBuild string
		wantRun   string
	}{
		{
			name:      "legacy node runtime uses fallback version",
			runtime:   "node",
			wantBuild: "fnproject/node:22-dev",
			wantRun:   "fnproject/node:22",
		},
		{
			name:      "explicit node24 runtime keeps requested version",
			runtime:   "node24",
			wantBuild: "fnproject/node:24-dev",
			wantRun:   "fnproject/node:24",
		},
		{
			name:      "legacy java runtime uses fallback version",
			runtime:   "java",
			wantBuild: "fnproject/fn-java-fdk-build:jdk17-1.2.3",
			wantRun:   "fnproject/fn-java-fdk:jre17-1.2.3",
		},
		{
			name:      "explicit java21 runtime keeps requested version",
			runtime:   "java21",
			wantBuild: "fnproject/fn-java-fdk-build:jdk21-1.2.3",
			wantRun:   "fnproject/fn-java-fdk:jre21-1.2.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "func.yaml")
			ff := &FuncFileV20180708{Runtime: tt.runtime}

			stamped, err := imageStampFuncFileV20180708(path, ff)
			if err != nil {
				t.Fatalf("imageStampFuncFileV20180708() returned error: %v", err)
			}

			if stamped.Build_image != tt.wantBuild {
				t.Fatalf("expected build image %q, got %q", tt.wantBuild, stamped.Build_image)
			}
			if stamped.Run_image != tt.wantRun {
				t.Fatalf("expected run image %q, got %q", tt.wantRun, stamped.Run_image)
			}
		})
	}
}