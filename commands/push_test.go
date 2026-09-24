package commands

import (
	"net/url"
	"testing"
)

func TestObjectStorageHostFromFnAPIURL(t *testing.T) {
	tests := []struct {
		name      string
		functions string
		want      string
	}{
		{
			name:      "commercial region",
			functions: "https://functions.ca-toronto-1.oraclecloud.com",
			want:      "https://objectstorage.ca-toronto-1.oraclecloud.com",
		},
		{
			name:      "development region",
			functions: "https://functions-dev.ca-toronto-1.oci.oraclecloud.com",
			want:      "https://objectstorage.ca-toronto-1.oci.oraclecloud.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			functionsURL, err := url.Parse(tt.functions)
			if err != nil {
				t.Fatalf("parse Functions API URL: %v", err)
			}

			got, err := objectStorageHostFromFnAPIURL(functionsURL)
			if err != nil {
				t.Fatalf("derive Object Storage host: %v", err)
			}
			if got != tt.want {
				t.Fatalf("derived Object Storage host = %q, want %q", got, tt.want)
			}
		})
	}
}
