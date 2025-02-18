// Package config parse /etc/os-release
// More about the os-release: https://www.linux.org/docs/man5/os-release.html
// We parse /etc/os-release, this is done specifically for a use case on Oracle Cloud Cloud Shell where DockerEngine is not available and Docker cli point to podman cli
package config

import (
	"fmt"
	"os"
	"strings"
)

type OSRelease struct {
	Name       string
	PlatformID string
}

// getLines read the OSReleasePath and return it line by line.
// Empty lines and comments (beginning with a "#") are ignored.
func getLines(path string) ([]string, error) {

	output, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %s", path, err)
	}

	lines := make([]string, 0)

	for _, line := range strings.Split(string(output), "\n") {

		switch true {
		case line == "":
			continue
		case []byte(line)[0] == '#':
			continue
		}

		lines = append(lines, line)
	}

	return lines, nil
}

// parseLine parse a single line.
// Return key, value, error (if any)
func parseLine(line string) (string, string, error) {

	subs := strings.SplitN(line, "=", 2)

	if len(subs) != 2 {
		return "", "", fmt.Errorf("invalid length of the substrings: %d", len(subs))
	}

	return subs[0], strings.Trim(subs[1], "\"'"), nil
}

// Parse parses the os-release file pointing to by path.
// The fields are saved into the Release global variable.
func Parse(path string) (*OSRelease, error) {
	Release := &OSRelease{}
	lines, err := getLines(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get lines of %s: %s", path, err)
	}

	for i := range lines {

		key, value, err := parseLine(lines[i])
		if err != nil {
			return nil, fmt.Errorf("failed to parse line '%s': %s", lines[i], err)
		}

		switch key {
		case "NAME":
			Release.Name = value
		case "PLATFORM_ID":
			Release.PlatformID = value
		default:
		}
	}

	return Release, nil
}
