package main

import (
	"testing"
)

func TestValidateContextName(t *testing.T) {
	var testsCases = []struct {
		name        string
		expectedErr string
	}{
		{name: "local", expectedErr: ""},
		{name: "Local", expectedErr: ""},
		{name: "../local", expectedErr: "please enter a context name with ASCII characters only"},
		{name: "local-context", expectedErr: ""},
		{name: "local_context", expectedErr: ""},
		{name: "local1", expectedErr: ""},
		{name: "local-context-1", expectedErr: ""},
		{name: "local?context", expectedErr: "please enter a context name with ASCII characters only"},
		{name: "context>?", expectedErr: "please enter a context name with ASCII characters only"},
	}
	for _, tc := range testsCases {
		t.Run(tc.name, func(t *testing.T) {
			errString := ""
			if err := ValidateContextName(tc.name); err != nil {
				errString = err.Error()
			}
			if tc.expectedErr != errString {
				t.Fatalf("expected %s but got %s", tc.expectedErr, errString)
			}
		})
	}
}
