package common

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	ProvisionedConcurrencyStrategyNone     = "NONE"
	ProvisionedConcurrencyStrategyConstant = "CONSTANT"
	ProvisionedConcurrencyCountStep        = 40
)

// ValidateProvisionedConcurrencyConfig validates OCI provisioned concurrency values
// before they are sent to OCI.
func ValidateProvisionedConcurrencyConfig(cfg *OCIProvisionedConcurrencyConfig) error {
	if cfg == nil {
		return nil
	}

	switch strings.ToUpper(strings.TrimSpace(cfg.Strategy)) {
	case ProvisionedConcurrencyStrategyNone:
		return nil
	case ProvisionedConcurrencyStrategyConstant:
		if cfg.Count == nil || *cfg.Count <= 0 {
			return fmt.Errorf("provisioned concurrency count must be a positive integer")
		}
		if *cfg.Count%ProvisionedConcurrencyCountStep != 0 {
			return fmt.Errorf("provisioned concurrency count must be a multiple of %d", ProvisionedConcurrencyCountStep)
		}
		return nil
	default:
		return fmt.Errorf("unsupported provisioned concurrency strategy %q", cfg.Strategy)
	}
}

// ParseProvisionedConcurrencySpec parses a curated provisioned concurrency CLI value.
// Supported values are:
//   - none
//   - constant:<count>
func ParseProvisionedConcurrencySpec(spec string) (*OCIProvisionedConcurrencyConfig, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return nil, nil
	}

	if strings.EqualFold(spec, "none") {
		return &OCIProvisionedConcurrencyConfig{Strategy: ProvisionedConcurrencyStrategyNone}, nil
	}

	parts := strings.SplitN(spec, ":", 2)
	if len(parts) != 2 || !strings.EqualFold(strings.TrimSpace(parts[0]), "constant") {
		return nil, fmt.Errorf("invalid value for --provisioned-concurrency: %q (expected 'none' or 'constant:<count>')", spec)
	}

	count, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil || count <= 0 {
		return nil, fmt.Errorf("invalid value for --provisioned-concurrency: %q (count must be a positive integer)", spec)
	}

	cfg := &OCIProvisionedConcurrencyConfig{
		Strategy: ProvisionedConcurrencyStrategyConstant,
		Count:    &count,
	}
	if err := ValidateProvisionedConcurrencyConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid value for --provisioned-concurrency: %q (%s)", spec, err)
	}
	return cfg, nil
}

// SetProvisionedConcurrency stores provisioned concurrency config in the function deploy section.
func SetProvisionedConcurrency(ff *FuncFileV20180708, cfg *OCIProvisionedConcurrencyConfig) {
	if ff == nil || cfg == nil {
		return
	}
	if ff.Deploy == nil {
		ff.Deploy = &FuncDeployConfig{}
	}
	if ff.Deploy.OCI == nil {
		ff.Deploy.OCI = &OCIFunctionDeployConfig{}
	}
	ff.Deploy.OCI.ProvisionedConcurrency = cfg
}
