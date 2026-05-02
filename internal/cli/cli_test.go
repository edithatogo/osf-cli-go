package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestRunPrintsHelpWithoutArgs(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run(nil, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run returned %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "Open Science Framework") {
		t.Fatalf("help output did not describe OSF CLI: %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunPrintsHelpWithHelpFlag(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"--help"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run returned %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "Open Science Framework") {
		t.Fatalf("help output did not describe OSF CLI: %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunPrintsHelpWithShortHelpFlag(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"-h"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run returned %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "Open Science Framework") {
		t.Fatalf("help output did not describe OSF CLI: %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunPrintsVersion(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"--version"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run returned %d, want 0", code)
	}
	if got := strings.TrimSpace(stdout.String()); got != version {
		t.Fatalf("version output = %q, want %q", got, version)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunRejectsUnknownCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"nope"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("Run returned %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "unknown command") || !strings.Contains(stderr.String(), "nope") {
		t.Fatalf("stderr = %q, want unknown command message", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestRunRejectsInvalidOutputMode(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"--output", "csv"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("Run returned %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "invalid output mode") {
		t.Fatalf("stderr = %q, want invalid output mode message", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestRunRejectsConflictingOutputModes(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"--output", "table", "--json"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("Run returned %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "cannot combine --json") {
		t.Fatalf("stderr = %q, want conflicting output mode message", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestRunEmitsJSONContract(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"--output", "json"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run returned %d, want 0", code)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	var contract rootContract
	if err := json.Unmarshal(stdout.Bytes(), &contract); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\n%s", err, stdout.String())
	}

	if contract.Name != "osf" {
		t.Fatalf("contract name = %q, want osf", contract.Name)
	}
	if contract.Version != version {
		t.Fatalf("contract version = %q, want %q", contract.Version, version)
	}
	if contract.DefaultOutput != outputModeTable {
		t.Fatalf("default output = %q, want %q", contract.DefaultOutput, outputModeTable)
	}
	if len(contract.Commands) != 4 {
		t.Fatalf("command count = %d, want 4", len(contract.Commands))
	}
	if contract.ExitCodes["success"] != 0 || contract.ExitCodes["planned_command"] != 1 || contract.ExitCodes["usage_or_argument"] != 2 {
		t.Fatalf("unexpected exit code contract: %#v", contract.ExitCodes)
	}
}
