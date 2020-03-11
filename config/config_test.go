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
