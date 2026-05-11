package common

import "testing"

func TestParseOCIDestinationSpec(t *testing.T) {
	tests := []struct {
		name     string
		spec     string
		wantNil  bool
		wantErr  bool
		wantType string
		wantOCID string
	}{
		{name: "empty", spec: "", wantNil: true},
		{name: "stream", spec: "stream:ocid1.stream.oc1..abc", wantType: DestinationTypeStream, wantOCID: "ocid1.stream.oc1..abc"},
		{name: "queue", spec: "queue:ocid1.queue.oc1..abc", wantType: DestinationTypeQueue, wantOCID: "ocid1.queue.oc1..abc"},
		{name: "notifications alias", spec: "notification:ocid1.onstopic.oc1..abc", wantType: DestinationTypeNotifications, wantOCID: "ocid1.onstopic.oc1..abc"},
		{name: "notifications plural", spec: "notifications:ocid1.onstopic.oc1..abc", wantType: DestinationTypeNotifications, wantOCID: "ocid1.onstopic.oc1..abc"},
		{name: "invalid missing ocid", spec: "stream:", wantErr: true},
		{name: "invalid missing colon", spec: "stream", wantErr: true},
		{name: "invalid type", spec: "topic:abc", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dest, err := ParseOCIDestinationSpec("--on-success", tt.spec)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseOCIDestinationSpec() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantNil {
				if dest != nil {
					t.Fatalf("expected nil destination, got %#v", dest)
				}
				return
			}
			if tt.wantErr {
				return
			}
			if dest == nil {
				t.Fatal("expected non-nil destination")
			}
			if dest.Type != tt.wantType || dest.OCID != tt.wantOCID {
				t.Fatalf("expected (%q, %q), got (%q, %q)", tt.wantType, tt.wantOCID, dest.Type, dest.OCID)
			}
		})
	}
}