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

package common

import (
	"os"
	"reflect"
	"testing"
)

func TestValidateImageName(t *testing.T) {
	testCases := []struct {
		name        string
		expectedErr string
	}{
		{name: "docker.io/sally/img:0.0.1", expectedErr: ""},
		{name: "sally/img:0.0.1", expectedErr: ""},
		{name: "img:0.0.1", expectedErr: "image name must have a dockerhub owner or private registry. Be sure to set FN_REGISTRY env var, pass in --registry or configure your context file"},
		{name: "owner/img", expectedErr: "image name must have a tag"},
	}
	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			errString := ""
			if err := ValidateFullImageName(c.name); err != nil {
				errString = err.Error()
			}
			if c.expectedErr != errString {
				t.Fatalf("expected %s but got %s", c.expectedErr, errString)
			}
		})
	}
}

func Test_proxyArgs(t *testing.T) {
	tests := []struct {
		name string
		set  []string
		want []string
	}{
		{"empty", []string{}, []string{}},
		{"populated", []string{"http_proxy", "https_proxy", "no_proxy", "foo"}, []string{
			"-e", "http_proxy=value_of_http_proxy",
			"-e", "https_proxy=value_of_https_proxy",
			"-e", "no_proxy=value_of_no_proxy"}},
		{"partial", []string{"http_proxy", "no_proxy", "foo"}, []string{
			"-e", "http_proxy=value_of_http_proxy",
			"-e", "no_proxy=value_of_no_proxy"}},
	}
	for _, tt := range tests {
		old := map[string]string{
			"http_proxy":  "",
			"https_proxy": "",
			"no_proxy":    "",
			"foo":         "",
		}
		for k, _ := range old {
			old[k] = os.Getenv(k)
			_ = os.Unsetenv(k)
		}
		t.Run(tt.name, func(t *testing.T) {
			for _, k := range tt.set {
				_ = os.Setenv(k, "value_of_"+k)
			}
			if got := proxyArgs(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("proxyArgs() = %v, want %v", got, tt.want)
			}
		})
		for k, v := range old {
			_ = os.Setenv(k, v)
		}
	}
}

func Test_addNamespace(t *testing.T) {
	testCases := []struct {
		image         string
		fullImageName string
	}{
		{image: "fnproject/go", fullImageName: ContainerRegistryNamespace + "fnproject/go"},
		{image: "fnproject/dotnet", fullImageName: ContainerRegistryNamespace + "fnproject/dotnet"},
	}

	for _, c := range testCases {
		t.Run(c.image, func(t *testing.T) {
			imageName := AddContainerNamespace(c.image)
			if imageName != c.fullImageName {
				t.Fatalf("expected %s but got %s", c.fullImageName, imageName)
			}
			t.Logf("Output %s", imageName)
		})
	}
}
