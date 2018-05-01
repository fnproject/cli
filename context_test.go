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
		{name: "local-context", expectedErr: ""},
		{name: "local_context", expectedErr: ""},
		{name: "local1", expectedErr: ""},
		{name: "local-context-1", expectedErr: ""},
		{name: "../local", expectedErr: "please enter a context name with only Alphanumeric, _, or -"},
		{name: "local?context", expectedErr: "please enter a context name with only Alphanumeric, _, or -"},
		{name: "context>?", expectedErr: "please enter a context name with only Alphanumeric, _, or -"},
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

func TestValidateAPIURL(t *testing.T) {
	var testCases = []struct {
		apiURL      string
		expectedErr string
	}{
		{apiURL: "http://test.com", expectedErr: ""},
		{apiURL: "http://test", expectedErr: ""},
		{apiURL: "test.com", expectedErr: "invalid url does not contain ://"},
		{apiURL: "http:/test.com", expectedErr: "invalid url does not contain ://"},
		{apiURL: "://testcom", expectedErr: "invalid Fn API URL: parse ://testcom: missing protocol scheme"},
	}
	for _, tc := range testCases {
		t.Run(tc.apiURL, func(t *testing.T) {
			errString := ""
			if err := ValidateAPIURL(tc.apiURL); err != nil {
				errString = err.Error()
			}
			if tc.expectedErr != errString {
				t.Fatalf("expected %s but got %s", tc.expectedErr, errString)
			}
		})
	}
}
