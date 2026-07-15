package observability

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenFromEnvSupportsFileAndRejectsStdout(t *testing.T) {
	t.Setenv("OSF_EVENT_LEVEL", LevelDebug)
	t.Setenv("OSF_EVENT_LOG", "stdout")
	if _, _, err := OpenFromEnv(nil); err == nil {
		t.Fatal("stdout event log returned nil error")
	}

	path := filepath.Join(t.TempDir(), "nested", "events.jsonl")
	t.Setenv("OSF_EVENT_LOG", path)
	emitter, closer, err := OpenFromEnv(nil)
	if err != nil {
		t.Fatalf("OpenFromEnv(file): %v", err)
	}
	Emit(context.Background(), emitter, Event{Name: "test.event"})
	if err := closer.Close(); err != nil {
		t.Fatalf("close file emitter: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil || !strings.Contains(string(data), `"name":"test.event"`) {
		t.Fatalf("event log=%q err=%v", data, err)
	}
}

func TestObservabilityClassifiesAndRedactsErrors(t *testing.T) {
	if got := ClassifyError(nil); got != "" {
		t.Fatalf("ClassifyError(nil)=%q", got)
	}
	for _, test := range []struct{ err, want string }{
		{context.Canceled.Error(), "canceled"},
		{context.DeadlineExceeded.Error(), "timeout"},
		{"unauthorized request", "authentication"},
		{"forbidden request", "authorization"},
		{"not found", "not_found"},
		{"invalid input", "validation"},
		{"decode json", "decode"},
		{"unexpected failure", "internal"},
	} {
		err := error(errors.New(test.err))
		if test.err == context.Canceled.Error() {
			err = context.Canceled
		}
		if test.err == context.DeadlineExceeded.Error() {
			err = context.DeadlineExceeded
		}
		if got := ClassifyError(err); got != test.want {
			t.Errorf("ClassifyError(%q)=%q, want %q", test.err, got, test.want)
		}
	}
	if RedactedError(nil) != nil {
		t.Fatal("RedactedError(nil) returned a value")
	}
	redacted := RedactedError(errors.New("unauthorized token=secret"), "secret")
	if redacted == nil || redacted.Class != "authentication" || strings.Contains(redacted.Message, "secret") {
		t.Fatalf("redacted error=%+v", redacted)
	}
}
