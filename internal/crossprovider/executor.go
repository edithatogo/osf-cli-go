package crossprovider

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/edithatogo/osf-cli-go/internal/auth"
)

var (
	ErrIntegrityMismatch           = errors.New("cross-provider destination integrity mismatch")
	ErrPublicationApprovalRequired = errors.New("cross-provider publication requires separate exact confirmation")
	ErrPublicationOutcomeUnknown   = errors.New("cross-provider publication outcome is unknown; inspect destination before retrying")
)

// SourceReader opens one file from the immutable source snapshot.
type SourceReader interface {
	Open(context.Context, File) (io.ReadCloser, error)
}

// DraftDestination performs idempotency-keyed, unpublished destination writes.
type DraftDestination interface {
	CreateDraft(context.Context, string) (string, error)
	ApplyMetadata(context.Context, string, Metadata, Provenance, string) (string, error)
	CopyFile(context.Context, string, File, io.Reader, ConflictPolicy, string) (FileReceipt, error)
	VerifyDraft(context.Context, string, Report, string) error
	FinalizeDraft(context.Context, string, Provenance, string) error
}

// Publisher is deliberately separate from DraftDestination.
type Publisher interface {
	Publish(context.Context, string, string) error
}

// Compensator applies explicit draft-only recovery actions.
type Compensator interface {
	Compensate(context.Context, CompensationAction) error
}

// FileReceipt is the destination's integrity and recovery acknowledgement.
type FileReceipt struct {
	ResourceRef string `json:"resourceRef"`
	RollbackRef string `json:"rollbackRef,omitempty"`
	Size        int64  `json:"size"`
	Checksum    string `json:"checksum"`
	Skipped     bool   `json:"skipped"`
}

// ExecutionResult always includes the latest truthful checkpoint and partial result.
type ExecutionResult struct {
	Checkpoint Checkpoint    `json:"checkpoint"`
	Partial    PartialResult `json:"partial"`
}

// PartialExecutionError identifies the step that stopped safe execution.
type PartialExecutionError struct {
	StepID string
	Cause  error
}

func (err *PartialExecutionError) Error() string {
	return fmt.Sprintf("cross-provider transfer stopped at step %s: %v", err.StepID, err.Cause)
}

func (err *PartialExecutionError) Unwrap() error { return err.Cause }

// Execute advances draft-only steps and never performs publication.
func Execute(ctx context.Context, report Report, checkpoint Checkpoint, source SourceReader, destination DraftDestination) (ExecutionResult, error) {
	if source == nil || destination == nil {
		return ExecutionResult{}, fmt.Errorf("%w: source and destination adapters are required", ErrInvalidCheckpoint)
	}
	pending, err := Resume(report, checkpoint)
	if err != nil {
		return ExecutionResult{}, err
	}
	for _, step := range pending {
		if step.Kind == StepPublish {
			result, resultErr := executionResult(checkpoint)
			if resultErr != nil {
				return ExecutionResult{}, resultErr
			}
			return result, ErrPublicationApprovalRequired
		}
		result, stepErr := executeDraftStep(ctx, report, &checkpoint, step, source, destination)
		if stepErr != nil {
			return failExecution(&checkpoint, step.ID, stepErr)
		}
		if err := checkpoint.Complete(step.ID, result); err != nil {
			return failExecution(&checkpoint, step.ID, err)
		}
	}
	return executionResult(checkpoint)
}

func executeDraftStep(ctx context.Context, report Report, checkpoint *Checkpoint, step Step, source SourceReader, destination DraftDestination) (StepResult, error) {
	draft := checkpoint.DestinationRef
	switch step.Kind {
	case StepCreateDestination:
		ref, err := destination.CreateDraft(ctx, step.ID)
		return StepResult{DestinationRef: ref}, err
	case StepApplyMetadata:
		rollback, err := destination.ApplyMetadata(ctx, draft, report.Target, report.Provenance, step.ID)
		return StepResult{RollbackRef: rollback}, err
	case StepCopyFile:
		reader, err := source.Open(ctx, step.File)
		if err != nil {
			return StepResult{}, fmt.Errorf("open source file %s: %w", step.File.Path, err)
		}
		receipt, copyErr := destination.CopyFile(ctx, draft, step.File, reader, report.Conflict, step.ID)
		closeErr := reader.Close()
		if copyErr != nil {
			if closeErr != nil {
				return StepResult{}, errors.Join(copyErr, fmt.Errorf("close source file %s: %w", step.File.Path, closeErr))
			}
			return StepResult{}, copyErr
		}
		if closeErr != nil {
			return StepResult{}, fmt.Errorf("close source file %s: %w", step.File.Path, closeErr)
		}
		if receipt.Size != step.File.Size || !strings.EqualFold(strings.TrimSpace(receipt.Checksum), strings.TrimSpace(step.File.Checksum)) {
			return StepResult{}, fmt.Errorf("%w: file %s expected size=%d checksum=%s, received size=%d checksum=%s", ErrIntegrityMismatch, step.File.Path, step.File.Size, step.File.Checksum, receipt.Size, receipt.Checksum)
		}
		if strings.TrimSpace(receipt.ResourceRef) == "" {
			return StepResult{}, fmt.Errorf("%w: file %s destination resource reference is missing", ErrIntegrityMismatch, step.File.Path)
		}
		return StepResult{DestinationRef: receipt.ResourceRef, RollbackRef: receipt.RollbackRef}, nil
	case StepVerifyDraft:
		return StepResult{}, destination.VerifyDraft(ctx, draft, report, step.ID)
	case StepFinalizeDraft:
		return StepResult{}, destination.FinalizeDraft(ctx, draft, report.Provenance, step.ID)
	default:
		return StepResult{}, fmt.Errorf("%w: unsupported draft step %q", ErrInvalidCheckpoint, step.Kind)
	}
}

func failExecution(checkpoint *Checkpoint, stepID string, cause error) (ExecutionResult, error) {
	if err := checkpoint.Fail(stepID, cause); err != nil {
		return ExecutionResult{}, errors.Join(cause, err)
	}
	result, err := executionResult(*checkpoint)
	if err != nil {
		return ExecutionResult{}, errors.Join(cause, err)
	}
	return result, &PartialExecutionError{StepID: stepID, Cause: safeExecutionCause(cause)}
}

func safeExecutionCause(cause error) error {
	message := auth.Redact(cause.Error())
	for _, sentinel := range []error{ErrIntegrityMismatch, ErrPublicationOutcomeUnknown} {
		if errors.Is(cause, sentinel) {
			return fmt.Errorf("%w: %s", sentinel, message)
		}
	}
	return errors.New(message)
}

func executionResult(checkpoint Checkpoint) (ExecutionResult, error) {
	partial, err := checkpoint.PartialResult()
	if err != nil {
		return ExecutionResult{}, err
	}
	return ExecutionResult{Checkpoint: checkpoint, Partial: partial}, nil
}

// PublicationChallenge returns the only accepted confirmation for the pending publish step.
func PublicationChallenge(checkpoint Checkpoint) (string, error) {
	if err := checkpoint.Validate(); err != nil {
		return "", err
	}
	pending, err := Resume(Report{IdempotencyKey: checkpoint.IdempotencyKey}, checkpoint)
	if err != nil {
		return "", err
	}
	if len(pending) != 1 || pending[0].Kind != StepPublish || strings.TrimSpace(checkpoint.DestinationRef) == "" {
		return "", ErrPublicationApprovalRequired
	}
	return "publish:" + checkpoint.IdempotencyKey + ":" + checkpoint.DestinationRef, nil
}

// Publish executes only an already-pending publication step with exact confirmation.
func Publish(ctx context.Context, report Report, checkpoint Checkpoint, publisher Publisher, confirmation string) (ExecutionResult, error) {
	if publisher == nil || report.IdempotencyKey != checkpoint.IdempotencyKey {
		return ExecutionResult{}, ErrCheckpointMismatch
	}
	challenge, err := PublicationChallenge(checkpoint)
	if err != nil {
		return ExecutionResult{}, err
	}
	if confirmation != challenge {
		return ExecutionResult{}, ErrPublicationApprovalRequired
	}
	pending, _ := Resume(report, checkpoint)
	step := pending[0]
	if err := publisher.Publish(ctx, checkpoint.DestinationRef, confirmation); err != nil {
		result, failErr := failExecution(&checkpoint, step.ID, ErrPublicationOutcomeUnknown)
		return result, errors.Join(failErr, safeExecutionCause(err))
	}
	if err := checkpoint.Complete(step.ID, StepResult{}); err != nil {
		return failExecution(&checkpoint, step.ID, err)
	}
	return executionResult(checkpoint)
}

// Compensate applies the reverse draft-only plan and marks each recovered step.
func Compensate(ctx context.Context, checkpoint Checkpoint, compensator Compensator) (Checkpoint, error) {
	if compensator == nil {
		return checkpoint, errors.New("cross-provider compensator is required")
	}
	actions, err := checkpoint.CompensationPlan()
	if err != nil {
		return checkpoint, err
	}
	checkpoint.Status = SagaCompensating
	for _, action := range actions {
		if err := compensator.Compensate(ctx, action); err != nil {
			checkpoint.Status = SagaCompensationFailed
			return checkpoint, fmt.Errorf("compensate step %s: %w", action.StepID, err)
		}
		for i := range checkpoint.Steps {
			if checkpoint.Steps[i].ID == action.StepID {
				checkpoint.Steps[i].State = StepCompensated
				break
			}
		}
	}
	for i := range checkpoint.Steps {
		if checkpoint.Steps[i].State == StepPending || checkpoint.Steps[i].State == StepFailed {
			checkpoint.Steps[i].State = StepAbandoned
		}
	}
	checkpoint.Status = SagaCompensated
	return checkpoint, checkpoint.Validate()
}
