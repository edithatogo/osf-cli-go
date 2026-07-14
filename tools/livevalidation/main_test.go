package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/edithatogo/osf-cli-go/internal/auth"
)

type mapSource map[string]string

func (m mapSource) Lookup(name string) (string, bool) {
	v, ok := m[name]
	return v, ok
}

func TestLoadValidationEnv(t *testing.T) {
	t.Parallel()

	env := loadValidationEnv(mapSource{
		"OSF_LIVE_VALIDATION":  "yes",
		"OSF_VALIDATE_WRITES":  "true",
		auth.TokenEnv:          "token-value",
		"OSF_VALIDATE_PROJECT": "https://osf.io/abcd1/",
	})
	if !env.liveEnabled {
		t.Fatalf("liveEnabled = false, want true")
	}
	if !env.writesEnabled {
		t.Fatalf("writesEnabled = false, want true")
	}
	if env.token != "token-value" {
		t.Fatalf("token = %q, want token-value", env.token)
	}
	if env.projectRef != "https://osf.io/abcd1/" {
		t.Fatalf("projectRef = %q, want project url", env.projectRef)
	}
}

func TestDefaultEvidencePathMatchesCurrentTrack(t *testing.T) {
	t.Parallel()

	want := filepath.Join("docs", "live-osf-validation-evidence.md")
	if defaultEvidencePath != want {
		t.Fatalf("defaultEvidencePath = %q, want %q", defaultEvidencePath, want)
	}
}

func TestRunValidationDryRunPlansAllSteps(t *testing.T) {
	t.Parallel()

	report, err := runValidation(t.Context(), validationEnv{}, false, time.Second)
	if err != nil {
		t.Fatalf("runValidation returned error: %v", err)
	}
	if report.Mode != "dry-run" {
		t.Fatalf("mode = %q, want dry-run", report.Mode)
	}
	if len(report.Steps) == 0 {
		t.Fatal("expected planned steps")
	}
	if got := report.Steps[len(report.Steps)-1].Status; got != "pending" {
		t.Fatalf("last step status = %q, want pending", got)
	}
}

func TestMissingEnvMessage(t *testing.T) {
	t.Parallel()

	got := missingEnvMessage(validationEnv{})
	for _, want := range []string{auth.TokenEnv, "OSF_VALIDATE_PROJECT"} {
		if !strings.Contains(got, want) {
			t.Fatalf("message %q missing %q", got, want)
		}
	}
}

func TestWriteEvidenceRedactsInputs(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "evidence.md")
	report := validationReport{
		GeneratedAt: time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC),
		Mode:        "dry-run",
		Env: validationEnv{
			liveEnabled: true,
			token:       "super-secret-token",
			projectRef:  "https://osf.io/abcd1/",
		},
		Steps: plannedSteps(validationEnv{
			liveEnabled: true,
			token:       "super-secret-token",
			projectRef:  "https://osf.io/abcd1/",
		}),
	}

	if err := writeEvidence(path, report); err != nil {
		t.Fatalf("writeEvidence: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read evidence: %v", err)
	}
	text := string(content)
	for _, forbidden := range []string{"super-secret-token", "abcd1"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("evidence leaked %q: %s", forbidden, text)
		}
	}
	for _, want := range []string{"OSF_TOKEN: set", "OSF_VALIDATE_PROJECT: set", "files download: pending", "search: planned", "preprints list: planned"} {
		if !strings.Contains(text, want) {
			t.Fatalf("evidence missing %q: %s", want, text)
		}
	}
}

func TestWriteEvidenceLabelsSkippedStepsAccurately(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "evidence.md")
	report := validationReport{
		GeneratedAt: time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC),
		Mode:        "skipped",
		Steps:       plannedSteps(validationEnv{}),
	}
	if err := writeEvidence(path, report); err != nil {
		t.Fatalf("writeEvidence: %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read evidence: %v", err)
	}
	text := string(content)
	if !strings.Contains(text, "not executed; live validation was skipped") {
		t.Fatalf("evidence did not identify skipped execution: %s", text)
	}
	if strings.Contains(text, "not executed in dry-run mode") {
		t.Fatalf("evidence mislabeled skipped execution as dry-run: %s", text)
	}
}

func TestExecutableStepsGateFileDownload(t *testing.T) {
	t.Parallel()

	steps := executableSteps(validationEnv{projectRef: "xj6qc"})
	for _, step := range steps {
		if step.Name == "files download" && step.Executable {
			t.Fatalf("files download executable without download ref: %+v", step)
		}
	}

	steps = executableSteps(validationEnv{projectRef: "xj6qc", downloadRef: "file-1"})
	for _, step := range steps {
		if step.Name == "files download" {
			if !step.Executable {
				t.Fatalf("files download not executable with download ref: %+v", step)
			}
			if !strings.Contains(step.Command, "file-1") {
				t.Fatalf("download command = %q, want file ref", step.Command)
			}
			return
		}
	}
	t.Fatal("files download step not found")
}

func TestRunValidationExercisesWriteCancellationAndMCP(t *testing.T) {
	env := validationEnv{liveEnabled: true, writesEnabled: true, token: "token", projectRef: "xj6qc", downloadRef: "file-1"}
	var commands []string
	osfRunner := func(_ context.Context, timeout time.Duration, _ validationEnv, step validationStep) (string, error) {
		commands = append(commands, step.Command)
		if timeout == cancellationProbeTimeout {
			return "", context.DeadlineExceeded
		}
		if strings.Contains(step.Command, "--conflict fail") {
			return "file already exists", errors.New("request failed")
		}
		return "ok", nil
	}
	mcpCalled := false
	mcpRunner := func(_ context.Context, _ time.Duration, got validationEnv) (string, error) {
		mcpCalled = got.projectRef == env.projectRef
		return "MCP osf_project_get returned structured content", nil
	}

	report, err := runValidationWithRunners(t.Context(), env, true, time.Second, osfRunner, mcpRunner)
	if err != nil {
		t.Fatalf("runValidationWithRunners: %v", err)
	}
	if !mcpCalled {
		t.Fatal("MCP runner was not called")
	}
	for _, name := range []string{"files upload", "files upload conflict", "cancellation", "MCP project get", "files cleanup"} {
		result := findResult(report.Steps, name)
		if result == nil || result.Status != "passed" {
			t.Fatalf("step %q = %+v, want passed", name, result)
		}
	}
	joined := strings.Join(commands, "\n")
	for _, want := range []string{"files upload --node xj6qc", "files rm --node xj6qc", "--yes"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("commands missing %q: %s", want, joined)
		}
	}
}

func TestRunValidationRejectsUnrelatedConflictProbeFailure(t *testing.T) {
	env := validationEnv{liveEnabled: true, writesEnabled: true, token: "token", projectRef: "xj6qc", downloadRef: "file-1"}
	osfRunner := func(_ context.Context, timeout time.Duration, _ validationEnv, step validationStep) (string, error) {
		if timeout == cancellationProbeTimeout {
			return "", context.DeadlineExceeded
		}
		if strings.Contains(step.Command, "--conflict fail") {
			return "request unauthorized", errors.New("exit status 1")
		}
		return "private project output", nil
	}
	mcpRunner := func(context.Context, time.Duration, validationEnv) (string, error) { return "ok", nil }

	report, err := runValidationWithRunners(t.Context(), env, true, time.Second, osfRunner, mcpRunner)
	if err == nil || !strings.Contains(err.Error(), "files upload conflict") {
		t.Fatalf("error = %v, want conflict probe failure", err)
	}
	result := findResult(report.Steps, "files upload conflict")
	if result == nil || result.Status != "failed" || strings.Contains(result.Output, "unauthorized") {
		t.Fatalf("result = %+v, want sanitized failure", result)
	}
	for _, result := range report.Steps {
		if strings.Contains(result.Output, "private project output") {
			t.Fatalf("result leaked command output: %+v", result)
		}
	}
}

func TestRunValidationAttemptsCleanupAfterUploadFailure(t *testing.T) {
	env := validationEnv{liveEnabled: true, writesEnabled: true, token: "token", projectRef: "xj6qc", downloadRef: "file-1"}
	cleanupCalled := false
	osfRunner := func(_ context.Context, _ time.Duration, _ validationEnv, step validationStep) (string, error) {
		if strings.HasPrefix(step.Command, "files upload") {
			return "upload failed", errors.New("upload failed")
		}
		if strings.HasPrefix(step.Command, "files rm") {
			cleanupCalled = true
		}
		return "ok", nil
	}
	mcpRunner := func(context.Context, time.Duration, validationEnv) (string, error) { return "ok", nil }

	report, err := runValidationWithRunners(t.Context(), env, true, time.Second, osfRunner, mcpRunner)
	if err == nil {
		t.Fatal("runValidationWithRunners returned nil error")
	}
	if !cleanupCalled {
		t.Fatal("cleanup was not attempted after upload failure")
	}
	path := filepath.Join(t.TempDir(), "failed-evidence.md")
	if err := writeEvidence(path, report); err != nil {
		t.Fatalf("writeEvidence: %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil || !strings.Contains(string(content), "files upload: failed") || !strings.Contains(string(content), "files cleanup: passed") {
		t.Fatalf("failed evidence = %q, err=%v", content, err)
	}
}

func TestRunValidationRejectsIncompleteLiveCoverage(t *testing.T) {
	env := validationEnv{liveEnabled: true, token: "token", projectRef: "xj6qc"}
	osfRunner := func(context.Context, time.Duration, validationEnv, validationStep) (string, error) { return "ok", nil }
	mcpRunner := func(context.Context, time.Duration, validationEnv) (string, error) { return "ok", nil }
	report, err := runValidationWithRunners(t.Context(), env, true, time.Second, osfRunner, mcpRunner)
	if err == nil || !strings.Contains(err.Error(), "incomplete") {
		t.Fatalf("error = %v, want incomplete live validation", err)
	}
	if result := findResult(report.Steps, "files upload"); result == nil || result.Status != "pending" || !strings.Contains(result.Output, "OSF_VALIDATE_WRITES") {
		t.Fatalf("write result = %+v", result)
	}
	if result := findResult(report.Steps, "files download"); result == nil || result.Status != "pending" || !strings.Contains(result.Output, "OSF_VALIDATE_DOWNLOAD") {
		t.Fatalf("download result = %+v", result)
	}
}

func findResult(results []validationResult, name string) *validationResult {
	for i := range results {
		if results[i].Step.Name == name {
			return &results[i]
		}
	}
	return nil
}
