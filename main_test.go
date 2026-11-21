package main

import (
	"errors"
	"strings"
	"testing"
)

func TestWorkRequestFailureHint(t *testing.T) {
	workRequestID := "ocid1.functionsworkrequest.oc1.ca-toronto-1.exampleuniqueid"
	hint := workRequestFailureHint(errors.New("work request " + workRequestID + " ended with status FAILED: Function creation failed"))
	for _, want := range []string{
		"fn work-request error " + workRequestID,
		"fn work-request status " + workRequestID,
	} {
		if !strings.Contains(hint, want) {
			t.Fatalf("hint = %q, want %q", hint, want)
		}
	}
}

func TestWorkRequestFailureHintIgnoresOtherErrors(t *testing.T) {
	if hint := workRequestFailureHint(errors.New("invalid function configuration")); hint != "" {
		t.Fatalf("hint = %q, want empty", hint)
	}
}
