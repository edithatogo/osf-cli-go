package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func TestRunLoadsStrictManifest(t *testing.T) {
	m := validManifest()
	m.SnapshotSHA256 = snapshotDigest(m.Snapshot)
	path := filepath.Join(t.TempDir(), "manifest.json")
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := run(path, false, nil)
	if err != nil || got.ReviewedDate != m.ReviewedDate {
		t.Fatalf("run() = %#v, %v", got, err)
	}
	if _, err := run(filepath.Join(t.TempDir(), "missing.json"), false, nil); err == nil {
		t.Fatal("run() accepted missing manifest")
	}
	if _, err := run(path, true, nil); err == nil || !strings.Contains(err.Error(), "HTTP client") {
		t.Fatalf("run() online error = %v", err)
	}
}

func TestExecuteValidatesAndPrintsDigest(t *testing.T) {
	m := validManifest()
	m.SnapshotSHA256 = snapshotDigest(m.Snapshot)
	path := filepath.Join(t.TempDir(), "manifest.json")
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	var output strings.Builder
	if err := execute([]string{"-manifest", path}, &output, nil); err != nil || !strings.Contains(output.String(), "valid") {
		t.Fatalf("execute() output=%q error=%v", output.String(), err)
	}
	output.Reset()
	if err := execute([]string{"-manifest", path, "-print-digest"}, &output, nil); err != nil || strings.TrimSpace(output.String()) != m.SnapshotSHA256 {
		t.Fatalf("execute() digest=%q error=%v", output.String(), err)
	}
	if err := execute([]string{"-unknown"}, io.Discard, nil); err == nil {
		t.Fatal("execute() accepted unknown flag")
	}
	if err := execute([]string{"-manifest", "missing"}, io.Discard, nil); err == nil {
		t.Fatal("execute() accepted missing manifest")
	}
}

func TestLoadManifestRejectsUnknownAndTrailingJSON(t *testing.T) {
	for name, data := range map[string]string{
		"unknown":  `{"unknown":true}`,
		"trailing": `{}` + `{}`,
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "manifest.json")
			if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := loadManifest(path); err == nil {
				t.Fatal("loadManifest() returned nil error")
			}
		})
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

func TestValidateManifestRejectsHeaderAndPolicyDrift(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*manifest)
	}{
		{name: "schema", mutate: func(m *manifest) { m.SchemaVersion = 2 }},
		{name: "date", mutate: func(m *manifest) { m.ReviewedDate = "today" }},
		{name: "policy", mutate: func(m *manifest) { m.VersionPolicy = "latest" }},
		{name: "decision", mutate: func(m *manifest) { m.VersionDecision = "" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			m := validManifest()
			test.mutate(&m)
			m.SnapshotSHA256 = snapshotDigest(m.Snapshot)
			if err := validateManifest(m); err == nil {
				t.Fatal("validateManifest() returned nil error")
			}
		})
	}
}

func TestValidateSourcesRejectsStructuralDrift(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*manifest)
	}{
		{name: "count", mutate: func(m *manifest) { m.Sources = m.Sources[:3] }},
		{name: "order", mutate: func(m *manifest) { m.Sources[0].ID = "sandbox" }},
		{name: "kind", mutate: func(m *manifest) { m.Sources[0].Kind = "" }},
		{name: "markers", mutate: func(m *manifest) { m.Sources[0].RequiredMarkers = []string{"one"} }},
		{name: "date", mutate: func(m *manifest) { m.Sources[0].RetrievedDate = "2020-01-01" }},
		{name: "query", mutate: func(m *manifest) { m.Sources[0].URL += "?x=1" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			m := validManifest()
			test.mutate(&m)
			if err := validateSources(m); err == nil {
				t.Fatal("validateSources() returned nil error")
			}
		})
	}
}

func TestValidateSnapshotRejectsContractDrift(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*snapshot)
	}{
		{name: "generation", mutate: func(s *snapshot) { s.APIGeneration = "other" }},
		{name: "base", mutate: func(s *snapshot) { s.ProductionBaseURL = "https://example.com/" }},
		{name: "transport", mutate: func(s *snapshot) { s.Authentication.PreferredTransport = "query" }},
		{name: "public token", mutate: func(s *snapshot) { s.Authentication.PublicRecordsRequireToken = true }},
		{name: "scopes", mutate: func(s *snapshot) { s.Authentication.Scopes = []string{"deposit:write"} }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := validManifest().Snapshot
			test.mutate(&s)
			if err := validateSnapshot(s); err == nil {
				t.Fatal("validateSnapshot() returned nil error")
			}
		})
	}
}

func TestValidateCapabilitiesRejectsMalformedDecisions(t *testing.T) {
	tests := []struct {
		name   string
		mutate func([]capability) []capability
	}{
		{name: "missing", mutate: func(c []capability) []capability { return c[:len(c)-1] }},
		{name: "protocol", mutate: func(c []capability) []capability { c[0].Protocol = "ftp"; return c }},
		{name: "risk", mutate: func(c []capability) []capability { c[0].Risk = "maybe"; return c }},
		{name: "source", mutate: func(c []capability) []capability { c[0].Source = "other"; return c }},
		{name: "empty semantics", mutate: func(c []capability) []capability { c[0].Path = ""; return c }},
		{name: "public write", mutate: func(c []capability) []capability { c[0].Access = "public"; return c }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			items := append([]capability(nil), validManifest().Snapshot.Capabilities...)
			if err := validateCapabilities(test.mutate(items)); err == nil {
				t.Fatal("validateCapabilities() returned nil error")
			}
		})
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

func TestValidateOnlineSuccessAndFailures(t *testing.T) {
	body := "records depositions sandbox zenodo terms CC0 policy access"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/status" {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = io.WriteString(w, body)
	}))
	defer server.Close()
	m := validManifest()
	for i := range m.Sources {
		m.Sources[i].URL = server.URL + "/ok"
	}
	if err := validateOnline(m, server.Client()); err != nil {
		t.Fatalf("validateOnline() error = %v", err)
	}
	m.Sources[0].URL = server.URL + "/status"
	if err := validateOnline(m, server.Client()); err == nil || !strings.Contains(err.Error(), "HTTP 503") {
		t.Fatalf("status error = %v", err)
	}
	m.Sources[0].URL = server.URL + "/ok"
	m.Sources[0].RequiredMarkers = []string{"missing", "records"}
	if err := validateOnline(m, server.Client()); err == nil || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("marker error = %v", err)
	}
}

func TestValidateOnlineReadAndTransportFailures(t *testing.T) {
	m := validManifest()
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return nil, errors.New("network down")
	})}
	if err := validateOnline(m, client); err == nil || !strings.Contains(err.Error(), "network down") {
		t.Fatalf("transport error = %v", err)
	}
	client.Transport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(errorReader{}), Request: request}, nil
	})
	if err := validateOnline(m, client); err == nil || !strings.Contains(err.Error(), "read failed") {
		t.Fatalf("read error = %v", err)
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

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) { return 0, errors.New("read failed") }
