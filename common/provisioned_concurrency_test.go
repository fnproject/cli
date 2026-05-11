package common

import "testing"

func TestParseProvisionedConcurrencySpec(t *testing.T) {
	tests := []struct {
		name      string
		spec      string
		wantNil   bool
		wantErr   bool
		strategy  string
		wantCount int
	}{
		{name: "empty", spec: "", wantNil: true},
		{name: "none", spec: "none", strategy: ProvisionedConcurrencyStrategyNone},
		{name: "none case insensitive", spec: "NoNe", strategy: ProvisionedConcurrencyStrategyNone},
		{name: "constant", spec: "constant:40", strategy: ProvisionedConcurrencyStrategyConstant, wantCount: 40},
		{name: "constant with spaces", spec: " constant:80 ", strategy: ProvisionedConcurrencyStrategyConstant, wantCount: 80},
		{name: "invalid missing count", spec: "constant", wantErr: true},
		{name: "invalid non positive", spec: "constant:0", wantErr: true},
		{name: "invalid non numeric", spec: "constant:abc", wantErr: true},
		{name: "invalid non multiple", spec: "constant:5", wantErr: true},
		{name: "invalid strategy", spec: "foo:1", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := ParseProvisionedConcurrencySpec(tt.spec)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseProvisionedConcurrencySpec() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantNil {
				if cfg != nil {
					t.Fatalf("expected nil config, got %#v", cfg)
				}
				return
			}
			if tt.wantErr {
				return
			}
			if cfg == nil {
				t.Fatal("expected non-nil config")
			}
			if cfg.Strategy != tt.strategy {
				t.Fatalf("expected strategy %q, got %q", tt.strategy, cfg.Strategy)
			}
			if tt.strategy == ProvisionedConcurrencyStrategyConstant {
				if cfg.Count == nil || *cfg.Count != tt.wantCount {
					t.Fatalf("expected count %d, got %#v", tt.wantCount, cfg.Count)
				}
			}
		})
	}
}
