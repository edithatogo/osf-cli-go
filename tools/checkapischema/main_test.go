package main

import (
	"strings"
	"testing"
)

func validManifest() sourceManifest {
	commit := "52215db60cc8dd95f86841b467d2ed339fd67dec"
	return sourceManifest{
		SchemaVersion:    1,
		SourceRepository: "https://github.com/CenterForOpenScience/developer.osf.io",
		SourcePath:       "swagger-spec/swagger.yaml", SourceRef: "master", SourceCommit: commit,
		SourceURL:     "https://raw.githubusercontent.com/CenterForOpenScience/developer.osf.io/" + commit + "/swagger-spec/swagger.yaml",
		RetrievedDate: "2026-07-14", License: "CC BY-NC 4.0", Handling: "remote-source-manifest",
		Decision: "deferred", DeferIssue: "#46", TrackedTags: []string{"nodes"}, Rationale: "review",
	}
}

func TestValidateManifest(t *testing.T) {
	manifest := validManifest()
	if err := validateManifest(manifest); err != nil {
		t.Fatalf("validateManifest() error = %v", err)
	}
}

func TestValidateManifestRejectsUnpinnedSource(t *testing.T) {
	manifest := validManifest()
	manifest.SourceURL = strings.Replace(manifest.SourceURL, manifest.SourceCommit, "master", 1)
	if err := validateManifest(manifest); err == nil || !strings.Contains(err.Error(), "sourceUrl") {
		t.Fatalf("validateManifest() error = %v, want pinned URL failure", err)
	}
}
