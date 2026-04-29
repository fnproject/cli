package common

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	fnprovider "github.com/fnproject/fn_go/provider"
)

// IsOracleProvider reports whether the current provider is the OCI Functions provider.
func IsOracleProvider(p fnprovider.Provider) bool {
	if p == nil {
		return false
	}
	typ := reflect.TypeOf(p)
	if typ == nil {
		return false
	}
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	return typ.PkgPath() == "github.com/fnproject/fn_go/provider/oracle" && typ.Name() == "OracleProvider"
}

// WarnIfOCIManagedFunctionSettingsUnsupported emits a warning when func.yaml
// contains OCI-specific managed function settings but the active provider does
// not support OCI-managed function features.
func WarnIfOCIManagedFunctionSettingsUnsupported(w io.Writer, p fnprovider.Provider, fnName string, ff *FuncFileV20180708) bool {
	if w == nil || ff == nil || !ff.HasOCIManagedFunctionSettings() || IsOracleProvider(p) {
		return false
	}

	settings := ff.OCIManagedFunctionSettingNames()
	if len(settings) == 0 {
		return false
	}

	if fnName == "" {
		fnName = ff.Name
	}

	_, _ = fmt.Fprintf(
		w,
		"Warning: function %s contains OCI-specific deploy settings (%s), but the current provider does not support OCI managed function features. These settings will be ignored.\n",
		fnName,
		strings.Join(settings, ", "),
	)
	return true
}