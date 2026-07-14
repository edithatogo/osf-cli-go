package main

import (
	"strings"
	"testing"
	"time"
)

func TestOfflineValidationEvidence(t *testing.T) {
	t.Setenv(liveOptIn, "")
	result, err := run(t.Context(), false)
	if err != nil {
		t.Fatal(err)
	}
	if result.ValidationLevel != "integration-ready" || result.DraftCleanup != "not-created" {
		t.Fatalf("result = %+v", result)
	}
	markdown := result.markdown()
	if !strings.Contains(markdown, "Publication is irreversible") || !strings.Contains(markdown, "user:email excluded") || strings.Contains(markdown, "ZENODO_TOKEN") {
		t.Fatalf("evidence = %s", markdown)
	}
}

func TestLiveValidationFailsClosedWithoutOptIn(t *testing.T) {
	t.Setenv(liveOptIn, "")
	if _, err := run(t.Context(), true); err == nil {
		t.Fatal("live validation returned nil error")
	}
}

func TestValidationMetadataIsPublishable(t *testing.T) {
	metadata := validationMetadata(time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC))
	if metadata.Title == "" || metadata.License != "cc-by-4.0" || len(metadata.Creators) != 1 {
		t.Fatalf("metadata = %+v", metadata)
	}
}
