package config

import (
	"fmt"
	"os"
	"testing"
)

// This test is written for testing the parsing of file /etc/os-release on OL8 Cloud Shell
// Please modify the test-case when we move to OL8+ on cloudshell
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
	case osrelease.PlatformID != "platform:el8":
		t.Errorf("Test failed on PLATFORM_ID: want 'platform:el8', got '%s'\n", osrelease.PlatformID)
	}
}
