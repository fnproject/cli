package langs

import "testing"

func TestDefaultRuntimeVersions(t *testing.T) {
	tests := []struct {
		runtime string
		want    string
	}{
		{runtime: "node", want: "node24"},
		{runtime: "java", want: "java21"},
	}

	for _, tt := range tests {
		t.Run(tt.runtime, func(t *testing.T) {
			helper := GetLangHelper(tt.runtime)
			if helper == nil {
				t.Fatalf("expected helper for runtime %q", tt.runtime)
			}

			got := helper.LangStrings()[1]
			if got != tt.want {
				t.Fatalf("expected default runtime %q to resolve to %q, got %q", tt.runtime, tt.want, got)
			}
		})
	}
}

func TestFallbackRuntimeVersions(t *testing.T) {
	tests := []struct {
		runtime string
		want    string
	}{
		{runtime: "node", want: "node22"},
		{runtime: "java", want: "java17"},
	}

	for _, tt := range tests {
		t.Run(tt.runtime, func(t *testing.T) {
			helper := GetFallbackLangHelper(tt.runtime)
			if helper == nil {
				t.Fatalf("expected fallback helper for runtime %q", tt.runtime)
			}

			got := helper.LangStrings()[1]
			if got != tt.want {
				t.Fatalf("expected fallback runtime %q to resolve to %q, got %q", tt.runtime, tt.want, got)
			}
		})
	}
}

func TestNode24Images(t *testing.T) {
	helper := &NodeLangHelper{Version: "24"}

	buildImage, err := helper.BuildFromImage()
	if err != nil {
		t.Fatalf("BuildFromImage() returned error: %v", err)
	}
	runImage, err := helper.RunFromImage()
	if err != nil {
		t.Fatalf("RunFromImage() returned error: %v", err)
	}

	if buildImage != "fnproject/node:24-dev" {
		t.Fatalf("expected node24 build image %q, got %q", "fnproject/node:24-dev", buildImage)
	}
	if runImage != "fnproject/node:24" {
		t.Fatalf("expected node24 run image %q, got %q", "fnproject/node:24", runImage)
	}
}

func TestJava21Images(t *testing.T) {
	helper := &JavaLangHelper{Version: "21", latestFdkVersion: "1.2.3"}

	buildImage, err := helper.BuildFromImage()
	if err != nil {
		t.Fatalf("BuildFromImage() returned error: %v", err)
	}
	runImage, err := helper.RunFromImage()
	if err != nil {
		t.Fatalf("RunFromImage() returned error: %v", err)
	}

	if buildImage != "fnproject/fn-java-fdk-build:jdk21-1.2.3" {
		t.Fatalf("expected java21 build image %q, got %q", "fnproject/fn-java-fdk-build:jdk21-1.2.3", buildImage)
	}
	if runImage != "fnproject/fn-java-fdk:jre21-1.2.3" {
		t.Fatalf("expected java21 run image %q, got %q", "fnproject/fn-java-fdk:jre21-1.2.3", runImage)
	}
}