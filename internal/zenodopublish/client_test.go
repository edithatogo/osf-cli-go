package zenodopublish

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestClientRejectsProductionAndMissingCredentials(t *testing.T) {
	t.Parallel()
	for name, input := range map[string]struct {
		baseURL string
		token   string
	}{
		"missing token": {baseURL: "https://sandbox.zenodo.org/api/"},
		"production":    {baseURL: "https://zenodo.org/api/", token: "secret"},
		"other host":    {baseURL: "https://example.com/api/", token: "secret"},
		"url token":     {baseURL: "https://secret@sandbox.zenodo.org/api/", token: "secret"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := New(input.baseURL, input.token, []Scope{ScopeDepositWrite}); err == nil {
				t.Fatal("New returned nil error")
			}
		})
	}
}

func TestExecuteRejectsInvalidAndDryRunBeforeNetwork(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { calls.Add(1) }))
	defer server.Close()
	client, err := New(server.URL+"/api/", "secret", []Scope{ScopeDepositWrite, ScopeDepositActions})
	if err != nil {
		t.Fatal(err)
	}
	request := Request{RecordID: "123", State: StatePublished, Action: ActionPublish, Authorized: true, Metadata: validMetadata()}
	if _, err := client.Execute(t.Context(), request, time.Now()); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("invalid transition error = %v", err)
	}
	request.State = StateDraft
	request.DryRun = true
	result, err := client.Execute(t.Context(), request, time.Now())
	if err != nil || result.Executed || result.Plan.Confirmation == "" {
		t.Fatalf("dry-run result = %+v err=%v", result, err)
	}
	if calls.Load() != 0 {
		t.Fatalf("network calls = %d, want 0", calls.Load())
	}
}

func TestExecuteUsesClientScopesAndEmitsAuditOutcomes(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	var events []AuditEvent
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		http.Error(w, "denied", http.StatusForbidden)
	}))
	defer server.Close()
	client, err := New(server.URL+"/api/", "secret", []Scope{ScopeDepositActions}, WithAuditSink(func(event AuditEvent) {
		events = append(events, event)
	}))
	if err != nil {
		t.Fatal(err)
	}
	request := Request{
		RecordID: "123", State: StateDraft, Action: ActionPublish, Authorized: true,
		DryRun: true, Scopes: []Scope{ScopeDepositWrite, ScopeDepositActions}, Metadata: validMetadata(),
	}
	if _, err := client.Execute(t.Context(), request, time.Now()); !errors.Is(err, ErrScopeRequired) {
		t.Fatalf("scope error = %v", err)
	}
	if calls.Load() != 0 || len(events) != 0 {
		t.Fatalf("calls=%d events=%+v", calls.Load(), events)
	}

	client, err = New(server.URL+"/api/", "secret", []Scope{ScopeDepositWrite, ScopeDepositActions}, WithAuditSink(func(event AuditEvent) {
		events = append(events, event)
	}))
	if err != nil {
		t.Fatal(err)
	}
	preview, err := client.Execute(t.Context(), request, time.Now())
	if err != nil || len(events) != 1 || events[0].Outcome != "planned" {
		t.Fatalf("preview = %+v events=%+v err=%v", preview, events, err)
	}
	request.DryRun = false
	request.Confirmation = preview.Plan.Confirmation
	if _, err := client.Execute(t.Context(), request, time.Now()); err == nil {
		t.Fatal("failed publish returned nil error")
	}
	if len(events) != 2 || events[1].Outcome != "failed" || events[1].Error == "" {
		t.Fatalf("events = %+v", events)
	}
}

func TestExecutePublishAppliesMetadataThenPublishes(t *testing.T) {
	t.Parallel()
	const token = "sandbox-secret-token"
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+token {
			t.Errorf("authorization = %q", r.Header.Get("Authorization"))
		}
		switch call := calls.Add(1); call {
		case 1:
			if r.Method != http.MethodPut || r.URL.Path != "/api/deposit/depositions/123" {
				t.Errorf("metadata request = %s %s", r.Method, r.URL.Path)
			}
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatal(err)
			}
			metadata, _ := payload["metadata"].(map[string]any)
			keywords, _ := metadata["keywords"].([]any)
			if metadata["title"] != "Reproducible example" || metadata["access_right"] != "open" || metadata["license"] != "cc-by-4.0" || len(keywords) != 1 || keywords[0] != "reproducible" {
				t.Errorf("metadata payload = %#v", metadata)
			}
			_, _ = io.WriteString(w, `{"id":123,"state":"unsubmitted"}`)
		case 2:
			if r.Method != http.MethodPost || r.URL.Path != "/api/deposit/depositions/123/actions/publish" {
				t.Errorf("publish request = %s %s", r.Method, r.URL.Path)
			}
			_, _ = io.WriteString(w, `{"id":123,"doi":"10.5072/zenodo.123","conceptdoi":"10.5072/zenodo.100","state":"done","submitted":true}`)
		default:
			t.Errorf("unexpected call %d", call)
		}
	}))
	defer server.Close()
	client, err := New(server.URL+"/api/", token, []Scope{ScopeDepositWrite, ScopeDepositActions})
	if err != nil {
		t.Fatal(err)
	}
	dry := Request{RecordID: "123", State: StateDraft, Action: ActionPublish, Authorized: true, DryRun: true, Metadata: validMetadata()}
	preview, err := client.Execute(t.Context(), dry, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	dry.DryRun = false
	dry.Confirmation = preview.Plan.Confirmation
	result, err := client.Execute(t.Context(), dry, time.Now())
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !result.Executed || result.Plan.To != StatePublished || result.DOI != "10.5072/zenodo.123" || result.ConceptDOI != "10.5072/zenodo.100" || calls.Load() != 2 {
		t.Fatalf("result = %+v calls=%d", result, calls.Load())
	}
}

func TestPublishFailureReportsAppliedMetadataBoundary(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			_, _ = io.WriteString(w, `{"id":123}`)
			return
		}
		http.Error(w, "publication unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()
	client, err := New(server.URL+"/api/", "secret", []Scope{ScopeDepositWrite, ScopeDepositActions})
	if err != nil {
		t.Fatal(err)
	}
	request := Request{RecordID: "123", State: StateDraft, Action: ActionPublish, Authorized: true, DryRun: true, Metadata: validMetadata()}
	preview, err := client.Execute(t.Context(), request, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	request.DryRun = false
	request.Confirmation = preview.Plan.Confirmation
	_, err = client.Execute(t.Context(), request, time.Now())
	var partial *PartialPublicationError
	if !errors.As(err, &partial) || partial.RecordID != "123" || !strings.Contains(err.Error(), "inspect the draft before retrying") || calls.Load() != 2 {
		t.Fatalf("partial error = %v calls=%d", err, calls.Load())
	}
}

func TestExecuteReserveDOINewVersionAndDiscard(t *testing.T) {
	t.Parallel()
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/deposit/depositions/123":
			_, _ = io.WriteString(w, `{"id":123,"metadata":{"prereserve_doi":{"doi":"10.5072/zenodo.123","recid":123}}}`)
		case r.Method == http.MethodPost && r.URL.Path == "/api/deposit/depositions/123/actions/newversion":
			_, _ = fmt.Fprintf(w, `{"id":123,"links":{"latest_draft":%q}}`, server.URL+"/api/deposit/depositions/124")
		case r.Method == http.MethodGet && r.URL.Path == "/api/deposit/depositions/124":
			_, _ = io.WriteString(w, `{"id":124,"state":"unsubmitted"}`)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/deposit/depositions/124":
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "unexpected", http.StatusNotFound)
		}
	}))
	defer server.Close()
	client, err := New(server.URL+"/api/", "secret", []Scope{ScopeDepositWrite, ScopeDepositActions})
	if err != nil {
		t.Fatal(err)
	}
	reserved, err := client.Execute(t.Context(), Request{RecordID: "123", State: StateDraft, Action: ActionReserveDOI, Authorized: true}, time.Now())
	if err != nil || reserved.DOI != "10.5072/zenodo.123" || reserved.Plan.To != StateDOIReserved {
		t.Fatalf("reserve = %+v err=%v", reserved, err)
	}
	version, err := client.Execute(t.Context(), Request{RecordID: "123", State: StatePublished, Action: ActionNewVersion, Authorized: true}, time.Now())
	if err != nil || version.RecordID != "124" || version.Plan.To != StateVersionDraft {
		t.Fatalf("version = %+v err=%v", version, err)
	}
	discard := Request{RecordID: "124", State: StateVersionDraft, Action: ActionDiscard, Authorized: true, DryRun: true}
	preview, err := client.Execute(t.Context(), discard, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	discard.DryRun = false
	discard.Confirmation = preview.Plan.Confirmation
	result, err := client.Execute(t.Context(), discard, time.Now())
	if err != nil || !result.Executed || result.Plan.To != StatePublished {
		t.Fatalf("discard = %+v err=%v", result, err)
	}
}

func TestExecuteRedactsErrorsAndRejectsCrossOriginVersionLink(t *testing.T) {
	t.Parallel()
	const token = "top-secret-token"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/actions/newversion") {
			_, _ = io.WriteString(w, `{"id":123,"links":{"latest_draft":"https://example.com/api/deposit/depositions/124"}}`)
			return
		}
		http.Error(w, "Bearer "+token, http.StatusBadRequest)
	}))
	defer server.Close()
	client, err := New(server.URL+"/api/", token, []Scope{ScopeDepositWrite, ScopeDepositActions}, WithMaxResponseBytes(1024))
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Execute(t.Context(), Request{RecordID: "123", State: StatePublished, Action: ActionNewVersion, Authorized: true}, time.Now())
	if !errors.Is(err, ErrCrossOrigin) {
		t.Fatalf("cross-origin error = %v", err)
	}
	request := Request{RecordID: "123", State: StateDraft, Action: ActionReserveDOI, Authorized: true}
	_, err = client.Execute(context.Background(), request, time.Now())
	if err == nil || strings.Contains(err.Error(), token) || !strings.Contains(err.Error(), "REDACTED") {
		t.Fatalf("redacted error = %v", err)
	}
	for cause := err; cause != nil; cause = errors.Unwrap(cause) {
		if strings.Contains(cause.Error(), token) {
			t.Fatalf("unwrapped error leaked token: %v", cause)
		}
	}
}

func TestExecuteBoundsResponses(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, strings.Repeat("x", 33))
	}))
	defer server.Close()
	client, err := New(server.URL+"/api/", "secret", []Scope{ScopeDepositWrite}, WithMaxResponseBytes(32))
	if err != nil {
		t.Fatal(err)
	}
	request := Request{RecordID: "123", State: StateDraft, Action: ActionReserveDOI, Authorized: true}
	if _, err := client.Execute(t.Context(), request, time.Now()); !errors.Is(err, ErrResponseTooLarge) {
		t.Fatalf("response error = %v", err)
	}
}

func TestApplyDraftMetadataRequiresWriteScopeAndNeverPublishes(t *testing.T) {
	t.Parallel()
	var methods []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method+" "+r.URL.Path)
		_, _ = io.WriteString(w, `{"id":123}`)
	}))
	defer server.Close()
	client, err := New(server.URL+"/api/", "secret", []Scope{ScopeDepositActions})
	if err != nil {
		t.Fatal(err)
	}
	if err := client.ApplyDraftMetadata(t.Context(), "123", validMetadata(), time.Now()); !errors.Is(err, ErrScopeRequired) {
		t.Fatalf("scope error = %v", err)
	}
	client, err = New(server.URL+"/api/", "secret", []Scope{ScopeDepositWrite})
	if err != nil {
		t.Fatal(err)
	}
	if err := client.ApplyDraftMetadata(t.Context(), "123", validMetadata(), time.Now()); err != nil {
		t.Fatal(err)
	}
	if len(methods) != 1 || methods[0] != "PUT /api/deposit/depositions/123" {
		t.Fatalf("methods = %+v", methods)
	}
}
