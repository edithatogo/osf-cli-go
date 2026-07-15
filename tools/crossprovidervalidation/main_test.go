package main

import (
	"context"
	"strings"
	"testing"
)

func TestOfflineEvidenceIsIntegrationReadyAndUnpublished(t *testing.T) {
	t.Setenv(liveOptIn, "")
	result, err := run(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	if result.Level != "integration-ready" || result.Published || result.IdempotencyKey == "" || result.DeclaredScopes != "deposit:write" {
		t.Fatalf("result = %+v", result)
	}
	markdown := result.markdown()
	if !strings.Contains(markdown, "Published: false") || strings.Contains(markdown, "deposit:actions") {
		t.Fatalf("markdown = %s", markdown)
	}
}

func TestLiveRequiresExplicitOptInBeforeCredentials(t *testing.T) {
	t.Setenv(liveOptIn, "")
	t.Setenv("ZENODO_TOKEN", "")
	_, err := run(context.Background(), true)
	if err == nil || !strings.Contains(err.Error(), liveOptIn) {
		t.Fatalf("error = %v", err)
	}
}
