package zenodooai

import (
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
)

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func newFixtureClient(t *testing.T, handler http.HandlerFunc, options ...Option) (*Client, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)
	options = append([]Option{WithClock(func() time.Time { return time.Date(2026, 7, 15, 1, 3, 30, 0, time.UTC) })}, options...)
	client, err := New(server.URL+"/oai2d", options...)
	if err != nil {
		server.Close()
		t.Fatal(err)
	}
	return client, server
}

func TestListRecordsPreservesMetadataAndContinuation(t *testing.T) {
	client, server := newFixtureClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("verb") != "ListRecords" || r.URL.Query().Get("metadataPrefix") != "oai_dc" || r.URL.Query().Get("set") != "user-demo" {
			t.Errorf("unexpected query: %s", r.URL.RawQuery)
		}
		_, _ = w.Write(fixture(t, "records_page1.xml"))
	})
	defer server.Close()
	page, err := client.ListRecords(context.Background(), Request{MetadataPrefix: "oai_dc", Set: "user-demo"})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Records) != 1 || page.Records[0].Header.Identifier != "oai:zenodo.org:1001" {
		t.Fatalf("unexpected records: %#v", page.Records)
	}
	if page.Records[0].NativeMetadata == nil || !strings.Contains(string(page.Records[0].NativeMetadata.Bytes()), "Preserved title") {
		t.Fatalf("native metadata not preserved: %#v", page.Records[0])
	}
	if page.Records[0].Provenance.MetadataPrefix != "oai_dc" || page.Records[0].Provenance.Set != "user-demo" {
		t.Fatalf("missing provenance: %#v", page.Records[0].Provenance)
	}
	if page.Next.Value != "opaque-token-1" || page.Next.Cursor != 0 || page.Next.CompleteListSize != 2 || page.Next.MetadataPrefix != "oai_dc" {
		t.Fatalf("unexpected token: %#v", page.Next)
	}
	encoded, err := json.Marshal(page.Records[0])
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(encoded), "application/xml") {
		t.Fatalf("metadata envelope missing: %s", encoded)
	}
}

func TestHarvestFollowsOpaqueTokenAndPreservesDeletedRecord(t *testing.T) {
	var calls atomic.Int32
	client, server := newFixtureClient(t, func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			_, _ = w.Write(fixture(t, "records_page1.xml"))
			return
		}
		if got := r.URL.Query().Get("resumptionToken"); got != "opaque-token-1" {
			t.Errorf("token = %q", got)
		}
		if r.URL.Query().Get("metadataPrefix") != "" {
			t.Error("selectors repeated with token")
		}
		_, _ = w.Write(fixture(t, "records_page2.xml"))
	})
	defer server.Close()
	records, err := client.Harvest(context.Background(), Request{MetadataPrefix: "oai_dc", Set: "user-demo"})
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 || !records[1].Header.Deleted || records[1].NativeMetadata != nil || len(records[1].AboutXML) == 0 {
		t.Fatalf("unexpected records: %#v", records)
	}
	if records[1].Provenance.MetadataPrefix != "oai_dc" || records[1].Provenance.Set != "user-demo" {
		t.Fatalf("continuation provenance lost: %#v", records[1].Provenance)
	}
}

func TestListSetsAndFormats(t *testing.T) {
	var setCalls atomic.Int32
	client, server := newFixtureClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("verb") {
		case "ListSets":
			if setCalls.Add(1) == 1 {
				_, _ = w.Write(fixture(t, "sets_page1.xml"))
			} else {
				if r.URL.Query().Get("resumptionToken") != "sets-token" {
					t.Error("missing set token")
				}
				_, _ = w.Write(fixture(t, "sets_page2.xml"))
			}
		case "ListMetadataFormats":
			if r.URL.Query().Get("identifier") != "oai:zenodo.org:1001" {
				t.Error("missing identifier")
			}
			_, _ = w.Write(fixture(t, "formats.xml"))
		default:
			t.Fatalf("unexpected verb %q", r.URL.Query().Get("verb"))
		}
	})
	defer server.Close()
	sets, err := client.ListSets(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(sets) != 2 || sets[0].Spec != "user-demo" || !strings.Contains(string(sets[0].Description), "Synthetic set") {
		t.Fatalf("sets = %#v", sets)
	}
	formats, err := client.ListMetadataFormats(context.Background(), "oai:zenodo.org:1001")
	if err != nil {
		t.Fatal(err)
	}
	if len(formats) != 2 || formats[1].Prefix != "datacite" {
		t.Fatalf("formats = %#v", formats)
	}
}

func TestProtocolMalformedAndHTTPFailures(t *testing.T) {
	tests := []struct {
		name, fixture string
		status        int
		want          string
	}{
		{name: "protocol", fixture: "bad_resumption_token.xml", status: 200, want: "badResumptionToken"},
		{name: "malformed", fixture: "malformed.xml", status: 200, want: "decode Zenodo"},
		{name: "http", status: 403, want: "HTTP 403"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client, server := newFixtureClient(t, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(test.status)
				if test.fixture != "" {
					_, _ = w.Write(fixture(t, test.fixture))
				}
			}, WithRetryPolicy(0, 0))
			defer server.Close()
			_, err := client.ListRecords(context.Background(), Request{MetadataPrefix: "oai_dc"})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v", err)
			}
			if test.name == "protocol" {
				var protocolErr *ProtocolError
				if !errors.As(err, &protocolErr) || protocolErr.Code != "badResumptionToken" {
					t.Fatalf("not typed: %T", err)
				}
			}
		})
	}
}

func TestRetryCancellationAndLimits(t *testing.T) {
	t.Run("retry", func(t *testing.T) {
		var calls atomic.Int32
		client, server := newFixtureClient(t, func(w http.ResponseWriter, _ *http.Request) {
			if calls.Add(1) == 1 {
				w.Header().Set("Retry-After", "0")
				w.WriteHeader(429)
				return
			}
			_, _ = w.Write(fixture(t, "records_page1.xml"))
		}, WithRetryPolicy(1, 0))
		defer server.Close()
		if _, err := client.ListRecords(context.Background(), Request{MetadataPrefix: "oai_dc"}); err != nil {
			t.Fatal(err)
		}
		if calls.Load() != 2 {
			t.Fatalf("calls = %d", calls.Load())
		}
	})
	t.Run("cancel", func(t *testing.T) {
		client, server := newFixtureClient(t, func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(503) }, WithRetryPolicy(2, time.Hour))
		defer server.Close()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := client.ListRecords(ctx, Request{MetadataPrefix: "oai_dc"})
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("response bytes", func(t *testing.T) {
		client, server := newFixtureClient(t, func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(fixture(t, "records_page1.xml")) }, WithLimits(20, 2, 2))
		defer server.Close()
		_, err := client.ListRecords(context.Background(), Request{MetadataPrefix: "oai_dc"})
		if !errors.Is(err, ErrResponseTooLarge) {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("records", func(t *testing.T) {
		var calls atomic.Int32
		client, server := newFixtureClient(t, func(w http.ResponseWriter, _ *http.Request) {
			if calls.Add(1) == 1 {
				_, _ = w.Write(fixture(t, "records_page1.xml"))
			} else {
				_, _ = w.Write(fixture(t, "records_page2.xml"))
			}
		}, WithLimits(1<<20, 2, 1))
		defer server.Close()
		_, err := client.Harvest(context.Background(), Request{MetadataPrefix: "oai_dc"})
		if !errors.Is(err, ErrRecordLimit) {
			t.Fatalf("error = %v", err)
		}
	})
}

func TestValidationAndExpiry(t *testing.T) {
	client, server := newFixtureClient(t, func(http.ResponseWriter, *http.Request) { t.Fatal("request should not be made") })
	defer server.Close()
	tests := []Request{
		{},
		{MetadataPrefix: "oai_dc", Token: ResumptionToken{Value: "token"}},
		{MetadataPrefix: "oai_dc", From: time.Date(2026, 7, 16, 0, 0, 0, 0, time.UTC), Until: time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)},
		{Token: ResumptionToken{Value: "expired", ExpiresAt: time.Date(2026, 7, 15, 1, 3, 0, 0, time.UTC)}},
	}
	for _, request := range tests {
		if _, err := client.ListRecords(context.Background(), request); err == nil {
			t.Fatalf("expected error for %#v", request)
		}
	}
}

func TestNewAndRedirectBoundaries(t *testing.T) {
	for _, endpoint := range []string{"http://example.com/oai", "https://evil.example/oai", "https://user@example.com/oai", "https://zenodo.org/oai?q=x"} {
		if _, err := New(endpoint); err == nil {
			t.Fatalf("accepted %q", endpoint)
		}
	}
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, "unexpected") }))
	defer target.Close()
	client, server := newFixtureClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Location", target.URL)
		w.WriteHeader(http.StatusFound)
	})
	defer server.Close()
	_, err := client.ListRecords(context.Background(), Request{MetadataPrefix: "oai_dc"})
	if !errors.Is(err, ErrCrossOrigin) {
		t.Fatalf("error = %v", err)
	}
}

func TestOptionsAndConstructorValidation(t *testing.T) {
	var events strings.Builder
	httpClient := &http.Client{Timeout: time.Second}
	client, err := New("", WithHTTPClient(httpClient), WithObserver(observability.NewJSONEmitter(&events, observability.LevelInfo)))
	if err != nil || client.httpClient == httpClient {
		t.Fatalf("client=%#v err=%v", client, err)
	}
	for _, options := range [][]Option{
		{WithHTTPClient(nil), WithLimits(0, 1, 1)},
		{WithLimits(1, 0, 1)},
		{WithLimits(1, 1, 0)},
		{WithMaxConcurrency(0)},
		{WithRetryPolicy(-1, 0)},
		{WithRetryPolicy(0, -1)},
		{WithClock(nil)},
	} {
		if _, err := New("", options...); err == nil {
			t.Fatalf("options %#v accepted", options)
		}
	}
	if _, err := New("%zz"); err == nil {
		t.Fatal("malformed URL accepted")
	}
}

func TestMalformedProtocolDatesAndBudgets(t *testing.T) {
	tests := []struct {
		name, body string
		sets       bool
		want       string
	}{
		{name: "response date", body: `<OAI-PMH xmlns="http://www.openarchives.org/OAI/2.0/"><responseDate>invalid</responseDate><ListRecords/></OAI-PMH>`, want: "response date"},
		{name: "record token expiry", body: `<OAI-PMH xmlns="http://www.openarchives.org/OAI/2.0/"><responseDate>2026-07-15T00:00:00Z</responseDate><ListRecords><resumptionToken expirationDate="invalid">x</resumptionToken></ListRecords></OAI-PMH>`, want: "resumption expiry"},
		{name: "set token expiry", body: `<OAI-PMH xmlns="http://www.openarchives.org/OAI/2.0/"><responseDate>2026-07-15T00:00:00Z</responseDate><ListSets><resumptionToken expirationDate="invalid">x</resumptionToken></ListSets></OAI-PMH>`, sets: true, want: "resumption expiry"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client, server := newFixtureClient(t, func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, test.body) })
			defer server.Close()
			var err error
			if test.sets {
				_, err = client.ListSets(context.Background())
			} else {
				_, err = client.ListRecords(context.Background(), Request{MetadataPrefix: "oai_dc"})
			}
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v", err)
			}
		})
	}

	client, server := newFixtureClient(t, func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(fixture(t, "records_page1.xml")) }, WithLimits(1<<20, 1, 10))
	defer server.Close()
	if _, err := client.Harvest(context.Background(), Request{MetadataPrefix: "oai_dc"}); !errors.Is(err, ErrPageLimit) {
		t.Fatalf("error = %v", err)
	}
	setClient, setServer := newFixtureClient(t, func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(fixture(t, "sets_page1.xml")) }, WithLimits(1<<20, 1, 10))
	defer setServer.Close()
	if _, err := setClient.ListSets(context.Background()); !errors.Is(err, ErrPageLimit) {
		t.Fatalf("error = %v", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

type failingBody struct{ readErr, closeErr error }

func (body failingBody) Read([]byte) (int, error) { return 0, body.readErr }
func (body failingBody) Close() error             { return body.closeErr }

func TestTransportReadCloseAndRetryFailures(t *testing.T) {
	tests := []struct {
		name      string
		transport roundTripFunc
		want      string
	}{
		{name: "network", transport: func(*http.Request) (*http.Response, error) { return nil, errors.New("offline") }, want: "offline"},
		{name: "read", transport: func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Header: make(http.Header), Body: failingBody{readErr: errors.New("read failed")}}, nil
		}, want: "read failed"},
		{name: "close", transport: func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Header: make(http.Header), Body: failingBody{readErr: io.EOF, closeErr: errors.New("close failed")}}, nil
		}, want: "close failed"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			httpClient := &http.Client{Transport: test.transport}
			client, err := New("https://zenodo.org/oai2d", WithHTTPClient(httpClient), WithRetryPolicy(0, 0))
			if err != nil {
				t.Fatal(err)
			}
			_, err = client.ListMetadataFormats(context.Background(), "")
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestErrorAndTokenEdgeMethods(t *testing.T) {
	var protocol *ProtocolError
	var httpFailure *HTTPError
	if protocol.Error() != "<nil>" || httpFailure.Error() != "<nil>" {
		t.Fatal("nil error rendering changed")
	}
	if protocol.Unwrap() != nil {
		t.Fatal("nil protocol error unwrap changed")
	}
	if !(ResumptionToken{}).Empty() {
		t.Fatal("zero token should be empty")
	}
	client, err := New("")
	if err != nil {
		t.Fatal(err)
	}
	if err := client.validateToken(ResumptionToken{}); err == nil {
		t.Fatal("empty token accepted")
	}
}
