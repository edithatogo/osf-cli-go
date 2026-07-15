package crossprovider

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNewCheckpointCreatesDeterministicDraftOnlySaga(t *testing.T) {
	t.Parallel()
	request := validRequest(t)
	request.Source.Files = append(request.Source.Files, File{Path: "a.txt", Size: 1, Checksum: "sha256:def"})
	report, err := BuildMapping(request, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	checkpoint, err := NewCheckpoint(report)
	if err != nil {
		t.Fatal(err)
	}
	want := []StepKind{StepCreateDestination, StepApplyMetadata, StepCopyFile, StepCopyFile, StepVerifyDraft, StepFinalizeDraft}
	if checkpoint.Status != SagaPending || len(checkpoint.Steps) != len(want) {
		t.Fatalf("checkpoint = %+v", checkpoint)
	}
	for i, kind := range want {
		if checkpoint.Steps[i].Kind != kind || checkpoint.Steps[i].State != StepPending || checkpoint.Steps[i].ID == "" {
			t.Fatalf("step %d = %+v", i, checkpoint.Steps[i])
		}
		if i > 0 && checkpoint.Steps[i-1].ID == checkpoint.Steps[i].ID {
			t.Fatal("step IDs are not unique")
		}
	}
	if checkpoint.Steps[2].File.Path != "a.txt" || checkpoint.Steps[3].File.Path != "data.csv" {
		t.Fatalf("file steps are not sorted: %+v", checkpoint.Steps)
	}
}

func TestPublicationIsSeparateAndIrreversible(t *testing.T) {
	t.Parallel()
	request := validRequest(t)
	request.PublishIntent = PublishAfterCopy
	report, err := BuildMapping(request, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	checkpoint, err := NewCheckpoint(report)
	if err != nil {
		t.Fatal(err)
	}
	last := checkpoint.Steps[len(checkpoint.Steps)-1]
	if last.Kind != StepPublish || !last.RequiresConfirmation || last.Compensation != CompensationNone {
		t.Fatalf("publish step = %+v", last)
	}
}

func TestCheckpointProgressPartialFailureAndReplay(t *testing.T) {
	t.Parallel()
	report, err := BuildMapping(validRequest(t), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	checkpoint, err := NewCheckpoint(report)
	if err != nil {
		t.Fatal(err)
	}
	first := checkpoint.Steps[0]
	if err := checkpoint.Complete(first.ID, StepResult{DestinationRef: "draft-123"}); err != nil {
		t.Fatal(err)
	}
	if checkpoint.Status != SagaRunning || checkpoint.DestinationRef != "draft-123" {
		t.Fatalf("checkpoint = %+v", checkpoint)
	}
	second := checkpoint.Steps[1]
	if err := checkpoint.Fail(second.ID, errors.New("Bearer top-secret-token")); err != nil {
		t.Fatal(err)
	}
	if checkpoint.Status != SagaPartial || strings.Contains(checkpoint.Steps[1].Error, "top-secret-token") || !strings.Contains(checkpoint.Steps[1].Error, "REDACTED") {
		t.Fatalf("partial checkpoint = %+v", checkpoint)
	}
	pending, err := Resume(report, checkpoint)
	if err != nil || len(pending) == 0 || pending[0].ID != second.ID {
		t.Fatalf("pending = %+v err=%v", pending, err)
	}

	changed := validRequest(t)
	changed.Source.Files[0].Checksum = "sha256:changed"
	changedReport, err := BuildMapping(changed, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Resume(changedReport, checkpoint); !errors.Is(err, ErrCheckpointMismatch) {
		t.Fatalf("resume mismatch error = %v", err)
	}
}

func TestCheckpointRejectsOutOfOrderAndDuplicateCompletion(t *testing.T) {
	t.Parallel()
	report, err := BuildMapping(validRequest(t), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	checkpoint, err := NewCheckpoint(report)
	if err != nil {
		t.Fatal(err)
	}
	if err := checkpoint.Complete(checkpoint.Steps[1].ID, StepResult{}); !errors.Is(err, ErrInvalidCheckpoint) {
		t.Fatalf("out-of-order error = %v", err)
	}
	if err := checkpoint.Complete(checkpoint.Steps[0].ID, StepResult{DestinationRef: "draft-123"}); err != nil {
		t.Fatal(err)
	}
	if err := checkpoint.Complete(checkpoint.Steps[0].ID, StepResult{}); !errors.Is(err, ErrInvalidCheckpoint) {
		t.Fatalf("duplicate error = %v", err)
	}
}

func TestCompensationPlanReversesCompletedDraftMutations(t *testing.T) {
	t.Parallel()
	report, err := BuildMapping(validRequest(t), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	checkpoint, err := NewCheckpoint(report)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		result := StepResult{}
		switch i {
		case 0:
			result.DestinationRef = "draft-123"
		case 2:
			result.DestinationRef = "file-456"
		}
		if err := checkpoint.Complete(checkpoint.Steps[i].ID, result); err != nil {
			t.Fatal(err)
		}
	}
	actions, err := checkpoint.CompensationPlan()
	if err != nil {
		t.Fatal(err)
	}
	if len(actions) != 2 || actions[0].Kind != CompensationDeleteFile || actions[0].DestinationRef != "draft-123" || actions[0].ResourceRef != "file-456" || actions[1].Kind != CompensationDiscardDraft {
		t.Fatalf("actions = %+v", actions)
	}
	if checkpoint.DestinationRef != "draft-123" {
		t.Fatalf("destination ref was overwritten: %s", checkpoint.DestinationRef)
	}
}

func TestCheckpointJSONRoundTripAndValidation(t *testing.T) {
	t.Parallel()
	report, err := BuildMapping(validRequest(t), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	want, err := NewCheckpoint(report)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var got Checkpoint
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatal(err)
	}
	if err := got.Validate(); err != nil || got.IdempotencyKey != want.IdempotencyKey {
		t.Fatalf("round trip = %+v err=%v", got, err)
	}
	got.Steps[0].ID = "tampered"
	if err := got.Validate(); !errors.Is(err, ErrInvalidCheckpoint) {
		t.Fatalf("tampered error = %v", err)
	}
}

func TestCheckpointRejectsTamperedCompensationAndStatus(t *testing.T) {
	t.Parallel()
	report, err := BuildMapping(validRequest(t), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	checkpoint, err := NewCheckpoint(report)
	if err != nil {
		t.Fatal(err)
	}
	checkpoint.Steps[0].Compensation = CompensationNone
	if err := checkpoint.Validate(); !errors.Is(err, ErrInvalidCheckpoint) {
		t.Fatalf("compensation error = %v", err)
	}
	checkpoint, _ = NewCheckpoint(report)
	checkpoint.Status = SagaCompleted
	if err := checkpoint.Validate(); !errors.Is(err, ErrInvalidCheckpoint) {
		t.Fatalf("status error = %v", err)
	}
}

func TestPartialResultReportsEveryFileTruthfully(t *testing.T) {
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
	if err := checkpoint.Complete(checkpoint.Steps[0].ID, StepResult{DestinationRef: "draft-123"}); err != nil {
		t.Fatal(err)
	}
	if err := checkpoint.Complete(checkpoint.Steps[1].ID, StepResult{}); err != nil {
		t.Fatal(err)
	}
	if err := checkpoint.Complete(checkpoint.Steps[2].ID, StepResult{}); err != nil {
		t.Fatal(err)
	}
	if err := checkpoint.Fail(checkpoint.Steps[3].ID, errors.New("copy failed")); err != nil {
		t.Fatal(err)
	}
	partial, err := checkpoint.PartialResult()
	if err != nil {
		t.Fatal(err)
	}
	if len(partial.CompletedFiles) != 1 || len(partial.FailedFiles) != 1 || partial.Published {
		t.Fatalf("partial = %+v", partial)
	}
}

func TestCompensatePartialSagaMarksUnexecutedStepsAbandoned(t *testing.T) {
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
	if err := checkpoint.Complete(checkpoint.Steps[0].ID, StepResult{DestinationRef: "draft-123"}); err != nil {
		t.Fatal(err)
	}
	if err := checkpoint.Complete(checkpoint.Steps[1].ID, StepResult{}); err != nil {
		t.Fatal(err)
	}
	if err := checkpoint.Fail(checkpoint.Steps[2].ID, errors.New("copy failed")); err != nil {
		t.Fatal(err)
	}
	compensated, err := Compensate(t.Context(), checkpoint, &fakeDestination{})
	if err != nil || compensated.Status != SagaCompensated {
		t.Fatalf("checkpoint = %+v err=%v", compensated, err)
	}
	partial, err := compensated.PartialResult()
	if err != nil || len(partial.AbandonedFiles) != 2 || len(partial.CompletedFiles) != 0 {
		t.Fatalf("partial = %+v err=%v", partial, err)
	}
}

func TestFileOrderingDoesNotChangeIdempotency(t *testing.T) {
	t.Parallel()
	request := validRequest(t)
	request.Source.Files = append(request.Source.Files, File{Path: "a.txt", Size: 1, Checksum: "sha256:def"})
	first, err := BuildMapping(request, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	request.Source.Files[0], request.Source.Files[1] = request.Source.Files[1], request.Source.Files[0]
	second, err := BuildMapping(request, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if first.IdempotencyKey != second.IdempotencyKey {
		t.Fatalf("keys differ: %s %s", first.IdempotencyKey, second.IdempotencyKey)
	}
}
