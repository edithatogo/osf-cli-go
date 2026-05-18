package main

import (
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
		auth.TokenEnv:          "token-value",
		"OSF_VALIDATE_PROJECT": "https://osf.io/abcd1/",
	})
	if !env.liveEnabled {
		t.Fatalf("liveEnabled = false, want true")
	}
	if env.token != "token-value" {
		t.Fatalf("token = %q, want token-value", env.token)
	}
	if env.projectRef != "https://osf.io/abcd1/" {
		t.Fatalf("projectRef = %q, want project url", env.projectRef)
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
