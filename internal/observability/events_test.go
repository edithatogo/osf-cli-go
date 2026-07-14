package observability

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestJSONEmitterWritesStableRedactedEvent(t *testing.T) {
	var output strings.Builder
	emitter := NewJSONEmitter(&output, LevelInfo)
	ctx := WithOperationID(context.Background(), "op-test")
	Emit(ctx, emitter, Event{
		Level:   LevelError,
		Name:    "api.request",
		Fields:  map[string]any{"token": "secret-token", "path": "/Users/test/private.txt", "nested": map[string]any{"authorization": "Bearer secret"}},
		Error:   RedactedError(errors.New("request failed with Bearer abcdefghijklmnop")),
		Outcome: OutcomeError,
	})
	var event Event
	if err := json.Unmarshal([]byte(output.String()), &event); err != nil {
		t.Fatalf("decode event: %v", err)
	}
	if event.SchemaVersion != SchemaVersion || event.OperationID != "op-test" || event.RequestID == "" {
		t.Fatalf("event envelope=%+v", event)
	}
	encoded := output.String()
	for _, secret := range []string{"secret-token", "private.txt", "abcdefghijklmnop"} {
		if strings.Contains(encoded, secret) {
			t.Fatalf("event leaked %q: %s", secret, encoded)
		}
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

func TestRedactedErrorRemovesLocalPaths(t *testing.T) {
	err := RedactedError(errors.New("failed to open /Users/example/private/token.txt"))
	if strings.Contains(err.Message, "/Users/example/private") || !strings.Contains(err.Message, "[REDACTED_PATH]") {
		t.Fatalf("error=%+v", err)
	}
}
