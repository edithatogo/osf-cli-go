package observability

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestJSONEmitterWritesStableRedactedEvent(t *testing.T) {
	var output strings.Builder
	emitter := NewJSONEmitter(&output, LevelInfo)
	ctx := WithOperationID(context.Background(), "op-test")
	Emit(ctx, emitter, Event{
		Level:    LevelError,
		Name:     "api.request",
		Provider: "zenodo",
		Fields:   map[string]any{"token": "secret-token", "path": "/Users/test/private.txt", "nested": map[string]any{"authorization": "Bearer secret"}},
		Error:    RedactedError(errors.New("request failed with Bearer abcdefghijklmnop")),
		Outcome:  OutcomeError,
	})
	var event Event
	if err := json.Unmarshal([]byte(output.String()), &event); err != nil {
		t.Fatalf("decode event: %v", err)
	}
	if event.SchemaVersion != SchemaVersion || event.OperationID != "op-test" || event.RequestID == "" || event.Provider != "zenodo" {
		t.Fatalf("event envelope=%+v", event)
	}
	encoded := output.String()
	for _, secret := range []string{"secret-token", "private.txt", "abcdefghijklmnop"} {
		if strings.Contains(encoded, secret) {
			t.Fatalf("event leaked %q: %s", secret, encoded)
		}
	}
}

func TestJSONEmitterNormalizesProviderCardinality(t *testing.T) {
	var output strings.Builder
	Emit(context.Background(), NewJSONEmitter(&output, LevelInfo), Event{Name: "api.request", Provider: "attacker.example"})
	var event Event
	if err := json.Unmarshal([]byte(output.String()), &event); err != nil {
		t.Fatal(err)
	}
	if event.Provider != "unknown" || strings.Contains(output.String(), "attacker.example") {
		t.Fatalf("event = %+v", event)
	}
}

func TestJSONEmitterRedactsDirectEventErrors(t *testing.T) {
	var output strings.Builder
	emitter := NewJSONEmitter(&output, LevelInfo)
	Emit(context.Background(), emitter, Event{
		Name:    "api.request",
		Level:   LevelError,
		Outcome: OutcomeError,
		Error:   &Error{Class: "internal", Message: "Authorization: Bearer osf_live_token_abc123def456ghi789xyz"},
	})
	if strings.Contains(output.String(), "osf_live_token_abc123def456ghi789xyz") {
		t.Fatalf("event leaked direct error secret: %s", output.String())
	}
}

func TestJSONEmitterHonorsMinimumLevel(t *testing.T) {
	var output strings.Builder
	emitter := NewJSONEmitter(&output, LevelWarn)
	Emit(context.Background(), emitter, Event{Name: "debug", Level: LevelInfo})
	Emit(context.Background(), emitter, Event{Name: "warning", Level: LevelWarn})
	if strings.Contains(output.String(), "debug") || !strings.Contains(output.String(), "warning") {
		t.Fatalf("level filtering output=%q", output.String())
	}
}

func TestClassifyErrorAndOpenFromEnv(t *testing.T) {
	for _, test := range []struct {
		err  error
		want string
	}{
		{context.Canceled, "canceled"},
		{context.DeadlineExceeded, "timeout"},
		{errors.New("missing OSF credentials"), "authentication"},
		{errors.New("invalid path"), "validation"},
	} {
		if got := ClassifyError(test.err); got != test.want {
			t.Fatalf("ClassifyError(%v)=%q, want %q", test.err, got, test.want)
		}
	}
	t.Setenv("OSF_EVENT_LOG", "stdout")
	if _, _, err := OpenFromEnv(io.Discard); err == nil {
		t.Fatal("stdout event destination returned nil error")
	}
}

func TestOpenFromEnvRestrictsExistingFile(t *testing.T) {
	destination := filepath.Join(t.TempDir(), "events.jsonl")
	if err := os.WriteFile(destination, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("OSF_EVENT_LOG", destination)
	_, closer, err := OpenFromEnv(io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = closer.Close() }()
	info, err := os.Stat(destination)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("permissions=%#o, want 0600", got)
	}
}

func TestRedactedErrorRemovesLocalPaths(t *testing.T) {
	err := RedactedError(errors.New("failed to open /Users/example/private/token.txt"))
	if strings.Contains(err.Message, "/Users/example/private") || !strings.Contains(err.Message, "[REDACTED_PATH]") {
		t.Fatalf("error=%+v", err)
	}
}

func TestObservabilityContractsCoverContextClassificationAndEndpointClasses(t *testing.T) {
	t.Parallel()
	//nolint:staticcheck // nil is an intentional input for the package's nil-safe context contract.
	if OperationID(nil) != "" || RequestID(nil) != "" || EmitterFromContext(nil) != nil {
		t.Fatal("nil context returned values")
	}
	ctx := WithRequestID(WithOperationID(context.Background(), " op-1 "), " req-1 ")
	if OperationID(ctx) != "op-1" || RequestID(ctx) != "req-1" {
		t.Fatalf("context ids = %q/%q", OperationID(ctx), RequestID(ctx))
	}
	if got := NewID(" test "); !strings.HasPrefix(got, "test-") {
		t.Fatalf("NewID = %q", got)
	}
	for _, test := range []struct {
		err  error
		want string
	}{
		{errors.New("forbidden operation"), "authorization"},
		{errors.New("resource not found"), "not_found"},
		{errors.New("decode json response"), "decode"},
		{errors.New("unexpected failure"), "internal"},
	} {
		if got := ClassifyError(test.err); got != test.want {
			t.Fatalf("ClassifyError(%v) = %q, want %q", test.err, got, test.want)
		}
	}
	for _, test := range []struct {
		value, want string
	}{
		{"https://files.osf.io/file", "storage"},
		{"https://api.osf.io/v2/nodes", "api"},
		{"https://example.test/path", "external"},
		{"not a URL", "unknown"},
		{"/relative", "unknown"},
	} {
		if got := EndpointClass(test.value); got != test.want {
			t.Fatalf("EndpointClass(%q) = %q, want %q", test.value, got, test.want)
		}
	}
	fields := redactMap(map[string]any{
		"password": "secret", "source": "/tmp/source", "error": "Bearer token",
		"fields": map[string]any{"api_key": "secret", "count": 3},
		"items":  []any{"/tmp/path", map[string]any{"credential": "secret"}},
	})
	if fields["password"] != "[REDACTED]" || fields["source"] != "[REDACTED_PATH]" {
		t.Fatalf("redacted fields = %#v", fields)
	}
	if normalizeLevel("debug") != LevelDebug || normalizeLevel("unknown") != LevelInfo || !levelAllowed(LevelError, LevelWarn) || levelAllowed(LevelDebug, LevelError) {
		t.Fatal("level policy mismatch")
	}
	NopEmitter{}.Emit(Event{})
}

func TestObservabilityEdgeContracts(t *testing.T) {
	var output strings.Builder
	emitter := NewJSONEmitter(nil, "")
	Emit(context.Background(), nil, Event{})
	Emit(context.Background(), emitter, Event{Provider: "cross-provider", Fields: map[string]any{
		"api_key": "secret", "source": "/tmp/source", "error": fmt.Errorf("bad /tmp/detail"),
		"fields": []any{"/tmp/item", map[string]any{"credential": "secret"}}, "count": 2,
	}})
	if emitter == nil {
		t.Fatal("nil writer emitter was not constructed")
	}

	for _, err := range []error{
		net.UnknownNetworkError("test"),
		errors.New("unauthorized request"),
		errors.New("forbidden request"),
		errors.New("resource not found"),
		errors.New("decode json response"),
		errors.New("required value"),
		errors.New("other failure"),
	} {
		if ClassifyError(err) == "" {
			t.Fatalf("ClassifyError(%v) returned empty class", err)
		}
	}
	for _, raw := range []string{"http://[::1", "https://files.osf.io:443/x", "https://api.osf.io/x", "relative"} {
		if EndpointClass(raw) == "" {
			t.Fatalf("EndpointClass(%q) returned empty class", raw)
		}
	}
	for _, level := range []string{"debug", "info", "warn", "error", "unknown"} {
		if normalizeLevel(level) == "" {
			t.Fatalf("normalizeLevel(%q) returned empty level", level)
		}
	}
	if got := normalizeProvider("  OSF "); got != "osf" || normalizeProvider("other") != "unknown" {
		t.Fatalf("provider normalization = %q", got)
	}
	if got := redactValue([]any{"/tmp/path", map[string]any{"secret": "value"}}); got == nil {
		t.Fatal("redactValue returned nil")
	}
	if got := redactMessage("C:\\Users\\person\\token.txt"); strings.Contains(got, "person") {
		t.Fatalf("redacted Windows path = %q", got)
	}

	t.Setenv("OSF_EVENT_LOG", "")
	if _, closer, err := OpenFromEnv(&output); err != nil {
		t.Fatal(err)
	} else if closer == nil {
		t.Fatal("disabled event log returned nil closer")
	}
	t.Setenv("OSF_EVENT_LOG", "stderr")
	em, closer, err := OpenFromEnv(&output)
	if err != nil || em == nil || closer == nil {
		t.Fatalf("stderr event log = %v %T %T", err, em, closer)
	}
	if err := closer.Close(); err != nil {
		t.Fatal(err)
	}
}
