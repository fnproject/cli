package common

import (
	"fmt"
	"strings"
)

const (
	DestinationTypeStream        = "stream"
	DestinationTypeQueue         = "queue"
	DestinationTypeNotifications = "notifications"
)

// ParseOCIDestinationSpec parses a shorthand destination spec like stream:<ocid>.
func ParseOCIDestinationSpec(flagName, spec string) (*OCIDestination, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return nil, nil
	}
	parts := strings.SplitN(spec, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid value for %s: %q (expected <stream|queue|notifications>:<ocid>)", flagName, spec)
	}
	typePart := strings.ToLower(strings.TrimSpace(parts[0]))
	ocid := strings.TrimSpace(parts[1])
	if ocid == "" {
		return nil, fmt.Errorf("invalid value for %s: %q (destination OCID is required)", flagName, spec)
	}
	switch typePart {
	case DestinationTypeStream, DestinationTypeQueue:
		return &OCIDestination{Type: typePart, OCID: ocid}, nil
	case "notification", DestinationTypeNotifications:
		return &OCIDestination{Type: DestinationTypeNotifications, OCID: ocid}, nil
	default:
		return nil, fmt.Errorf("invalid value for %s: %q (unsupported destination type %q)", flagName, spec, typePart)
	}
}

func ensureDetachedModeConfig(ff *FuncFileV20180708) *OCIDetachedModeConfig {
	if ff.Deploy == nil {
		ff.Deploy = &FuncDeployConfig{}
	}
	if ff.Deploy.OCI == nil {
		ff.Deploy.OCI = &OCIFunctionDeployConfig{}
	}
	if ff.Deploy.OCI.DetachedMode == nil {
		ff.Deploy.OCI.DetachedMode = &OCIDetachedModeConfig{}
	}
	return ff.Deploy.OCI.DetachedMode
}

// SetOnSuccessDestination stores an OCI on-success destination in func.yaml config.
func SetOnSuccessDestination(ff *FuncFileV20180708, dest *OCIDestination) {
	if ff == nil || dest == nil {
		return
	}
	ensureDetachedModeConfig(ff).OnSuccess = dest
}

// SetOnFailureDestination stores an OCI on-failure destination in func.yaml config.
func SetOnFailureDestination(ff *FuncFileV20180708, dest *OCIDestination) {
	if ff == nil || dest == nil {
		return
	}
	ensureDetachedModeConfig(ff).OnFailure = dest
}