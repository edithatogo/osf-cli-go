package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunPassesWithRepoFixtures(t *testing.T) {
	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}

	dir := t.TempDir()
	copyFixture(t, filepath.Join("..", "..", "server.json"), filepath.Join(dir, "server.json"))
	copyFixture(t, filepath.Join("..", "..", "registry", "directory-submissions.json"), filepath.Join(dir, "registry", "directory-submissions.json"))
	copyFixture(t, filepath.Join("..", "..", "packaging", "mcpb", "manifest.json"), filepath.Join(dir, "packaging", "mcpb", "manifest.json"))
	copyFixture(t, filepath.Join("..", "..", "registry", "README.md"), filepath.Join(dir, "registry", "README.md"))
	copyFixture(t, filepath.Join("..", "..", "glama.json"), filepath.Join(dir, "glama.json"))
	copyFixture(t, filepath.Join("..", "..", "registry", "submission-scorecard.json"), filepath.Join(dir, "registry", "submission-scorecard.json"))
	copyFixture(t, filepath.Join("..", "..", "docs", "mcp-quality-report.json"), filepath.Join(dir, "docs", "mcp-quality-report.json"))

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldwd); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})

	if err := run(); err != nil {
		t.Fatalf("run() returned error: %v", err)
	}
}

func TestRunFailsWhenOfficialRegistryURLIsWrong(t *testing.T) {
	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}

	dir := t.TempDir()
	copyFixture(t, filepath.Join("..", "..", "server.json"), filepath.Join(dir, "server.json"))
	copyFixture(t, filepath.Join("..", "..", "registry", "directory-submissions.json"), filepath.Join(dir, "registry", "directory-submissions.json"))
	copyFixture(t, filepath.Join("..", "..", "packaging", "mcpb", "manifest.json"), filepath.Join(dir, "packaging", "mcpb", "manifest.json"))
	copyFixture(t, filepath.Join("..", "..", "registry", "README.md"), filepath.Join(dir, "registry", "README.md"))
	copyFixture(t, filepath.Join("..", "..", "glama.json"), filepath.Join(dir, "glama.json"))
	copyFixture(t, filepath.Join("..", "..", "registry", "submission-scorecard.json"), filepath.Join(dir, "registry", "submission-scorecard.json"))
	copyFixture(t, filepath.Join("..", "..", "docs", "mcp-quality-report.json"), filepath.Join(dir, "docs", "mcp-quality-report.json"))

	path := filepath.Join(dir, "registry", "directory-submissions.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	updated := strings.Replace(string(data), "https://registry.modelcontextprotocol.io/v0.1/servers?search=io.github.edithatogo%2Fosf-cli-go", "https://example.invalid", 1)
	if updated == string(data) {
		t.Fatal("fixture update did not change the registry URL")
	}
	if err := os.WriteFile(path, []byte(updated), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldwd); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})

	if err := run(); err == nil || !strings.Contains(err.Error(), "officialRegistryUrl") {
		t.Fatalf("run() error = %v, want officialRegistryUrl validation failure", err)
	}
}

func TestSubmissionScorecardRejectsPublishedWithoutReceipt(t *testing.T) {
	scorecard := submissionScorecard{
		SchemaVersion: 1,
		ReviewedDate:  "2026-07-14",
		Version:       "0.3.2",
		Targets: []scorecardTarget{{
			ID:          "openai-codex-cowork",
			Name:        "OpenAI Codex/Cowork",
			State:       "published",
			Score:       100,
			ScoreTarget: 100,
			Evidence:    []string{"evidence"},
			NextAction:  "next",
		}},
	}
	if err := checkSubmissionScorecard(scorecard, "0.3.2"); err == nil || !strings.Contains(err.Error(), "without receipt") {
		t.Fatalf("checkSubmissionScorecard() error = %v, want missing receipt failure", err)
	}
}

func copyFixture(t *testing.T, src, dst string) {
	t.Helper()

	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", src, err)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		t.Fatalf("MkdirAll(%s): %v", dst, err)
	}
	if err := os.WriteFile(dst, data, 0o600); err != nil {
		t.Fatalf("WriteFile(%s): %v", dst, err)
	}
}
