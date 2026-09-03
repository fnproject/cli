package config

import (
	"fmt"
	"os"
	"testing"
)

func TestParse(t *testing.T) {

	var path = "os-release-test"
	osrelease, err := Parse(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse os-relese: %s\n", err)
		os.Exit(1)
	}

	switch true {
	case osrelease.Name != "Oracle Linux Server":
		t.Errorf("Test failed on NAME: want 'Oracle Linux Server', got '%s'\n", osrelease.Name)
	case osrelease.PlatformID != OCI_CLOUDSHELL_OL9_PLATFORM_ID:
		t.Errorf("Test failed on PLATFORM_ID: want '%s', got '%s'\n", OCI_CLOUDSHELL_OL9_PLATFORM_ID, osrelease.PlatformID)
	}
}

func TestIsCloudShellOS(t *testing.T) {
	tests := []struct {
		name      string
		osrelease OSRelease
		want      bool
	}{
		{"OL8 Cloud Shell", OSRelease{Name: OCI_CLOUDSHELL_OS_NAME, PlatformID: OCI_CLOUDSHELL_OL8_PLATFORM_ID}, true},
		{"OL9 Cloud Shell", OSRelease{Name: OCI_CLOUDSHELL_OS_NAME, PlatformID: OCI_CLOUDSHELL_OL9_PLATFORM_ID}, true},
		{"unsupported Oracle Linux version", OSRelease{Name: OCI_CLOUDSHELL_OS_NAME, PlatformID: "platform:el10"}, false},
		{"different operating system", OSRelease{Name: "Other Linux", PlatformID: OCI_CLOUDSHELL_OL9_PLATFORM_ID}, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isCloudShellOS(&test.osrelease); got != test.want {
				t.Errorf("isCloudShellOS() = %t, want %t", got, test.want)
			}
		})
	}
}
