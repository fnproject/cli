package common

import (
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
			if err := ValidateImageName(c.name); err != nil {
				errString = err.Error()
			}
			if c.expectedErr != errString {
				t.Fatalf("expected %s but got %s", c.expectedErr, errString)
			}
		})
	}
}
