package zenodoapi

import (
	"bytes"
	"context"
	"encoding/json"
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
	native := record.NativeJSON()
	if !bytes.Contains(native, []byte(`"provider_extension"`)) {
		t.Fatalf("native JSON = %s", native)
	}
	native[0] = 'x'
	if bytes.Equal(native, record.NativeJSON()) {
		t.Fatal("NativeJSON returned shared storage")
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
	t.Run("numeric record identifiers", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
			_, _ = io.WriteString(w, `{"hits":{"hits":[{"id":1234,"conceptrecid":5678,"metadata":{"title":"Numeric ID"},"files":[],"links":{"self":"https://example.test/1234","thumbnails":{"10":"https://example.test/thumb"}}}]},"links":{}}`)
		}))
		defer server.Close()
		client := newTestClient(t, server)
		records, err := client.SearchRecords(t.Context(), "numeric", 0)
		if err != nil {
			t.Fatalf("SearchRecords() error = %v", err)
		}
		if len(records) != 1 || records[0].ID != "1234" || records[0].ConceptRecID != "5678" || records[0].Links["self"] != "https://example.test/1234" {
			t.Fatalf("records = %#v", records)
		}
	})

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

func TestNewDefaultsAndRejectsInvalidBudgets(t *testing.T) {
	client, err := New("")
	if err != nil || client.baseURL.String() != defaultBaseURL || client.httpClient == nil {
		t.Fatalf("New() = %#v, %v", client, err)
	}
	for _, option := range []Option{WithMaxResponseBytes(0), WithMaxPages(0), WithMaxConcurrency(0), WithRetryPolicy(-1, 0), WithRetryPolicy(0, -1)} {
		if _, err := New("", option); err == nil {
			t.Fatal("New() accepted invalid budget")
		}
	}
	if _, err := New("", WithHTTPClient(nil)); err != nil {
		t.Fatalf("New() with nil HTTP client error = %v", err)
	}
}

func TestSearchRecordLimitsCyclesAndLegacyShape(t *testing.T) {
	t.Run("negative limit", func(t *testing.T) {
		client, err := New("")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := client.SearchRecords(t.Context(), "", -1); err == nil {
			t.Fatal("SearchRecords() accepted negative limit")
		}
	})
	t.Run("early limit", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) { writeFixture(t, w, "records_page1.json") }))
		defer server.Close()
		client := newTestClient(t, server)
		records, err := client.SearchRecords(t.Context(), "open", 1)
		if err != nil || len(records) != 1 {
			t.Fatalf("records=%d error=%v", len(records), err)
		}
	})
	t.Run("cycle", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
			_, _ = io.WriteString(w, `{"hits":{"hits":[]},"links":{"next":"?size=100"}}`)
		}))
		defer server.Close()
		client := newTestClient(t, server)
		_, err := client.SearchRecords(t.Context(), "", 0)
		if err == nil || !strings.Contains(err.Error(), "cycle") {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("legacy array", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) { _, _ = io.WriteString(w, `[]`) }))
		defer server.Close()
		client := newTestClient(t, server)
		records, err := client.SearchRecords(t.Context(), "", 0)
		if err != nil || len(records) != 0 {
			t.Fatalf("records=%v error=%v", records, err)
		}
	})
}

func TestRecordValidationAndFileErrorPropagation(t *testing.T) {
	client, err := New("")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetRecord(t.Context(), " "); err == nil {
		t.Fatal("GetRecord() accepted empty id")
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		_, _ = io.WriteString(w, `{"metadata":{"title":"missing id"}}`)
	}))
	defer server.Close()
	client = newTestClient(t, server)
	if _, err := client.GetRecord(t.Context(), "missing"); err == nil || !strings.Contains(err.Error(), "id is missing") {
		t.Fatalf("GetRecord() error = %v", err)
	}
	if _, err := client.ListRecordFiles(t.Context(), "missing"); err == nil {
		t.Fatal("ListRecordFiles() did not propagate record error")
	}
}

func TestFileShapesContentURLsAndEnvelopeStates(t *testing.T) {
	files, err := decodeFiles([]byte(`{"order":["b"],"entries":{"a":{"size":1},"b":{"key":"named","links":{"download":"https://example/download"}}}}`))
	if err != nil || len(files) != 2 || files[0].Key != "named" || files[1].Key != "a" || files[0].ContentURL() == "" {
		t.Fatalf("files=%#v error=%v", files, err)
	}
	if got := (File{Links: map[string]string{"self": "https://example/self"}}).ContentURL(); got != "https://example/self" {
		t.Fatalf("ContentURL() = %q", got)
	}
	if got := (File{}).ContentURL(); got != "" {
		t.Fatalf("empty ContentURL() = %q", got)
	}
	for _, value := range []string{"null", `{}`, `{"entries":`} {
		if _, err := decodeFiles([]byte(value)); value == `{"entries":` && err == nil {
			t.Fatal("decodeFiles() accepted malformed object")
		}
	}
	if _, err := (Record{}).Envelope(); err == nil {
		t.Fatal("Envelope() accepted empty record")
	}
	var record Record
	if err := json.Unmarshal(mustReadFixture(t, "record_1001.json"), &record); err != nil {
		t.Fatal(err)
	}
	record.Metadata.AccessRight = "embargoed"
	record.Files = append(record.Files, File{Checksum: "unknown"})
	envelope, err := record.Envelope()
	if err != nil || envelope.Lifecycle.Common != repository.LifecycleEmbargoed {
		t.Fatalf("Envelope() = %#v, %v", envelope, err)
	}
}

func TestErrorAndTimingHelpers(t *testing.T) {
	var nilError *APIError
	if nilError.Error() != "<nil>" {
		t.Fatalf("nil APIError = %q", nilError.Error())
	}
	if retryDelay("999", time.Millisecond) != 5*time.Second || retryDelay("invalid", time.Millisecond) != time.Millisecond {
		t.Fatal("retryDelay() did not enforce fallback/cap")
	}
	if got := retryDelay(time.Now().Add(-time.Second).UTC().Format(http.TimeFormat), time.Second); got != 0 {
		t.Fatalf("past Retry-After = %v", got)
	}
	if got := retryDelay(time.Now().Add(time.Hour).UTC().Format(http.TimeFormat), time.Second); got <= 0 || got > 5*time.Second {
		t.Fatalf("future Retry-After = %v", got)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if !errors.Is(sleepContext(ctx, time.Second), context.Canceled) || !errors.Is(sleepContext(ctx, 0), context.Canceled) {
		t.Fatal("sleepContext() did not propagate cancellation")
	}
	if _, err := readBounded(errorReader{}, 10); err == nil {
		t.Fatal("readBounded() ignored reader error")
	}
}

func TestNetworkFailureRetriesAndFallbackErrorBody(t *testing.T) {
	var attempts atomic.Int32
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if attempts.Add(1) == 1 {
			return nil, errors.New("temporary transport failure")
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewReader(mustReadFixture(t, "record_1001.json"))),
			Request:    request,
		}, nil
	})
	client, err := New("", WithHTTPClient(&http.Client{Transport: transport}), WithRetryPolicy(1, time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetRecord(t.Context(), "1001"); err != nil || attempts.Load() != 2 {
		t.Fatalf("attempts=%d error=%v", attempts.Load(), err)
	}

	request := httptest.NewRequest(http.MethodGet, "https://zenodo.org/api/records/x", nil)
	apiErr := parseAPIError(&http.Response{StatusCode: http.StatusBadRequest, Header: make(http.Header), Request: request}, []byte("plain failure"), "")
	if apiErr.Message != "plain failure" || !strings.Contains(apiErr.Error(), "plain failure") {
		t.Fatalf("APIError = %#v", apiErr)
	}
}

func TestDecodeHelpersRejectEmptyAndMalformedRecordFiles(t *testing.T) {
	if _, _, err := decodeSearchPage(nil); err == nil {
		t.Fatal("decodeSearchPage() accepted empty response")
	}
	var record Record
	if err := json.Unmarshal([]byte(`{"id":"1","metadata":{},"files":{"entries":`), &record); err == nil {
		t.Fatal("Record.UnmarshalJSON() accepted malformed files")
	}
	client, err := New("")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.resolve("%zz"); err == nil {
		t.Fatal("resolve() accepted invalid URL escape")
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
	data := mustReadFixture(t, name)
	if _, err := writer.Write(data); err != nil {
		t.Fatal(err)
	}
}

func mustReadFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) { return 0, errors.New("read failed") }

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}
