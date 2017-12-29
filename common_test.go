package main

import (
	"strings"
	"testing"
)

func TestValidateImageName(t *testing.T) {
	longTagName := strings.Repeat("a", 128)
	tooLongTagName := strings.Repeat("a", 129)
	cases := []struct {
		name        string
		willBeError bool
	}{
		{"dockeruser/name:0.0.1", false},
		{"dockeruser/n:0.0.1", false},
		{"dockeruser/na.me:0.0.1", false},
		{"dockeruser/na__me:0.0.1", false},
		{"127.0.0.1/dockeruser/name:0.0.1", false},
		{"localhost/dockeruser/name:0.0.1", false},
		{"127.0.0.1:5000/dockeruser/name:0.0.1", false},
		{"localhost:5000/dockeruser/name:0.0.1", false},
		{"dockeruser/name:" + longTagName, false},
		{"dockeruser/name:" + tooLongTagName, true},
		{"dockeruser/name:", true},
		{"dockeruser/name:.0.0.1", true},
		{"dockeruser/name:-0.0.1", true},
		{"local_host:5000/dockeruser/name:0.0.1", true},
		{"dockeruser/Name:0.0.1", true},
		{"dockeruser/_name:0.0.1", true},
		{"dockeruser/name_:0.0.1", true},
		{"dockeruser/na___me:0.0.1", true},
		{"dockeruser/na..me:0.0.1", true},
	}
	for _, c := range cases {
		if err := validateImageName(c.name); (err != nil) != c.willBeError {
			if err != nil {
				t.Errorf("%v is valid name but validateImageName return error: %v", c.name, err)
			} else {
				t.Errorf("%v is invalid name but validateImageName return nil", c.name)
			}
		}
	}
}
