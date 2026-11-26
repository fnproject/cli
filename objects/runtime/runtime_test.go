package runtime

import (
	"testing"
	"time"

	ociCommon "github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/functions"
)

func TestSelectCodeOnlyRuntimeName(t *testing.T) {
	items := []functions.FunctionsRuntimeSummary{
		runtimeSummary("java17.ol9", "java 17", "2026-07-20T14:20:50Z", functions.FunctionsRuntimeLifecycleStateActive),
		runtimeSummary("java21.ol9", "java 21", "2027-01-17T14:20:50Z", functions.FunctionsRuntimeLifecycleStateActive),
		runtimeSummary("java22.ol9", "java 22", "2029-12-31T23:59:59Z", functions.FunctionsRuntimeLifecycleStateActive),
		runtimeSummary("node24.ol9", "node 24", "2027-01-17T14:20:50Z", functions.FunctionsRuntimeLifecycleStateActive),
		runtimeSummary("python311.ol9", "python 311", "2026-04-19T13:44:00Z", functions.FunctionsRuntimeLifecycleStateActive),
		runtimeSummary("python312.ol9", "python 3.12", "2027-01-17T14:20:50Z", functions.FunctionsRuntimeLifecycleStateActive),
		runtimeSummary("ol9", "-", "2027-01-17T14:20:50Z", functions.FunctionsRuntimeLifecycleStateActive),
		runtimeSummary("ol8", "-", "2030-01-01T00:00:00Z", functions.FunctionsRuntimeLifecycleStateActive),
		runtimeSummary("java99.ol9", "java 99", "2035-01-01T00:00:00Z", functions.FunctionsRuntimeLifecycleStateInactive),
	}

	for alias, want := range map[string]string{
		"java":       "java22.ol9",
		"java21":     "java21.ol9",
		"node":       "node24.ol9",
		"node24":     "node24.ol9",
		"python":     "python312.ol9",
		"python3.12": "python312.ol9",
		"go":         "ol9",
		"go1.24":     "ol9",
	} {
		got, err := SelectCodeOnlyRuntimeName(alias, items)
		if err != nil {
			t.Fatalf("SelectCodeOnlyRuntimeName(%q) returned error: %v", alias, err)
		}
		if got != want {
			t.Fatalf("SelectCodeOnlyRuntimeName(%q) = %q, want %q", alias, got, want)
		}
	}
}

func TestSelectCodeOnlyRuntimeNameReturnsHelpfulErrorWhenNoMatch(t *testing.T) {
	_, err := SelectCodeOnlyRuntimeName("java", nil)
	if err == nil || err.Error() != "no active managed runtime found for language \"java\"; run `fn list runtimes` to view supported runtimes" {
		t.Fatalf("error = %v", err)
	}
}

func TestValidateCodeOnlyRuntimeName(t *testing.T) {
	items := []functions.FunctionsRuntimeSummary{
		runtimeSummary("java21.ol9", "java 21", "2027-01-17T14:20:50Z", functions.FunctionsRuntimeLifecycleStateActive),
		runtimeSummary("java22.ol9", "java 22", "2029-12-31T23:59:59Z", functions.FunctionsRuntimeLifecycleStateInactive),
	}

	got, err := ValidateCodeOnlyRuntimeName("JAVA21.OL9", items)
	if err != nil || got != "java21.ol9" {
		t.Fatalf("ValidateCodeOnlyRuntimeName() = %q, %v; want java21.ol9, nil", got, err)
	}

	_, err = ValidateCodeOnlyRuntimeName("jvaa22", items)
	want := "no active managed runtime named \"jvaa22\"; run `fn list runtimes` to view supported runtimes"
	if err == nil || err.Error() != want {
		t.Fatalf("error = %v, want %q", err, want)
	}
}

func TestIsExplicitCodeOnlyRuntimeName(t *testing.T) {
	for name, want := range map[string]bool{
		"ol9":              true,
		"OL8":              true,
		"java21.ol9":       true,
		"node24.ol9":       true,
		"python311.ol9":    true,
		"java21":           true,
		"go":               false,
		"java":             false,
		"python":           false,
		"jvaa21.ol9":       false,
		"java21.al2023":    false,
		"java21.ol9.extra": false,
	} {
		if got := isExplicitCodeOnlyRuntimeName(name); got != want {
			t.Fatalf("isExplicitCodeOnlyRuntimeName(%q) = %t, want %t", name, got, want)
		}
	}
}

func runtimeSummary(name, language, deprecated string, state functions.FunctionsRuntimeLifecycleStateEnum) functions.FunctionsRuntimeSummary {
	deprecationTime, err := time.Parse(time.RFC3339, deprecated)
	if err != nil {
		panic(err)
	}
	return functions.FunctionsRuntimeSummary{
		Name:           &name,
		Language:       &language,
		LifecycleState: state,
		TimeDeprecated: &ociCommon.SDKTime{Time: deprecationTime},
	}
}
