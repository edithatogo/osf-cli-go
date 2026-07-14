package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunValidatesDigestBoundLevels(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	payload := []byte("Zenodo Sandbox live validation")
	if err := os.WriteFile(filepath.Join(root, "evidence.md"), payload, 0o600); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(payload)
	document := manifest{SchemaVersion: 1, GeneratedAt: "2026-07-15T00:00:00Z", Claims: []claim{{
		ID: "zenodo-transfer", Provider: "zenodo", Capability: "files.transfer", Level: "sandbox-validated", ValidatedAt: "2026-07-15",
		ResourceDisposition: "deleted",
		Evidence:            []evidence{{Path: "evidence.md", SHA256: "sha256:" + hex.EncodeToString(digest[:])}},
	}}}
	document.OptInWorkflow = writeWorkflow(t, root)
	writeManifest(t, root, document)
	if err := run(root, "manifest.json", time.Date(2026, 7, 15, 1, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
}

func TestRunRequiresSandboxResourceDisposition(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	payload := []byte("Zenodo Sandbox publication evidence")
	if err := os.WriteFile(filepath.Join(root, "evidence.md"), payload, 0o600); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(payload)
	base := claim{
		ID: "zenodo-publication", Provider: "zenodo", Capability: "publication", Level: "sandbox-validated", ValidatedAt: "2026-07-15",
		Evidence: []evidence{{Path: "evidence.md", SHA256: hex.EncodeToString(digest[:])}},
	}
	for name, mutate := range map[string]func(*claim){
		"missing disposition": func(*claim) {},
		"retained without record": func(value *claim) {
			value.ResourceDisposition = "published-retained"
		},
		"retained production record": func(value *claim) {
			value.ResourceDisposition = "published-retained"
			value.ResourceRecord = "https://zenodo.org/records/123"
		},
		"deleted with record": func(value *claim) {
			value.ResourceDisposition = "deleted"
			value.ResourceRecord = "https://sandbox.zenodo.org/records/123"
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			candidate := base
			mutate(&candidate)
			document := manifest{SchemaVersion: 1, GeneratedAt: "2026-07-15T00:00:00Z", OptInWorkflow: writeWorkflow(t, root), Claims: []claim{candidate}}
			writeManifest(t, root, document)
			if err := run(root, "manifest.json", time.Date(2026, 7, 15, 1, 0, 0, 0, time.UTC)); err == nil {
				t.Fatal("run returned nil error")
			}
		})
	}
}

func TestRunAcceptsRetainedSandboxPublication(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	payload := []byte("Zenodo Sandbox publication evidence")
	if err := os.WriteFile(filepath.Join(root, "evidence.md"), payload, 0o600); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(payload)
	document := manifest{SchemaVersion: 1, GeneratedAt: "2026-07-15T00:00:00Z", OptInWorkflow: writeWorkflow(t, root), Claims: []claim{{
		ID: "zenodo-publication", Provider: "zenodo", Capability: "publication", Level: "sandbox-validated", ValidatedAt: "2026-07-15",
		ResourceDisposition: "published-retained", ResourceRecord: "https://sandbox.zenodo.org/records/123",
		Evidence: []evidence{{Path: "evidence.md", SHA256: hex.EncodeToString(digest[:])}},
	}}}
	writeManifest(t, root, document)
	if err := run(root, "manifest.json", time.Date(2026, 7, 15, 1, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
}

func TestRunRejectsInflatedOrStaleClaims(t *testing.T) {
	t.Parallel()
	for name, mutate := range map[string]func(*manifest){
		"invalid level":              func(value *manifest) { value.Claims[0].Level = "live-validated" },
		"production without receipt": func(value *manifest) { value.Claims[0].Level = "production-validated" },
		"digest drift":               func(value *manifest) { value.Claims[0].Evidence[0].SHA256 = strings.Repeat("0", 64) },
		"path traversal":             func(value *manifest) { value.Claims[0].Evidence[0].Path = "../outside" },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			payload := []byte("offline evidence")
			if err := os.WriteFile(filepath.Join(root, "evidence.md"), payload, 0o600); err != nil {
				t.Fatal(err)
			}
			digest := sha256.Sum256(payload)
			document := manifest{SchemaVersion: 1, GeneratedAt: "2026-07-15T00:00:00Z", Claims: []claim{{
				ID: "osf-contract", Provider: "osf", Capability: "contract", Level: "offline-tested", ValidatedAt: "2026-07-15",
				Evidence: []evidence{{Path: "evidence.md", SHA256: hex.EncodeToString(digest[:])}},
			}}}
			document.OptInWorkflow = writeWorkflow(t, root)
			mutate(&document)
			writeManifest(t, root, document)
			if err := run(root, "manifest.json", time.Date(2026, 7, 15, 1, 0, 0, 0, time.UTC)); err == nil {
				t.Fatal("run returned nil error")
			}
		})
	}
}

func TestWriteReportIsReproducible(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	payload := []byte("offline evidence")
	if err := os.WriteFile(filepath.Join(root, "evidence.md"), payload, 0o600); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(payload)
	document := manifest{SchemaVersion: 1, GeneratedAt: "2026-07-15T00:00:00Z", OptInWorkflow: writeWorkflow(t, root), Claims: []claim{{
		ID: "osf-contract", Provider: "osf", Capability: "contract", Level: "offline-tested", ValidatedAt: "2026-07-15",
		Evidence: []evidence{{Path: "evidence.md", SHA256: hex.EncodeToString(digest[:])}},
	}}}
	writeManifest(t, root, document)
	if err := writeReport(root, "manifest.json", "report.md"); err != nil {
		t.Fatal(err)
	}
	first, _ := os.ReadFile(filepath.Join(root, "report.md"))
	if err := writeReport(root, "manifest.json", "report.md"); err != nil {
		t.Fatal(err)
	}
	second, _ := os.ReadFile(filepath.Join(root, "report.md"))
	if string(first) != string(second) || !strings.Contains(string(first), "Production-validated claims: 0") {
		t.Fatalf("report = %s", first)
	}
}

func writeWorkflow(t *testing.T, root string) string {
	t.Helper()
	name := "provider-validation.yml"
	content := `on:
  workflow_dispatch:
env:
  ZENODO_SANDBOX_VALIDATION: 1
  ZENODO_PUBLICATION_VALIDATION: 1
  CROSS_PROVIDER_SANDBOX_VALIDATION: 1
  OSF_LIVE_VALIDATION: 1
  OSF_VALIDATE_WRITES: 1
  ZENODO_SANDBOX_TOKEN: secret
  ZENODO_SANDBOX_PUBLICATION_TOKEN: secret
  OSF_VALIDATION_TOKEN: secret
  OSF_VALIDATE_PROJECT: secret
`
	if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return name
}

func writeManifest(t *testing.T, root string, document manifest) {
	t.Helper()
	payload, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "manifest.json"), payload, 0o600); err != nil {
		t.Fatal(err)
	}
}
