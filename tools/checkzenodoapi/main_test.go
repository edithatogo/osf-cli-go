package main

import (
	"strings"
	"testing"
)

func TestValidateManifestAcceptsReviewedContract(t *testing.T) {
	m := validManifest()
	m.SnapshotSHA256 = snapshotDigest(m.Snapshot)
	if err := validateManifest(m); err != nil {
		t.Fatalf("validateManifest() error = %v", err)
	}
}

func TestValidateManifestRejectsDigestDrift(t *testing.T) {
	m := validManifest()
	m.SnapshotSHA256 = snapshotDigest(m.Snapshot)
	m.Snapshot.Limits.OAIPageSize++
	err := validateManifest(m)
	if err == nil || (!strings.Contains(err.Error(), "limit snapshot changed") && !strings.Contains(err.Error(), "snapshotSha256")) {
		t.Fatalf("validateManifest() error = %v, want reviewed drift error", err)
	}
}

func TestValidateManifestRejectsUnapprovedSource(t *testing.T) {
	m := validManifest()
	m.Sources[0].URL = "https://example.com/developers"
	m.SnapshotSHA256 = snapshotDigest(m.Snapshot)
	err := validateManifest(m)
	if err == nil || !strings.Contains(err.Error(), "unapproved official host") {
		t.Fatalf("validateManifest() error = %v, want source authority error", err)
	}
}

func TestValidateCapabilitiesRejectsUnsortedEntries(t *testing.T) {
	m := validManifest()
	m.Snapshot.Capabilities[0], m.Snapshot.Capabilities[1] = m.Snapshot.Capabilities[1], m.Snapshot.Capabilities[0]
	err := validateCapabilities(m.Snapshot.Capabilities)
	if err == nil || !strings.Contains(err.Error(), "sorted") {
		t.Fatalf("validateCapabilities() error = %v, want ordering error", err)
	}
}

func TestValidateMarkersReportsActionableDrift(t *testing.T) {
	s := source{ID: "developers", RequiredMarkers: []string{"deposit:write", "/api/records"}}
	err := validateMarkers(s, []byte("<html><body><code>deposit:write</code></body></html>"))
	if err == nil || !strings.Contains(err.Error(), "/api/records") || !strings.Contains(err.Error(), "review upstream") {
		t.Fatalf("validateMarkers() error = %v, want missing marker guidance", err)
	}
}

func validManifest() manifest {
	m := manifest{
		SchemaVersion:   1,
		ReviewedDate:    "2026-07-15",
		VersionPolicy:   "documentation-date",
		VersionDecision: "The official API has no published semantic version.",
		Sources: []source{
			{ID: "developers", Kind: "api-documentation", URL: "https://developers.zenodo.org/", RetrievedDate: "2026-07-15", RequiredMarkers: []string{"records", "depositions"}},
			{ID: "sandbox", Kind: "sandbox-guidance", URL: "https://help.zenodo.org/docs/get-started/", RetrievedDate: "2026-07-15", RequiredMarkers: []string{"sandbox", "zenodo"}},
			{ID: "terms", Kind: "terms-of-use", URL: "https://about.zenodo.org/terms/", RetrievedDate: "2026-07-15", RequiredMarkers: []string{"terms", "CC0"}},
			{ID: "policies", Kind: "repository-policy", URL: "https://about.zenodo.org/policies/", RetrievedDate: "2026-07-15", RequiredMarkers: []string{"policy", "access"}},
		},
		Snapshot: snapshot{
			APIGeneration:     "documented-depositions-rest",
			ProductionBaseURL: "https://zenodo.org/api/",
			SandboxBaseURL:    "https://sandbox.zenodo.org/api/",
			OAIPMHBaseURL:     "https://zenodo.org/oai2d",
			Authentication:    authentication{PreferredTransport: "authorization-bearer-header", DepositionsRequireToken: true, Scopes: []string{"deposit:actions", "deposit:write"}},
			Limits:            limits{GuestRequestsPerMinute: 60, GuestRequestsPerHour: 2000, AuthenticatedRequestsPerMinute: 100, AuthenticatedRequestsPerHour: 5000, SearchRequestsPerMinute: 30, OAIRequestsPerMinute: 30, OAIPageSize: 50, OAIResumptionTokenSeconds: 120, RecordFileLimit: 100, RecordBytesLimit: 50 * 1024 * 1024 * 1024},
			Capabilities: []capability{
				{ID: "deposition-create", Protocol: "rest", Method: "POST", Path: "/api/deposit/depositions", Access: "authenticated", Scope: "deposit:write", Lifecycle: "draft", Risk: "write", Source: "developers"},
				{ID: "deposition-publish", Protocol: "rest", Method: "POST", Path: "/api/deposit/depositions/:id/actions/publish", Access: "authenticated", Scope: "deposit:actions", Lifecycle: "draft", Risk: "irreversible", Source: "developers"},
				{ID: "file-upload-bucket", Protocol: "rest", Method: "PUT", Path: "deposition.links.bucket/:filename", Access: "authenticated", Scope: "deposit:write", Lifecycle: "draft", Risk: "write", Source: "developers"},
				{ID: "oai-harvest", Protocol: "oai-pmh", Method: "GET", Path: "/oai2d", Access: "public", Lifecycle: "published", Risk: "read", Source: "developers"},
				{ID: "record-retrieve", Protocol: "rest", Method: "GET", Path: "/api/records/:id", Access: "public", Lifecycle: "published", Risk: "read", Source: "developers"},
				{ID: "record-search", Protocol: "rest", Method: "GET", Path: "/api/records", Access: "public", Lifecycle: "published", Risk: "read", Source: "developers"},
			},
		},
	}
	return m
}
