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
		Evidence: []evidence{{Path: "evidence.md", SHA256: "sha256:" + hex.EncodeToString(digest[:])}},
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
			mutate(&document)
			writeManifest(t, root, document)
			if err := run(root, "manifest.json", time.Date(2026, 7, 15, 1, 0, 0, 0, time.UTC)); err == nil {
				t.Fatal("run returned nil error")
			}
		})
	}
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
