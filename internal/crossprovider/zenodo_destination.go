package crossprovider

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/edithatogo/osf-cli-go/internal/download"
	"github.com/edithatogo/osf-cli-go/internal/zenodopublish"
	"github.com/edithatogo/osf-cli-go/internal/zenodotransfer"
)

const provenanceFilename = ".osf-cli-go-provenance.json"

// ZenodoSandboxDestination applies cross-provider writes to unpublished Zenodo
// sandbox depositions. It intentionally does not implement Publisher.
type ZenodoSandboxDestination struct {
	transfer *zenodotransfer.Client
	publish  *zenodopublish.Client
	now      func() time.Time

	mu       sync.Mutex
	drafts   map[string]zenodotransfer.Draft
	created  map[string]bool
	receipts map[string]FileReceipt
	files    map[string]map[string]string
}

// NewZenodoSandboxDestination creates a deposit:write-only destination.
func NewZenodoSandboxDestination(baseURL, token string) (*ZenodoSandboxDestination, error) {
	transfer, err := zenodotransfer.New(baseURL, token)
	if err != nil {
		return nil, err
	}
	publish, err := zenodopublish.New(baseURL, token, []zenodopublish.Scope{zenodopublish.ScopeDepositWrite})
	if err != nil {
		return nil, err
	}
	return &ZenodoSandboxDestination{
		transfer: transfer, publish: publish, now: time.Now,
		drafts: make(map[string]zenodotransfer.Draft), created: make(map[string]bool), receipts: make(map[string]FileReceipt),
		files: make(map[string]map[string]string),
	}, nil
}

func (destination *ZenodoSandboxDestination) CreateDraft(ctx context.Context, stepID string) (string, error) {
	if draft, ok := destination.draftForStep(stepID); ok {
		return draft.ID, nil
	}
	draft, err := destination.transfer.CreateDraft(ctx)
	if err != nil {
		return "", err
	}
	destination.mu.Lock()
	destination.drafts[stepID] = draft
	destination.drafts[draft.ID] = draft
	destination.created[draft.ID] = true
	destination.mu.Unlock()
	return draft.ID, nil
}

func (destination *ZenodoSandboxDestination) ApplyMetadata(ctx context.Context, draftID string, metadata Metadata, _ Provenance, _ string) (string, error) {
	destination.mu.Lock()
	created := destination.created[draftID]
	destination.mu.Unlock()
	if !created {
		return "", errors.New("existing Zenodo draft metadata requires a durable rollback snapshot")
	}
	converted, err := zenodoMetadata(metadata, destination.now())
	if err != nil {
		return "", err
	}
	if err := destination.publish.ApplyDraftMetadata(ctx, draftID, converted, destination.now()); err != nil {
		return "", err
	}
	return "", nil
}

func (destination *ZenodoSandboxDestination) CopyFile(ctx context.Context, draftID string, file File, reader io.Reader, conflict ConflictPolicy, stepID string) (FileReceipt, error) {
	if conflict == ConflictReplaceDraft {
		return FileReceipt{}, errors.New("zenodo draft replacement requires a durable rollback snapshot")
	}
	destination.mu.Lock()
	if receipt, ok := destination.receipts[stepID]; ok {
		destination.mu.Unlock()
		return receipt, nil
	}
	destination.mu.Unlock()

	temporary, err := os.CreateTemp("", "osf-cli-go-cross-provider-*")
	if err != nil {
		return FileReceipt{}, fmt.Errorf("create cross-provider staging file: %w", err)
	}
	temporaryName := temporary.Name()
	defer func() { _ = os.Remove(temporaryName) }()

	written, copyErr := io.Copy(temporary, io.LimitReader(reader, file.Size+1))
	closeErr := temporary.Close()
	if copyErr != nil || closeErr != nil {
		return FileReceipt{}, errors.Join(copyErr, closeErr)
	}
	if written != file.Size {
		return FileReceipt{}, fmt.Errorf("%w: source %s size changed: got %d, want %d", ErrIntegrityMismatch, file.Path, written, file.Size)
	}
	if err := verifyFileChecksum(temporaryName, file.Checksum); err != nil {
		return FileReceipt{}, fmt.Errorf("%w: source %s: %v", ErrIntegrityMismatch, file.Path, err)
	}
	draft, err := destination.resolveDraft(ctx, draftID)
	if err != nil {
		return FileReceipt{}, err
	}
	policy, err := zenodoConflict(conflict)
	if err != nil {
		return FileReceipt{}, err
	}
	remoteName := zenodoRemoteName(file.Path)
	result, err := destination.transfer.UploadFile(ctx, draft, temporaryName, remoteName, policy)
	if err != nil {
		return FileReceipt{}, err
	}
	if result.Skipped {
		localMD5, hashErr := fileDigest(temporaryName, "md5")
		if hashErr != nil || !strings.EqualFold(localMD5, strings.TrimPrefix(result.Remote.Checksum, "md5:")) {
			return FileReceipt{}, fmt.Errorf("%w: skipped Zenodo file %s is not identical", ErrIntegrityMismatch, file.Path)
		}
	}
	receipt := FileReceipt{
		ResourceRef: result.Remote.ID, Size: file.Size, Checksum: file.Checksum,
		Skipped: result.Skipped,
	}
	destination.mu.Lock()
	destination.receipts[stepID] = receipt
	if destination.files[draftID] == nil {
		destination.files[draftID] = make(map[string]string)
	}
	destination.files[draftID][file.Path] = remoteName
	destination.mu.Unlock()
	return receipt, nil
}

func (destination *ZenodoSandboxDestination) VerifyDraft(ctx context.Context, draftID string, report Report, _ string) error {
	remote, err := destination.transfer.ListDraftFiles(ctx, draftID)
	if err != nil {
		return err
	}
	inventory := make(map[string]zenodotransfer.RemoteFile, len(remote))
	for _, file := range remote {
		inventory[file.Name] = file
	}
	for _, source := range report.Files {
		remoteFile, ok := inventory[zenodoRemoteName(source.Path)]
		if !ok || remoteFile.Size != source.Size {
			return fmt.Errorf("%w: Zenodo draft file %s is missing or has the wrong size", ErrIntegrityMismatch, source.Path)
		}
	}
	return nil
}

func (destination *ZenodoSandboxDestination) FinalizeDraft(ctx context.Context, draftID string, provenance Provenance, stepID string) error {
	destination.mu.Lock()
	fileMap := make(map[string]string, len(destination.files[draftID]))
	for source, remote := range destination.files[draftID] {
		fileMap[source] = remote
	}
	destination.mu.Unlock()
	payload := struct {
		SchemaVersion int               `json:"schemaVersion"`
		StepID        string            `json:"stepId"`
		Provenance    Provenance        `json:"provenance"`
		Files         map[string]string `json:"files"`
	}{SchemaVersion: 1, StepID: stepID, Provenance: provenance, Files: fileMap}
	encoded, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	encoded = append(encoded, '\n')
	temporary, err := os.CreateTemp("", "osf-cli-go-provenance-*.json")
	if err != nil {
		return err
	}
	name := temporary.Name()
	defer func() { _ = os.Remove(name) }()
	if _, err := temporary.Write(encoded); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	draft, err := destination.resolveDraft(ctx, draftID)
	if err != nil {
		return err
	}
	result, err := destination.transfer.UploadFile(ctx, draft, name, provenanceFilename, download.ConflictOverwrite)
	if err != nil {
		return err
	}
	if !result.Completed || result.Remote.Size != int64(len(encoded)) {
		return fmt.Errorf("%w: provenance sidecar was not verified", ErrIntegrityMismatch)
	}
	return nil
}

func (destination *ZenodoSandboxDestination) Compensate(ctx context.Context, action CompensationAction) error {
	switch action.Kind {
	case CompensationDeleteFile:
		return destination.transfer.DeleteFile(ctx, action.DestinationRef, action.ResourceRef)
	case CompensationDiscardDraft:
		return destination.transfer.DeleteDraft(ctx, action.DestinationRef)
	case CompensationNone:
		return nil
	case CompensationRestoreFile, CompensationRestoreMetadata:
		return errors.New("zenodo sandbox destination cannot restore overwritten draft state")
	default:
		return fmt.Errorf("unsupported Zenodo compensation %q", action.Kind)
	}
}

func (destination *ZenodoSandboxDestination) draftForStep(stepID string) (zenodotransfer.Draft, bool) {
	destination.mu.Lock()
	defer destination.mu.Unlock()
	draft, ok := destination.drafts[stepID]
	return draft, ok
}

func (destination *ZenodoSandboxDestination) resolveDraft(ctx context.Context, draftID string) (zenodotransfer.Draft, error) {
	destination.mu.Lock()
	draft, ok := destination.drafts[draftID]
	destination.mu.Unlock()
	if ok {
		return draft, nil
	}
	draft, err := destination.transfer.GetDraft(ctx, draftID)
	if err != nil {
		return zenodotransfer.Draft{}, fmt.Errorf("resolve existing Zenodo draft: %w", err)
	}
	destination.mu.Lock()
	destination.drafts[draftID] = draft
	destination.mu.Unlock()
	return draft, nil
}

func zenodoMetadata(metadata Metadata, now time.Time) (zenodopublish.Metadata, error) {
	converted := zenodopublish.Metadata{
		Title: metadata.Title, Description: metadata.Description, UploadType: metadata.UploadType,
		Keywords: append([]string(nil), metadata.Keywords...), License: metadata.License,
		AccessConditions: metadata.Access.Conditions,
	}
	for _, creator := range metadata.Creators {
		converted.Creators = append(converted.Creators, zenodopublish.Creator{Name: creator.Name})
	}
	switch metadata.Access.Kind {
	case AccessOpen:
		converted.Access = zenodopublish.AccessOpen
	case AccessEmbargoed:
		converted.Access, converted.EmbargoDate = zenodopublish.AccessEmbargoed, metadata.Access.EmbargoUntil
	case AccessRestricted:
		converted.Access = zenodopublish.AccessRestricted
	case AccessClosed:
		converted.Access = zenodopublish.AccessClosed
	default:
		return converted, fmt.Errorf("unsupported Zenodo target access %q", metadata.Access.Kind)
	}
	if err := zenodopublish.ValidateMetadata(converted, now); err != nil {
		return converted, err
	}
	return converted, nil
}

func zenodoConflict(conflict ConflictPolicy) (download.ConflictPolicy, error) {
	switch conflict {
	case ConflictFail:
		return download.ConflictFail, nil
	case ConflictSkipIdentical:
		return download.ConflictSkip, nil
	case ConflictReplaceDraft:
		return download.ConflictOverwrite, nil
	default:
		return "", fmt.Errorf("invalid Zenodo conflict policy %q", conflict)
	}
}

func zenodoRemoteName(sourcePath string) string {
	clean := path.Clean(sourcePath)
	if !strings.Contains(clean, "/") && clean != provenanceFilename {
		return clean
	}
	return "osfpath-" + base64.RawURLEncoding.EncodeToString([]byte(clean))
}

func verifyFileChecksum(filename, checksum string) error {
	algorithm, expected, ok := strings.Cut(strings.TrimSpace(checksum), ":")
	if !ok || (algorithm != "sha256" && algorithm != "md5") {
		return errors.New("checksum must use sha256 or md5")
	}
	actual, err := fileDigest(filename, algorithm)
	if err != nil {
		return err
	}
	if !strings.EqualFold(actual, expected) {
		return fmt.Errorf("checksum mismatch: got %s:%s, want %s", algorithm, actual, checksum)
	}
	return nil
}

func fileDigest(filename, algorithm string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()
	if algorithm == "md5" {
		hash := md5.New() // #nosec G401 -- Zenodo's draft API supplies MD5 integrity receipts.
		if _, err := io.Copy(hash, file); err != nil {
			return "", err
		}
		return hex.EncodeToString(hash.Sum(nil)), nil
	}
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
