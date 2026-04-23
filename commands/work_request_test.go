package commands

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	ociFunctions "github.com/oracle/oci-go-sdk/v65/functions"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdout pipe: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	outC := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		outC <- buf.String()
	}()

	fn()
	_ = w.Close()
	output := <-outC
	_ = r.Close()
	return output
}

func TestIsWorkRequestID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "valid functions work request ocid", input: "ocid1.functionsworkrequest.oc1..exampleuniqueID", want: true},
		{name: "valid mixed case trimmed", input: "  OCID1.FunctionsWorkRequest.oc1..exampleuniqueID  ", want: true},
		{name: "normal function name", input: "hello", want: false},
		{name: "plain ocid without workrequest", input: "ocid1.fnfunc.oc1..exampleuniqueID", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isWorkRequestID(tc.input)
			if got != tc.want {
				t.Fatalf("isWorkRequestID(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestSimplifyWorkRequestOperation(t *testing.T) {
	tests := []struct {
		name  string
		input ociFunctions.OperationTypeEnum
		want  string
	}{
		{name: "create function", input: ociFunctions.OperationTypeEnum("CREATE_FUNCTION"), want: "CREATE"},
		{name: "update function", input: ociFunctions.OperationTypeEnum("UPDATE_FUNCTION"), want: "UPDATE"},
		{name: "blank", input: ociFunctions.OperationTypeEnum(""), want: "UNKNOWN"},
		{name: "single token", input: ociFunctions.OperationTypeEnum("DELETE"), want: "DELETE"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := simplifyWorkRequestOperation(tc.input)
			if got != tc.want {
				t.Fatalf("simplifyWorkRequestOperation(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestExtractFunctionResourceID(t *testing.T) {
	functionEntityType := "function"
	functionID := "ocid1.fnfunc.oc1..exampleuniqueID"
	otherEntityType := "application"
	otherID := "ocid1.fnapp.oc1..otherID"

	t.Run("prefers function resource identifiers", func(t *testing.T) {
		wr := ociFunctions.WorkRequest{
			Resources: []ociFunctions.WorkRequestResource{
				{EntityType: &otherEntityType, Identifier: &otherID},
				{EntityType: &functionEntityType, Identifier: &functionID},
			},
		}
		got := extractFunctionResourceID(wr)
		if got != functionID {
			t.Fatalf("extractFunctionResourceID() = %q, want %q", got, functionID)
		}
	})

	t.Run("falls back to first identifier if no function entity exists", func(t *testing.T) {
		wr := ociFunctions.WorkRequest{
			Resources: []ociFunctions.WorkRequestResource{
				{EntityType: &otherEntityType, Identifier: &otherID},
			},
		}
		got := extractFunctionResourceID(wr)
		if got != otherID {
			t.Fatalf("extractFunctionResourceID() = %q, want %q", got, otherID)
		}
	})

	t.Run("returns empty when no identifiers exist", func(t *testing.T) {
		wr := ociFunctions.WorkRequest{}
		got := extractFunctionResourceID(wr)
		if got != "" {
			t.Fatalf("extractFunctionResourceID() = %q, want empty string", got)
		}
	})
}

func TestPrintWorkRequestStatusView(t *testing.T) {
	view := &workRequestStatusView{
		WorkRequestID: "ocid1.functionsworkrequest.oc1..exampleuniqueID",
		FunctionName:  "hello",
		Operation:     "CREATE",
		Status:        "SUCCEEDED",
		Error:         "",
		RecentLogs:    []string{"accepted", "completed"},
	}

	output := captureStdout(t, func() {
		printWorkRequestStatusView(view)
	})

	checks := []string{
		"Work Request: ocid1.functionsworkrequest.oc1..exampleuniqueID",
		"Function: hello",
		"Operation: CREATE",
		"Status: SUCCEEDED",
		"Recent Logs:",
		"- accepted",
		"- completed",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Fatalf("expected output to contain %q, got: %s", check, output)
		}
	}
}

func TestWorkRequestCommandRegistration(t *testing.T) {
	cmd := WorkRequestCommand()
	if cmd.Name != "work-request" {
		t.Fatalf("command name = %q, want %q", cmd.Name, "work-request")
	}
	if len(cmd.Subcommands) == 0 {
		t.Fatal("expected work-request command to have subcommands")
	}
	if cmd.Subcommands[0].Name != "status" {
		t.Fatalf("first subcommand name = %q, want %q", cmd.Subcommands[0].Name, "status")
	}

	registered, ok := Commands["work-request"]
	if !ok {
		t.Fatal("expected work-request command to be registered in Commands map")
	}
	if registered.Name != "work-request" {
		t.Fatalf("registered command name = %q, want %q", registered.Name, "work-request")
	}
}