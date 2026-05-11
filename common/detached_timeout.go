package common

import (
	"fmt"
	"strings"
	"time"
)

// ParseDetachedTimeoutSpec validates a human-friendly detached timeout such as 20m or 1h
// and returns the original trimmed value plus the timeout in seconds.
func ParseDetachedTimeoutSpec(spec string) (string, int, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return "", 0, nil
	}
	d, err := time.ParseDuration(spec)
	if err != nil {
		return "", 0, fmt.Errorf("invalid value for --detached-timeout: %q", spec)
	}
	if d <= 0 {
		return "", 0, fmt.Errorf("invalid value for --detached-timeout: %q", spec)
	}
	if d%time.Second != 0 {
		return "", 0, fmt.Errorf("invalid value for --detached-timeout: %q (must resolve to whole seconds)", spec)
	}
	return spec, int(d / time.Second), nil
}

// SetDetachedTimeout stores detached timeout config in the function deploy section.
func SetDetachedTimeout(ff *FuncFileV20180708, timeout string) {
	if ff == nil || strings.TrimSpace(timeout) == "" {
		return
	}
	if ff.Deploy == nil {
		ff.Deploy = &FuncDeployConfig{}
	}
	if ff.Deploy.OCI == nil {
		ff.Deploy.OCI = &OCIFunctionDeployConfig{}
	}
	if ff.Deploy.OCI.DetachedMode == nil {
		ff.Deploy.OCI.DetachedMode = &OCIDetachedModeConfig{}
	}
	ff.Deploy.OCI.DetachedMode.Timeout = strings.TrimSpace(timeout)
}