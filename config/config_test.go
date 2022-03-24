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

package config

import (
	"os"
	"reflect"
	"testing"

	"github.com/fnproject/fn_go"
	"github.com/fnproject/fn_go/provider"
)

func TestDefaultContextConfigContents(t *testing.T) {
	tests := []struct {
		name           string
		OciCliAuth     string
		wantContextMap *ContextMap
	}{
		{
			name:       "unspecified",
			OciCliAuth: "",
			wantContextMap: &ContextMap{
				ContextProvider:      fn_go.DefaultProvider,
				provider.CfgFnAPIURL: defaultLocalAPIURL,
				EnvFnRegistry:        "",
			},
		},
		{
			name:       "api_key",
			OciCliAuth: "api_key",
			wantContextMap: &ContextMap{
				ContextProvider:      fn_go.DefaultProvider,
				provider.CfgFnAPIURL: defaultLocalAPIURL,
				EnvFnRegistry:        "",
			},
		},
		{
			name:       "instance_obo_user",
			OciCliAuth: "instance_obo_user",
			wantContextMap: &ContextMap{
				ContextProvider: fn_go.OracleCSProvider,
				EnvFnRegistry:   "",
			},
		},
		{
			name:       "instance_principal",
			OciCliAuth: "instance_principal",
			wantContextMap: &ContextMap{
				ContextProvider: fn_go.OracleIPProvider,
				EnvFnRegistry:   "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(OCI_CLI_AUTH_ENV_VAR, tt.OciCliAuth)
			if gotContextMap := DefaultContextConfigContents(); !reflect.DeepEqual(gotContextMap, tt.wantContextMap) {
				t.Errorf("DefaultContextConfigContents() = %v, want %v", gotContextMap, tt.wantContextMap)
			}
		})
	}
}

func TestValidateContainerEngineType(t *testing.T) {
	testCases := []struct {
		actual      string
		expectedErr string
	}{
		{actual: "docker", expectedErr: ""},
		{actual: "podman", expectedErr: ""},
		{actual: "default", expectedErr: "Invalid Container Engine"},
	}
	for _, c := range testCases {
		t.Run(c.actual, func(t *testing.T) {
			errString := ""
			if err := ValidateContainerEngineType(c.actual); err != nil {
				errString = err.Error()
			}
			if c.expectedErr != errString {
				t.Fatalf("expected %s but got %s", c.expectedErr, errString)
			}
		})
	}
}
