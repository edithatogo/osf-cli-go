package zenodoapi

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/edithatogo/osf-cli-go/internal/observability"
	"github.com/edithatogo/osf-cli-go/internal/repository"
)

func TestSearchRecordsPaginatesWithoutAuthentication(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "" {
			t.Fatal("public search sent authorization")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-RateLimit-Limit", "60")
		w.Header().Set("X-RateLimit-Remaining", "59")
		w.Header().Set("X-RateLimit-Reset", "1784060000")
		if request.URL.Query().Get("page") == "2" {
			writeFixture(t, w, "records_page2.json")
			return
		}
		if request.URL.Path != "/api/records/" || request.URL.Query().Get("q") != "open" {
			t.Fatalf("request URL = %s", request.URL.String())
		}
		writeFixture(t, w, "records_page1.json")
	}))
	defer server.Close()
	client := newTestClient(t, server, WithMaxPages(3))
	records, err := client.SearchRecords(t.Context(), "open", 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 3 || records[0].Metadata.Title != "Open methods" || records[2].Files[0].Key != "legacy.txt" {
		t.Fatalf("records = %#v", records)
	}
	if got := client.LastRateLimit(); got.Limit != 60 || got.Remaining != 59 || got.ResetUnix != 1784060000 {
		t.Fatalf("rate limit = %#v", got)
	}
}

func TestGetRecordUsesBearerHeaderAndPreservesNativeMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/records/1001" || request.Header.Get("Authorization") != "Bearer secret-token" {
			t.Fatalf("request = %s auth=%q", request.URL.Path, request.Header.Get("Authorization"))
		}
		writeFixture(t, w, "record_1001.json")
	}))
	defer server.Close()
	client := newTestClient(t, server, WithToken("secret-token"))
	record, err := client.GetRecord(t.Context(), "1001")
	if err != nil {
		t.Fatal(err)
	}
	if record.ID != "1001" || len(record.Files) != 1 || record.Files[0].ContentURL() == "" {
		t.Fatalf("record = %#v", record)
	}
	envelope, err := record.Envelope()
	if err != nil {
		t.Fatal(err)
	}
	if envelope.Identity.Provider != repository.ProviderZenodo || !bytes.Contains(envelope.NativeMetadata.Bytes(), []byte(`"provider_extension"`)) {
		t.Fatalf("envelope = %#v native=%s", envelope, envelope.NativeMetadata.Bytes())
	}
}

func TestListRecordFilesUsesEmbeddedRecordFiles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		writeFixture(t, w, "record_1001.json")
	}))
	defer server.Close()
	client := newTestClient(t, server)
	files, err := client.ListRecordFiles(t.Context(), "1001")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0].Checksum != "md5:9e107d9d372bb6826bd81d3542a419d6" {
		t.Fatalf("files = %#v", files)
	}
}

func TestRetryAndObservabilityAreBoundedAndRedacted(t *testing.T) {
	var attempts atomic.Int32
	var events bytes.Buffer
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if attempts.Add(1) == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = io.WriteString(w, `{"message":"secret-token temporarily limited"}`)
			return
		}
		writeFixture(t, w, "record_1001.json")
	}))
	defer server.Close()
	client := newTestClient(t, server,
		WithToken("secret-token"),
		WithRetryPolicy(2, time.Millisecond),
		WithObserver(observability.NewJSONEmitter(&events, observability.LevelInfo)),
	)
	if _, err := client.GetRecord(t.Context(), "1001"); err != nil {
		t.Fatal(err)
	}
	if attempts.Load() != 2 || !bytes.Contains(events.Bytes(), []byte(`"provider":"zenodo"`)) || !bytes.Contains(events.Bytes(), []byte(`"retryCount":1`)) || bytes.Contains(events.Bytes(), []byte("secret-token")) {
		t.Fatalf("attempts=%d events=%s", attempts.Load(), events.String())
	}
}

func TestAPIErrorIsTypedAndRedacted(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"message":"token secret-token is unauthorized"}`)
	}))
	defer server.Close()
	client := newTestClient(t, server, WithToken("secret-token"), WithRetryPolicy(0, 0))
	_, err := client.GetRecord(t.Context(), "private")
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusUnauthorized || strings.Contains(err.Error(), "secret-token") || apiErr.Path != "/api/records/private" {
		t.Fatalf("error = %#v (%v)", apiErr, err)
	}
}

func TestResponseAndPaginationBudgets(t *testing.T) {
	t.Run("response too large", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
			_, _ = io.WriteString(w, strings.Repeat("x", 65))
		}))
		defer server.Close()
		client := newTestClient(t, server, WithMaxResponseBytes(64), WithRetryPolicy(0, 0))
		_, err := client.GetRecord(t.Context(), "1001")
		if !errors.Is(err, ErrResponseTooLarge) {
			t.Fatalf("error = %v, want ErrResponseTooLarge", err)
		}
	})
	t.Run("cross origin next", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
			_, _ = io.WriteString(w, `{"hits":{"hits":[]},"links":{"next":"https://example.com/api/records?page=2"}}`)
		}))
		defer server.Close()
		client := newTestClient(t, server)
		_, err := client.SearchRecords(t.Context(), "", 0)
		if !errors.Is(err, ErrCrossOriginPagination) {
			t.Fatalf("error = %v, want ErrCrossOriginPagination", err)
		}
	})
	t.Run("page limit", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
			_, _ = io.WriteString(w, `{"hits":{"hits":[]},"links":{"next":"?page=2"}}`)
		}))
		defer server.Close()
		client := newTestClient(t, server, WithMaxPages(1))
		_, err := client.SearchRecords(t.Context(), "", 0)
		if !errors.Is(err, ErrPaginationLimit) {
			t.Fatalf("error = %v, want ErrPaginationLimit", err)
		}
	})
}

func TestMalformedResponseAndCancellation(t *testing.T) {
	t.Run("malformed", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) { _, _ = io.WriteString(w, `{`) }))
		defer server.Close()
		client := newTestClient(t, server)
		if _, err := client.GetRecord(t.Context(), "1001"); err == nil || !strings.Contains(err.Error(), "decode zenodo") {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("canceled", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) { <-request.Context().Done() }))
		defer server.Close()
		client := newTestClient(t, server, WithRetryPolicy(2, time.Millisecond))
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		_, err := client.GetRecord(ctx, "1001")
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("error = %v, want cancellation", err)
		}
	})
	t.Run("missing hits", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) { _, _ = io.WriteString(w, `{"links":{}}`) }))
		defer server.Close()
		client := newTestClient(t, server)
		if _, err := client.SearchRecords(t.Context(), "", 0); err == nil || !strings.Contains(err.Error(), "missing hits") {
			t.Fatalf("error = %v", err)
		}
	})
}

func TestNewRejectsUnsafeBaseURL(t *testing.T) {
	for _, raw := range []string{"ftp://zenodo.org/api/", "https://token@example.com/api/", "https://zenodo.org/api/?access_token=x", "https://example.com/api/"} {
		if _, err := New(raw); err == nil {
			t.Errorf("New(%q) returned nil error", raw)
		}
	}
}

func TestCrossOriginRedirectIsRejectedBeforeAuthorizationForwarding(t *testing.T) {
	var receivedAuthorization atomic.Bool
	destination := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "" {
			receivedAuthorization.Store(true)
		}
	}))
	defer destination.Close()
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		http.Redirect(w, request, destination.URL+"/capture", http.StatusFound)
	}))
	defer source.Close()
	client := newTestClient(t, source, WithToken("secret-token"))
	_, err := client.GetRecord(t.Context(), "1001")
	if !errors.Is(err, ErrCrossOriginPagination) || receivedAuthorization.Load() {
		t.Fatalf("error = %v authorization forwarded = %v", err, receivedAuthorization.Load())
	}
}

func TestConcurrencyBudgetSerializesRequests(t *testing.T) {
	entered := make(chan struct{}, 2)
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		entered <- struct{}{}
		<-release
		writeFixture(t, w, "record_1001.json")
	}))
	defer server.Close()
	client := newTestClient(t, server, WithMaxConcurrency(1))
	errorsCh := make(chan error, 2)
	for range 2 {
		go func() {
			_, err := client.GetRecord(t.Context(), "1001")
			errorsCh <- err
		}()
	}
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("first request did not enter handler")
	}
	select {
	case <-entered:
		t.Fatal("second request bypassed concurrency budget")
	case <-time.After(20 * time.Millisecond):
	}
	release <- struct{}{}
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("second request did not enter after release")
	}
	release <- struct{}{}
	for range 2 {
		if err := <-errorsCh; err != nil {
			t.Fatal(err)
		}
	}
}

func FuzzResolvePagination(f *testing.F) {
	for _, seed := range []string{"?page=2", "/api/records?page=2", "https://example.com/", "\x00"} {
		f.Add(seed)
	}
	client, err := New("https://zenodo.org/api/")
	if err != nil {
		f.Fatal(err)
	}
	f.Fuzz(func(t *testing.T, next string) {
		resolved, err := client.resolvePagination(next)
		if err == nil && (resolved == nil || resolved.Scheme != "https" || resolved.Host != "zenodo.org") {
			t.Fatalf("resolved = %v", resolved)
		}
	})
}

func FuzzDecodeSearchPage(f *testing.F) {
	for _, seed := range []string{`{"hits":{"hits":[]},"links":{}}`, `[]`, `{`, `null`} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, value string) {
		_, _, _ = decodeSearchPage([]byte(value))
	})
}

func FuzzDecodeFiles(f *testing.F) {
	for _, seed := range []string{`[]`, `{"entries":{},"order":[]}`, `{`, `null`} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, value string) {
		_, _ = decodeFiles([]byte(value))
	})
}

func newTestClient(t *testing.T, server *httptest.Server, options ...Option) *Client {
	t.Helper()
	options = append([]Option{WithHTTPClient(server.Client())}, options...)
	client, err := New(server.URL+"/api/", options...)
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func writeFixture(t *testing.T, writer io.Writer, name string) {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Write(data); err != nil {
		t.Fatal(err)
	}
}
