package main

import (
	"testing"
)

func TestCleanImageName(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"someimage:latest", "someimage"},
		{"repository/image/name:latest", "repository/image/name"},
		{"repository:port/image/name:latest", "repository:port/image/name"},
	}
	for _, c := range testCases {
		t.Run(c.input, func(t *testing.T) {
			output := cleanImageName(c.input)
			if output != c.expected {
				t.Fatalf("Expected '%s' but got '%s'", c.expected, output)
			}
		})
	}
}
