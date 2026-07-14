package crossprovider

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"
)

type memorySource struct{ files map[string][]byte }

func (source memorySource) Open(_ context.Context, file File) (io.ReadCloser, error) {
	content, ok := source.files[file.Path]
	if !ok {
		return nil, errors.New("missing source file")
	}
	return io.NopCloser(bytes.NewReader(content)), nil
}

type fakeDestination struct {
	calls         []string
	failPath      string
	badChecksum   bool
	published     bool
	publishErr    error
	compensateErr error
}

func (destination *fakeDestination) CreateDraft(_ context.Context, stepID string) (string, error) {
	destination.calls = append(destination.calls, "create:"+stepID)
	return "draft-123", nil
}

func (destination *fakeDestination) ApplyMetadata(_ context.Context, draft string, _ Metadata, _ Provenance, stepID string) (string, error) {
	destination.calls = append(destination.calls, "metadata:"+draft+":"+stepID)
	return "", nil
}

func (destination *fakeDestination) CopyFile(_ context.Context, draft string, file File, reader io.Reader, _ ConflictPolicy, stepID string) (FileReceipt, error) {
	destination.calls = append(destination.calls, "file:"+file.Path+":"+stepID)
	if file.Path == destination.failPath {
		return FileReceipt{}, errors.New("injected copy failure")
	}
	content, err := io.ReadAll(reader)
	if err != nil {
		return FileReceipt{}, err
	}
	checksum := file.Checksum
	if destination.badChecksum {
		checksum = "sha256:wrong"
	}
	return FileReceipt{ResourceRef: "remote-" + file.Path, Size: int64(len(content)), Checksum: checksum}, nil
}

func (destination *fakeDestination) VerifyDraft(_ context.Context, draft string, _ Report, stepID string) error {
	destination.calls = append(destination.calls, "verify:"+draft+":"+stepID)
	return nil
}

func (destination *fakeDestination) FinalizeDraft(_ context.Context, draft string, _ Provenance, stepID string) error {
	destination.calls = append(destination.calls, "finalize:"+draft+":"+stepID)
	return nil
}

func (destination *fakeDestination) Compensate(_ context.Context, action CompensationAction) error {
	destination.calls = append(destination.calls, "compensate:"+string(action.Kind))
	return destination.compensateErr
}

func (destination *fakeDestination) Publish(_ context.Context, draft, confirmation string) error {
	destination.calls = append(destination.calls, "publish:"+draft+":"+confirmation)
	destination.published = true
	return destination.publishErr
}

func TestExecutorCopiesDraftAndVerifiesReceipts(t *testing.T) {
	t.Parallel()
	report, checkpoint := executableFixture(t, PublishDraftOnly)
	source := memorySource{files: map[string][]byte{"data.csv": bytes.Repeat([]byte("x"), 42)}}
	destination := &fakeDestination{}
	result, err := Execute(t.Context(), report, checkpoint, source, destination)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if result.Checkpoint.Status != SagaCompleted || result.Partial.Published || len(result.Partial.CompletedFiles) != 1 {
		t.Fatalf("result = %+v", result)
	}
	if destination.published || !containsCall(destination.calls, "finalize:") {
		t.Fatalf("calls = %+v", destination.calls)
	}
}

func TestExecutorFailureAndReplayDoNotRepeatCompletedWrites(t *testing.T) {
	t.Parallel()
	request := validRequest(t)
	request.Source.Files = append(request.Source.Files, File{Path: "later.txt", Size: 1, Checksum: "sha256:def"})
	report, err := BuildMapping(request, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	checkpoint, err := NewCheckpoint(report)
	if err != nil {
		t.Fatal(err)
	}
	source := memorySource{files: map[string][]byte{"data.csv": bytes.Repeat([]byte("x"), 42), "later.txt": {0}}}
	destination := &fakeDestination{failPath: "later.txt"}
	failed, err := Execute(t.Context(), report, checkpoint, source, destination)
	var partial *PartialExecutionError
	if !errors.As(err, &partial) || failed.Checkpoint.Status != SagaPartial {
		t.Fatalf("failed = %+v err=%v", failed, err)
	}
	before := append([]string(nil), destination.calls...)
	destination.failPath = ""
	resumed, err := Execute(t.Context(), report, failed.Checkpoint, source, destination)
	if err != nil || resumed.Checkpoint.Status != SagaCompleted {
		t.Fatalf("resumed = %+v err=%v", resumed, err)
	}
	newCalls := destination.calls[len(before):]
	if containsCall(newCalls, "create:") || containsCall(newCalls, "metadata:") || countCall(destination.calls, "file:data.csv:") != 1 {
		t.Fatalf("replay calls = %+v", destination.calls)
	}
}

func TestExecutorRejectsIntegrityMismatchAsPartial(t *testing.T) {
	t.Parallel()
	report, checkpoint := executableFixture(t, PublishDraftOnly)
	source := memorySource{files: map[string][]byte{"data.csv": bytes.Repeat([]byte("x"), 42)}}
	result, err := Execute(t.Context(), report, checkpoint, source, &fakeDestination{badChecksum: true})
	if !errors.Is(err, ErrIntegrityMismatch) || result.Checkpoint.Status != SagaPartial || len(result.Partial.FailedFiles) != 1 {
		t.Fatalf("result = %+v err=%v", result, err)
	}
}

func TestExecutorStopsBeforePublicationUntilSeparateConfirmation(t *testing.T) {
	t.Parallel()
	report, checkpoint := executableFixture(t, PublishAfterCopy)
	source := memorySource{files: map[string][]byte{"data.csv": bytes.Repeat([]byte("x"), 42)}}
	destination := &fakeDestination{}
	result, err := Execute(t.Context(), report, checkpoint, source, destination)
	if !errors.Is(err, ErrPublicationApprovalRequired) || destination.published || result.Checkpoint.Status != SagaRunning {
		t.Fatalf("result = %+v err=%v calls=%+v", result, err, destination.calls)
	}
	challenge, err := PublicationChallenge(result.Checkpoint)
	if err != nil || challenge == "" {
		t.Fatalf("challenge = %q err=%v", challenge, err)
	}
	if _, err := Publish(t.Context(), report, result.Checkpoint, destination, "wrong"); !errors.Is(err, ErrPublicationApprovalRequired) {
		t.Fatalf("wrong confirmation error = %v", err)
	}
	completed, err := Publish(t.Context(), report, result.Checkpoint, destination, challenge)
	if err != nil || !destination.published || completed.Checkpoint.Status != SagaCompleted || !completed.Partial.Published {
		t.Fatalf("completed = %+v err=%v", completed, err)
	}
}

func TestPublishFailureReportsUnknownOutcome(t *testing.T) {
	t.Parallel()
	report, checkpoint := executableFixture(t, PublishAfterCopy)
	source := memorySource{files: map[string][]byte{"data.csv": bytes.Repeat([]byte("x"), 42)}}
	destination := &fakeDestination{publishErr: errors.New("Bearer hidden-token")}
	ready, err := Execute(t.Context(), report, checkpoint, source, destination)
	if !errors.Is(err, ErrPublicationApprovalRequired) {
		t.Fatal(err)
	}
	challenge, _ := PublicationChallenge(ready.Checkpoint)
	failed, err := Publish(t.Context(), report, ready.Checkpoint, destination, challenge)
	if !errors.Is(err, ErrPublicationOutcomeUnknown) || failed.Partial.PublicationOutcome != "unknown" || strings.Contains(err.Error(), "hidden-token") {
		t.Fatalf("failed = %+v err=%v", failed, err)
	}
}

func TestCompensationFailureReturnsValidCheckpoint(t *testing.T) {
	t.Parallel()
	report, checkpoint := executableFixture(t, PublishDraftOnly)
	source := memorySource{files: map[string][]byte{"data.csv": bytes.Repeat([]byte("x"), 42)}}
	destination := &fakeDestination{}
	result, err := Execute(t.Context(), report, checkpoint, source, destination)
	if err != nil {
		t.Fatal(err)
	}
	destination.compensateErr = errors.New("injected compensation failure")
	failed, err := Compensate(t.Context(), result.Checkpoint, destination)
	if err == nil || failed.Status != SagaCompensationFailed {
		t.Fatalf("failed = %+v err=%v", failed, err)
	}
	if err := failed.Validate(); err != nil {
		t.Fatalf("failed checkpoint is invalid: %v", err)
	}
}

func TestCompensateAppliesReversePlan(t *testing.T) {
	t.Parallel()
	report, checkpoint := executableFixture(t, PublishDraftOnly)
	source := memorySource{files: map[string][]byte{"data.csv": bytes.Repeat([]byte("x"), 42)}}
	destination := &fakeDestination{}
	result, err := Execute(t.Context(), report, checkpoint, source, destination)
	if err != nil {
		t.Fatal(err)
	}
	compensated, err := Compensate(t.Context(), result.Checkpoint, destination)
	if err != nil || compensated.Status != SagaCompensated {
		t.Fatalf("compensated = %+v err=%v", compensated, err)
	}
	if countCall(destination.calls, "compensate:") != 2 {
		t.Fatalf("calls = %+v", destination.calls)
	}
}

func executableFixture(t *testing.T, intent PublishIntent) (Report, Checkpoint) {
	t.Helper()
	request := validRequest(t)
	request.PublishIntent = intent
	report, err := BuildMapping(request, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	checkpoint, err := NewCheckpoint(report)
	if err != nil {
		t.Fatal(err)
	}
	return report, checkpoint
}

func containsCall(calls []string, prefix string) bool {
	for _, call := range calls {
		if strings.HasPrefix(call, prefix) {
			return true
		}
	}
	return false
}

func countCall(calls []string, prefix string) int {
	count := 0
	for _, call := range calls {
		if strings.HasPrefix(call, prefix) {
			count++
		}
	}
	return count
}
