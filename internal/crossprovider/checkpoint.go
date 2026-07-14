package crossprovider

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/edithatogo/osf-cli-go/internal/auth"
)

var (
	ErrInvalidCheckpoint  = errors.New("invalid cross-provider checkpoint")
	ErrCheckpointMismatch = errors.New("cross-provider checkpoint does not match mapping")
	ErrIrreversible       = errors.New("cross-provider transfer has crossed an irreversible boundary")
)

// SagaStatus is the truthful aggregate state of a transfer checkpoint.
type SagaStatus string

const (
	SagaPending     SagaStatus = "pending"
	SagaRunning     SagaStatus = "running"
	SagaPartial     SagaStatus = "partial"
	SagaCompleted   SagaStatus = "completed"
	SagaCompensated SagaStatus = "compensated"
)

// StepKind identifies one ordered transfer operation.
type StepKind string

const (
	StepCreateDestination StepKind = "create_destination"
	StepApplyMetadata     StepKind = "apply_metadata"
	StepCopyFile          StepKind = "copy_file"
	StepVerifyDraft       StepKind = "verify_draft"
	StepFinalizeDraft     StepKind = "finalize_draft"
	StepPublish           StepKind = "publish"
)

// StepState is a replay-safe operation state.
type StepState string

const (
	StepPending     StepState = "pending"
	StepCompleted   StepState = "completed"
	StepFailed      StepState = "failed"
	StepCompensated StepState = "compensated"
)

// CompensationKind describes a reversible draft-only action.
type CompensationKind string

const (
	CompensationNone            CompensationKind = "none"
	CompensationDiscardDraft    CompensationKind = "discard_draft"
	CompensationDeleteFile      CompensationKind = "delete_file"
	CompensationRestoreFile     CompensationKind = "restore_file"
	CompensationRestoreMetadata CompensationKind = "restore_metadata"
)

// Step records one operation and enough information for replay/compensation.
type Step struct {
	ID                   string           `json:"id"`
	Kind                 StepKind         `json:"kind"`
	State                StepState        `json:"state"`
	File                 File             `json:"file,omitempty"`
	Attempts             int              `json:"attempts"`
	Error                string           `json:"error,omitempty"`
	DestinationRef       string           `json:"destinationRef,omitempty"`
	RollbackRef          string           `json:"rollbackRef,omitempty"`
	Compensation         CompensationKind `json:"compensation"`
	RequiresConfirmation bool             `json:"requiresConfirmation"`
}

// StepResult records non-secret output required by later or compensating steps.
type StepResult struct {
	DestinationRef string `json:"destinationRef,omitempty"`
	RollbackRef    string `json:"rollbackRef,omitempty"`
}

// Checkpoint is a versioned, serializable saga state.
type Checkpoint struct {
	SchemaVersion  int            `json:"schemaVersion"`
	IdempotencyKey string         `json:"idempotencyKey"`
	Destination    Destination    `json:"destination"`
	Conflict       ConflictPolicy `json:"conflict"`
	PublishIntent  PublishIntent  `json:"publishIntent"`
	Status         SagaStatus     `json:"status"`
	DestinationRef string         `json:"destinationRef,omitempty"`
	Steps          []Step         `json:"steps"`
}

// CompensationAction is an ordered recovery instruction.
type CompensationAction struct {
	StepID         string           `json:"stepId"`
	Kind           CompensationKind `json:"kind"`
	DestinationRef string           `json:"destinationRef,omitempty"`
	RollbackRef    string           `json:"rollbackRef,omitempty"`
	File           File             `json:"file,omitempty"`
}

// PartialResult reports file and lifecycle outcomes without claiming completion.
type PartialResult struct {
	Status         SagaStatus `json:"status"`
	DestinationRef string     `json:"destinationRef,omitempty"`
	CompletedFiles []File     `json:"completedFiles,omitempty"`
	FailedFiles    []File     `json:"failedFiles,omitempty"`
	PendingFiles   []File     `json:"pendingFiles,omitempty"`
	Published      bool       `json:"published"`
}

// NewCheckpoint constructs deterministic ordered steps from an executable report.
func NewCheckpoint(report Report) (Checkpoint, error) {
	if !report.Executable || strings.TrimSpace(report.IdempotencyKey) == "" {
		return Checkpoint{}, fmt.Errorf("%w: mapping report is not executable", ErrInvalidCheckpoint)
	}
	var steps []Step
	if report.Destination.CreateNew {
		steps = append(steps, Step{Kind: StepCreateDestination, State: StepPending, Compensation: CompensationDiscardDraft})
	}
	metadataCompensation := CompensationRestoreMetadata
	if report.Destination.CreateNew {
		metadataCompensation = CompensationNone
	}
	steps = append(steps, Step{Kind: StepApplyMetadata, State: StepPending, Compensation: metadataCompensation})
	for _, file := range sortedFiles(report.Files) {
		compensation := CompensationDeleteFile
		if report.Conflict == ConflictReplaceDraft {
			compensation = CompensationRestoreFile
		}
		steps = append(steps, Step{Kind: StepCopyFile, State: StepPending, File: file, Compensation: compensation})
	}
	steps = append(steps,
		Step{Kind: StepVerifyDraft, State: StepPending, Compensation: CompensationNone},
		Step{Kind: StepFinalizeDraft, State: StepPending, Compensation: CompensationNone},
	)
	if report.PublishIntent == PublishAfterCopy {
		steps = append(steps, Step{Kind: StepPublish, State: StepPending, Compensation: CompensationNone, RequiresConfirmation: true})
	}
	for i := range steps {
		steps[i].ID = deterministicStepID(report.IdempotencyKey, i, steps[i])
	}
	checkpoint := Checkpoint{
		SchemaVersion: 1, IdempotencyKey: report.IdempotencyKey,
		Destination: report.Destination, Conflict: report.Conflict, PublishIntent: report.PublishIntent,
		Status: SagaPending, DestinationRef: report.Destination.NativeID, Steps: steps,
	}
	if err := checkpoint.Validate(); err != nil {
		return Checkpoint{}, err
	}
	return checkpoint, nil
}

// Validate rejects tampered or internally inconsistent checkpoints.
func (checkpoint Checkpoint) Validate() error {
	if checkpoint.SchemaVersion != 1 || !strings.HasPrefix(checkpoint.IdempotencyKey, "xfer-v1-") || len(checkpoint.Steps) == 0 {
		return fmt.Errorf("%w: schema, idempotency key, and steps are required", ErrInvalidCheckpoint)
	}
	validStatus := checkpoint.Status == SagaPending || checkpoint.Status == SagaRunning || checkpoint.Status == SagaPartial || checkpoint.Status == SagaCompleted || checkpoint.Status == SagaCompensated
	if !validStatus {
		return fmt.Errorf("%w: status %q is invalid", ErrInvalidCheckpoint, checkpoint.Status)
	}
	if checkpoint.Conflict != ConflictFail && checkpoint.Conflict != ConflictSkipIdentical && checkpoint.Conflict != ConflictReplaceDraft {
		return fmt.Errorf("%w: conflict policy is invalid", ErrInvalidCheckpoint)
	}
	if checkpoint.PublishIntent != PublishDraftOnly && checkpoint.PublishIntent != PublishAfterCopy {
		return fmt.Errorf("%w: publish intent is invalid", ErrInvalidCheckpoint)
	}
	if checkpoint.Destination.CreateNew == (strings.TrimSpace(checkpoint.Destination.NativeID) != "") {
		return fmt.Errorf("%w: destination boundary is invalid", ErrInvalidCheckpoint)
	}
	wantKinds := expectedStepKinds(checkpoint)
	if len(checkpoint.Steps) != len(wantKinds) {
		return fmt.Errorf("%w: step count does not match destination and publish intent", ErrInvalidCheckpoint)
	}
	seenIncomplete := false
	hasFailed := false
	allComplete := true
	for i, step := range checkpoint.Steps {
		if step.Kind != wantKinds[i] {
			return fmt.Errorf("%w: step %d kind %q is out of sequence", ErrInvalidCheckpoint, i+1, step.Kind)
		}
		if step.ID != deterministicStepID(checkpoint.IdempotencyKey, i, step) {
			return fmt.Errorf("%w: step %d id is invalid", ErrInvalidCheckpoint, i+1)
		}
		if step.State != StepPending && step.State != StepCompleted && step.State != StepFailed && step.State != StepCompensated {
			return fmt.Errorf("%w: step %s state %q is invalid", ErrInvalidCheckpoint, step.ID, step.State)
		}
		if seenIncomplete && step.State == StepCompleted {
			return fmt.Errorf("%w: completed step follows an incomplete step", ErrInvalidCheckpoint)
		}
		if step.State != StepCompleted && step.State != StepCompensated {
			seenIncomplete = true
			allComplete = false
		}
		if step.State == StepFailed {
			hasFailed = true
		}
		if step.Compensation != expectedCompensation(checkpoint, step) {
			return fmt.Errorf("%w: step %s compensation %q is invalid", ErrInvalidCheckpoint, step.ID, step.Compensation)
		}
	}
	if hasFailed && checkpoint.Status != SagaPartial {
		return fmt.Errorf("%w: failed step requires partial status", ErrInvalidCheckpoint)
	}
	if allComplete && checkpoint.Status != SagaCompleted && checkpoint.Status != SagaCompensated {
		return fmt.Errorf("%w: completed steps require terminal status", ErrInvalidCheckpoint)
	}
	if !allComplete && !hasFailed && (checkpoint.Status == SagaCompleted || checkpoint.Status == SagaCompensated) {
		return fmt.Errorf("%w: incomplete steps cannot have terminal status", ErrInvalidCheckpoint)
	}
	return nil
}

// Complete advances exactly the next pending or failed step.
func (checkpoint *Checkpoint) Complete(stepID string, result StepResult) error {
	index, err := checkpoint.nextIndex(stepID)
	if err != nil {
		return err
	}
	step := &checkpoint.Steps[index]
	if step.Kind == StepCreateDestination && strings.TrimSpace(result.DestinationRef) == "" {
		return fmt.Errorf("%w: destination creation requires its native reference", ErrInvalidCheckpoint)
	}
	if step.Compensation == CompensationRestoreFile || step.Compensation == CompensationRestoreMetadata {
		if strings.TrimSpace(result.RollbackRef) == "" {
			return fmt.Errorf("%w: step %s requires a rollback reference", ErrInvalidCheckpoint, step.ID)
		}
	}
	step.State, step.Error = StepCompleted, ""
	step.Attempts++
	step.DestinationRef, step.RollbackRef = result.DestinationRef, result.RollbackRef
	if result.DestinationRef != "" {
		checkpoint.DestinationRef = result.DestinationRef
	}
	checkpoint.Status = SagaRunning
	if index == len(checkpoint.Steps)-1 {
		checkpoint.Status = SagaCompleted
	}
	return nil
}

// Fail records a redacted truthful partial result for the next executable step.
func (checkpoint *Checkpoint) Fail(stepID string, cause error) error {
	index, err := checkpoint.nextIndex(stepID)
	if err != nil {
		return err
	}
	if cause == nil {
		return fmt.Errorf("%w: failure cause is required", ErrInvalidCheckpoint)
	}
	checkpoint.Steps[index].State = StepFailed
	checkpoint.Steps[index].Attempts++
	checkpoint.Steps[index].Error = auth.Redact(cause.Error())
	checkpoint.Status = SagaPartial
	return nil
}

func (checkpoint *Checkpoint) nextIndex(stepID string) (int, error) {
	if checkpoint == nil {
		return -1, fmt.Errorf("%w: checkpoint is nil", ErrInvalidCheckpoint)
	}
	if err := checkpoint.Validate(); err != nil {
		return -1, err
	}
	for i := range checkpoint.Steps {
		state := checkpoint.Steps[i].State
		if state == StepCompleted || state == StepCompensated {
			continue
		}
		if checkpoint.Steps[i].ID != stepID {
			return -1, fmt.Errorf("%w: step %s is not the next operation", ErrInvalidCheckpoint, stepID)
		}
		return i, nil
	}
	return -1, fmt.Errorf("%w: checkpoint has no pending step", ErrInvalidCheckpoint)
}

// Resume validates report/checkpoint identity and returns the remaining steps.
func Resume(report Report, checkpoint Checkpoint) ([]Step, error) {
	if report.IdempotencyKey != checkpoint.IdempotencyKey {
		return nil, ErrCheckpointMismatch
	}
	if err := checkpoint.Validate(); err != nil {
		return nil, err
	}
	for i, step := range checkpoint.Steps {
		if step.State != StepCompleted && step.State != StepCompensated {
			return append([]Step(nil), checkpoint.Steps[i:]...), nil
		}
	}
	return nil, nil
}

// CompensationPlan returns reversible completed actions in reverse order.
func (checkpoint Checkpoint) CompensationPlan() ([]CompensationAction, error) {
	if err := checkpoint.Validate(); err != nil {
		return nil, err
	}
	for _, step := range checkpoint.Steps {
		if step.Kind == StepPublish && step.State == StepCompleted {
			return nil, ErrIrreversible
		}
	}
	var actions []CompensationAction
	for i := len(checkpoint.Steps) - 1; i >= 0; i-- {
		step := checkpoint.Steps[i]
		if step.State != StepCompleted || step.Compensation == CompensationNone {
			continue
		}
		actions = append(actions, CompensationAction{
			StepID: step.ID, Kind: step.Compensation, DestinationRef: checkpoint.DestinationRef,
			RollbackRef: step.RollbackRef, File: step.File,
		})
	}
	return actions, nil
}

// PartialResult returns file-level outcomes and whether publication occurred.
func (checkpoint Checkpoint) PartialResult() (PartialResult, error) {
	if err := checkpoint.Validate(); err != nil {
		return PartialResult{}, err
	}
	result := PartialResult{Status: checkpoint.Status, DestinationRef: checkpoint.DestinationRef}
	for _, step := range checkpoint.Steps {
		if step.Kind == StepPublish && step.State == StepCompleted {
			result.Published = true
		}
		if step.Kind != StepCopyFile {
			continue
		}
		switch step.State {
		case StepCompleted:
			result.CompletedFiles = append(result.CompletedFiles, step.File)
		case StepFailed:
			result.FailedFiles = append(result.FailedFiles, step.File)
		default:
			result.PendingFiles = append(result.PendingFiles, step.File)
		}
	}
	return result, nil
}

func deterministicStepID(key string, index int, step Step) string {
	input := fmt.Sprintf("%s\x00%d\x00%s\x00%s", key, index, step.Kind, step.File.Path)
	digest := sha256.Sum256([]byte(input))
	return "step-" + hex.EncodeToString(digest[:8])
}

func expectedStepKinds(checkpoint Checkpoint) []StepKind {
	kinds := make([]StepKind, 0, len(checkpoint.Steps))
	if checkpoint.Destination.CreateNew {
		kinds = append(kinds, StepCreateDestination)
	}
	kinds = append(kinds, StepApplyMetadata)
	for _, step := range checkpoint.Steps {
		if step.Kind == StepCopyFile {
			kinds = append(kinds, StepCopyFile)
		}
	}
	kinds = append(kinds, StepVerifyDraft, StepFinalizeDraft)
	if checkpoint.PublishIntent == PublishAfterCopy {
		kinds = append(kinds, StepPublish)
	}
	return kinds
}

func expectedCompensation(checkpoint Checkpoint, step Step) CompensationKind {
	switch step.Kind {
	case StepCreateDestination:
		return CompensationDiscardDraft
	case StepApplyMetadata:
		if !checkpoint.Destination.CreateNew {
			return CompensationRestoreMetadata
		}
	case StepCopyFile:
		if checkpoint.Conflict == ConflictReplaceDraft {
			return CompensationRestoreFile
		}
		return CompensationDeleteFile
	}
	return CompensationNone
}
