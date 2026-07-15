// Command crossprovidervalidation proves an OSF-qualified snapshot can be
// copied to and compensated from an unpublished Zenodo sandbox draft.
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/edithatogo/osf-cli-go/internal/crossprovider"
	"github.com/edithatogo/osf-cli-go/internal/repository"
)

const (
	defaultBaseURL      = "https://sandbox.zenodo.org/api/"
	defaultEvidencePath = "docs/cross-provider-sandbox-validation-evidence.md"
	liveOptIn           = "CROSS_PROVIDER_SANDBOX_VALIDATION"
)

type evidence struct {
	Level              string
	Date               string
	SandboxHost        string
	SourceIdentity     string
	DestinationDraft   string
	IdempotencyKey     string
	FileSHA256         string
	FileBytes          int64
	ExecutionStatus    string
	Published          bool
	CompensationStatus string
	DraftDeleted       bool
	DeclaredScopes     string
}

func main() {
	live := flag.Bool("live", false, "run disposable cross-provider sandbox validation")
	evidencePath := flag.String("evidence", defaultEvidencePath, "write sanitized Markdown evidence")
	flag.Parse()
	result, err := run(context.Background(), *live)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := os.WriteFile(*evidencePath, []byte(result.markdown()), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write cross-provider evidence: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Cross-provider validation: %s (%s)\n", result.Level, *evidencePath)
}

func run(ctx context.Context, live bool) (result evidence, runErr error) {
	now := time.Now().UTC()
	result = evidence{
		Level: "integration-ready", Date: now.Format("2006-01-02"), SandboxHost: "sandbox.zenodo.org",
		SourceIdentity: "osf:project:synthetic-sandbox-validation", ExecutionStatus: "planned",
		CompensationStatus: "not-started", DeclaredScopes: "deposit:write",
	}
	directory, err := os.MkdirTemp("", "osf-cli-go-cross-provider-validation-")
	if err != nil {
		return result, err
	}
	defer func() { runErr = errors.Join(runErr, os.RemoveAll(directory)) }()
	content := []byte("osf-cli-go cross-provider sandbox validation\n")
	digest := sha256.Sum256(content)
	result.FileSHA256 = "sha256:" + hex.EncodeToString(digest[:])
	result.FileBytes = int64(len(content))
	if err := os.WriteFile(filepath.Join(directory, "validation.txt"), content, 0o600); err != nil {
		return result, err
	}
	native, err := repository.NewNativeMetadata("application/json", []byte(`{"id":"synthetic-sandbox-validation","provider":"osf","purpose":"cross-provider-validation"}`))
	if err != nil {
		return result, err
	}
	request := crossprovider.Request{
		Direction: crossprovider.DirectionOSFToZenodo,
		Source: crossprovider.Snapshot{
			Identity: repository.QualifiedID{Provider: repository.ProviderOSF, Kind: repository.KindProject, NativeID: "synthetic-sandbox-validation"},
			Metadata: crossprovider.Metadata{
				Title: "OSF CLI Go cross-provider sandbox validation", Description: "Disposable validation of draft-only OSF-to-Zenodo transfer semantics.",
				UploadType: "dataset", Creators: []crossprovider.Creator{{Name: "osf-cli-go validation"}},
				Keywords: []string{"osf", "zenodo", "validation"}, Access: crossprovider.AccessPolicy{Kind: crossprovider.AccessPublic}, License: "cc-by-4.0",
			},
			Files:          []crossprovider.File{{Path: "validation.txt", Size: int64(len(content)), Checksum: result.FileSHA256, MediaType: "text/plain"}},
			NativeMetadata: native,
		},
		Destination: crossprovider.Destination{Provider: repository.ProviderZenodo, CreateNew: true},
		Authorized:  true, PublishIntent: crossprovider.PublishDraftOnly, Conflict: crossprovider.ConflictFail,
	}
	report, err := crossprovider.BuildMapping(request, now)
	if err != nil {
		return result, err
	}
	result.IdempotencyKey = report.IdempotencyKey
	checkpoint, err := crossprovider.NewCheckpoint(report)
	if err != nil {
		return result, err
	}
	if !live {
		return result, nil
	}
	if os.Getenv(liveOptIn) != "1" {
		return result, fmt.Errorf("live cross-provider validation requires %s=1", liveOptIn)
	}
	baseURL := strings.TrimSpace(os.Getenv("ZENODO_BASE_URL"))
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	if baseURL != defaultBaseURL {
		return result, errors.New("live cross-provider validation is restricted to https://sandbox.zenodo.org/api/")
	}
	token := strings.TrimSpace(os.Getenv("ZENODO_TOKEN"))
	if token == "" {
		return result, errors.New("live cross-provider validation requires a dedicated ZENODO_TOKEN")
	}
	destination, err := crossprovider.NewZenodoSandboxDestination(baseURL, token)
	if err != nil {
		return result, err
	}
	executed, err := crossprovider.Execute(ctx, report, checkpoint, crossprovider.LocalSource{Root: directory}, destination)
	result.DestinationDraft = executed.Checkpoint.DestinationRef
	if err != nil {
		if result.DestinationDraft != "" {
			cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
			cleanupErr := destination.Compensate(cleanupCtx, crossprovider.CompensationAction{
				Kind: crossprovider.CompensationDiscardDraft, DestinationRef: result.DestinationDraft,
			})
			cancel()
			result.DraftDeleted = cleanupErr == nil
			result.CompensationStatus = "emergency-cleanup"
			err = errors.Join(err, cleanupErr)
		}
		return result, err
	}
	result.ExecutionStatus = string(executed.Checkpoint.Status)
	result.Published = executed.Partial.Published
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
	compensated, err := crossprovider.Compensate(cleanupCtx, executed.Checkpoint, destination)
	cancel()
	result.CompensationStatus = string(compensated.Status)
	result.DraftDeleted = err == nil && compensated.Status == crossprovider.SagaCompensated
	if err != nil {
		return result, err
	}
	if executed.Checkpoint.Status != crossprovider.SagaCompleted || result.Published || !result.DraftDeleted {
		return result, errors.New("cross-provider sandbox proof did not complete unpublished execution and compensation")
	}
	result.Level = "live-validated"
	return result, nil
}

func (result evidence) markdown() string {
	return fmt.Sprintf(`# Cross-provider sandbox validation evidence

- Validation level: %s
- Date: %s
- Sandbox host: %s
- Source identity: %s
- Destination draft: %s
- Idempotency key: %s
- File: bytes=%d checksum=%s
- Execution status: %s
- Published: %t
- Compensation status: %s
- Draft deleted: %t
- Declared token scopes: %s

The harness is restricted to Zenodo Sandbox, performs draft-only execution, and
uses a destination type that does not implement the publication interface.
`, result.Level, result.Date, result.SandboxHost, result.SourceIdentity, emptyAsNotCreated(result.DestinationDraft), result.IdempotencyKey,
		result.FileBytes, result.FileSHA256, result.ExecutionStatus, result.Published, result.CompensationStatus, result.DraftDeleted, result.DeclaredScopes)
}

func emptyAsNotCreated(value string) string {
	if strings.TrimSpace(value) == "" {
		return "not-created"
	}
	return value
}
