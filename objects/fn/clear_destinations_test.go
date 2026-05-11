package fn

import (
	"flag"
	"testing"

	models "github.com/fnproject/fn_go/modelsv2"
	"github.com/urfave/cli"
)

func TestValidateDestinationFlagCombination(t *testing.T) {
	set := flag.NewFlagSet("test", flag.ContinueOnError)
	_ = set.Bool("clear-on-success", true, "")
	_ = set.String("on-success", "stream:ocid1.stream.oc1..abc", "")
	ctx := cli.NewContext(nil, set, nil)
	if _, err := validateDestinationFlagCombination(ctx); err == nil {
		t.Fatal("expected conflict error for on-success + clear-on-success")
	}
}

func TestSetClearDestinationAnnotations(t *testing.T) {
	fn := &models.Fn{Annotations: map[string]interface{}{}}
	SetClearDestinationAnnotations(fn, true, true)
	if fn.Annotations[annotationSuccessDestinationKind] != "NONE" {
		t.Fatalf("expected success kind NONE, got %#v", fn.Annotations[annotationSuccessDestinationKind])
	}
	if fn.Annotations[annotationFailureDestinationKind] != "NONE" {
		t.Fatalf("expected failure kind NONE, got %#v", fn.Annotations[annotationFailureDestinationKind])
	}
}