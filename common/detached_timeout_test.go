package common

import "testing"

func TestParseDetachedTimeoutSpec(t *testing.T) {
	tests := []struct {
		name        string
		spec        string
		wantSpec    string
		wantSeconds int
		wantErr     bool
	}{
		{name: "empty", spec: "", wantSpec: "", wantSeconds: 0},
		{name: "minutes", spec: "20m", wantSpec: "20m", wantSeconds: 1200},
		{name: "hours", spec: "1h", wantSpec: "1h", wantSeconds: 3600},
		{name: "trimmed", spec: " 30m ", wantSpec: "30m", wantSeconds: 1800},
		{name: "invalid", spec: "banana", wantErr: true},
		{name: "non-positive", spec: "0s", wantErr: true},
		{name: "sub-second", spec: "1500ms", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSpec, gotSeconds, err := ParseDetachedTimeoutSpec(tt.spec)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseDetachedTimeoutSpec() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if gotSpec != tt.wantSpec || gotSeconds != tt.wantSeconds {
				t.Fatalf("expected (%q, %d), got (%q, %d)", tt.wantSpec, tt.wantSeconds, gotSpec, gotSeconds)
			}
		})
	}
}