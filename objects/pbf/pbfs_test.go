package pbf

import (
	"testing"
	"time"

	ocifunctions "github.com/oracle/oci-go-sdk/v65/functions"
)

func TestFormatListingTriggers(t *testing.T) {
	name1 := "http"
	name2 := "objectstorage"
	got := formatListingTriggers([]ocifunctions.Trigger{{Name: &name1}, {Name: &name2}})
	if got != "http,objectstorage" {
		t.Fatalf("expected trigger string, got %q", got)
	}
}

func TestFormatSDKTime(t *testing.T) {
	ts := time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)
	got := formatSDKTime(&ts)
	if got == "" {
		t.Fatal("expected non-empty formatted time")
	}
}

func TestIsOCID(t *testing.T) {
	if !isOCID("ocid1.pbflisting.oc1..example") {
		t.Fatal("expected ocid to be recognized")
	}
	if isOCID("hello-world") {
		t.Fatal("did not expect non-ocid to be recognized")
	}
}
