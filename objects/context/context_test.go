/*
 * Copyright (c) 2019, 2020 Oracle and/or its affiliates. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package context

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
		{name: "../local", expectedErr: "Please enter a context name with only Alphanumeric, _, or -"},
		{name: "local?context", expectedErr: "Please enter a context name with only Alphanumeric, _, or -"},
		{name: "context>?", expectedErr: "Please enter a context name with only Alphanumeric, _, or -"},
	}
	for _, tc := range testsCases {
		t.Run(tc.name, func(t *testing.T) {
			errString := ""
			if err := ValidateContextName(tc.name); err != nil {
				errString = err.Error()
			}
			if tc.expectedErr != errString {
				t.Fatalf("Expected %s but got %s", tc.expectedErr, errString)
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
		{apiURL: "test.com", expectedErr: "Invalid Fn API URL: does not contain ://"},
		{apiURL: "http:/test.com", expectedErr: "Invalid Fn API URL: does not contain ://"},
		{apiURL: "://testcom", expectedErr: "Invalid Fn API URL: parse \"://testcom\": missing protocol scheme"},
	}
	for _, tc := range testCases {
		t.Run(tc.apiURL, func(t *testing.T) {
			errString := ""
			if err := ValidateAPIURL(tc.apiURL); err != nil {
				errString = err.Error()
			}
			if tc.expectedErr != errString {
				t.Fatalf("Expected %s but got %s", tc.expectedErr, errString)
			}
		})
	}
}
