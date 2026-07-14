// Command zenodopublicationvalidation records an offline plan or runs the
// explicitly opted-in Zenodo sandbox publication lifecycle proof.
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/edithatogo/osf-cli-go/internal/download"
	"github.com/edithatogo/osf-cli-go/internal/zenodopublish"
	"github.com/edithatogo/osf-cli-go/internal/zenodotransfer"
)

const (
	defaultBaseURL      = "https://sandbox.zenodo.org/api/"
	liveOptIn           = "ZENODO_PUBLICATION_VALIDATION"
	defaultEvidencePath = "docs/zenodo-publication-validation-evidence.md"
)

type evidence struct {
	ValidationLevel  string
	Date             string
	SandboxHost      string
	RecordID         string
	VersionDraftID   string
	ReservedDOI      string
	PublishedDOI     string
	FileSHA256       string
	FileBytes        int64
	AuditOutcomes    []string
	VersionDiscarded bool
	DraftCleanup     string
}

func main() {
	live := flag.Bool("live", false, "run the irreversible Zenodo sandbox lifecycle proof")
	evidencePath := flag.String("evidence", defaultEvidencePath, "write redacted Markdown evidence")
	flag.Parse()

	result, err := run(context.Background(), *live)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := os.WriteFile(*evidencePath, []byte(result.markdown()), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write Zenodo publication evidence: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Zenodo publication validation: %s (%s)\n", result.ValidationLevel, *evidencePath)
}

func run(ctx context.Context, live bool) (result evidence, runErr error) {
	now := time.Now()
	result = evidence{ValidationLevel: "integration-ready", Date: now.Format("2006-01-02"), SandboxHost: "sandbox.zenodo.org", DraftCleanup: "not-created"}
	metadata := validationMetadata(now)
	preview, err := zenodopublish.BuildPlan(zenodopublish.Request{
		RecordID: "offline-draft", State: zenodopublish.StateDraft, Action: zenodopublish.ActionPublish,
		Authorized: true, DryRun: true, Scopes: []zenodopublish.Scope{zenodopublish.ScopeDepositWrite, zenodopublish.ScopeDepositActions}, Metadata: metadata,
	}, now)
	if err != nil || preview.Confirmation == "" {
		return result, fmt.Errorf("build offline Zenodo publication plan: %w", err)
	}
	if !live {
		return result, nil
	}
	if os.Getenv(liveOptIn) != "1" {
		return result, fmt.Errorf("live Zenodo publication validation requires %s=1", liveOptIn)
	}
	baseURL := strings.TrimSpace(os.Getenv("ZENODO_BASE_URL"))
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	if baseURL != defaultBaseURL {
		return result, errors.New("live Zenodo publication validation is restricted to https://sandbox.zenodo.org/api/")
	}
	token := strings.TrimSpace(os.Getenv("ZENODO_TOKEN"))
	if token == "" {
		return result, errors.New("live Zenodo publication validation requires a dedicated ZENODO_TOKEN")
	}

	transfer, err := zenodotransfer.New(baseURL, token)
	if err != nil {
		return result, err
	}
	var audits []zenodopublish.AuditEvent
	publication, err := zenodopublish.New(baseURL, token, []zenodopublish.Scope{zenodopublish.ScopeDepositWrite, zenodopublish.ScopeDepositActions}, zenodopublish.WithAuditSink(func(event zenodopublish.AuditEvent) {
		audits = append(audits, event)
	}))
	if err != nil {
		return result, err
	}
	draft, err := transfer.CreateDraft(ctx)
	if err != nil {
		return result, fmt.Errorf("create Zenodo publication validation draft: %w", err)
	}
	result.RecordID = draft.ID
	result.DraftCleanup = "pending"
	published := false
	versionDraftID := ""
	defer func() {
		if !published {
			if err := transfer.DeleteDraft(context.WithoutCancel(ctx), draft.ID); err != nil {
				result.DraftCleanup = "failed"
				runErr = errors.Join(runErr, fmt.Errorf("cleanup unpublished Zenodo validation draft: %w", err))
			} else {
				result.DraftCleanup = "deleted"
			}
			return
		}
		if versionDraftID != "" && !result.VersionDiscarded {
			if err := discardVersionDraft(context.WithoutCancel(ctx), publication, versionDraftID, now); err != nil {
				runErr = errors.Join(runErr, fmt.Errorf("cleanup unpublished Zenodo version draft: %w", err))
			} else {
				result.VersionDiscarded = true
			}
		}
	}()

	directory, err := os.MkdirTemp("", "osf-cli-go-zenodo-publication-")
	if err != nil {
		return result, fmt.Errorf("create validation workspace: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(directory); err != nil {
			runErr = errors.Join(runErr, fmt.Errorf("remove Zenodo validation workspace: %w", err))
		}
	}()
	content := []byte("osf-cli-go Zenodo sandbox publication validation\n")
	digest := sha256.Sum256(content)
	result.FileSHA256 = "sha256:" + hex.EncodeToString(digest[:])
	result.FileBytes = int64(len(content))
	source := filepath.Join(directory, "validation.txt")
	if err := os.WriteFile(source, content, 0o600); err != nil {
		return result, fmt.Errorf("write validation file: %w", err)
	}
	if _, err := transfer.UploadFile(ctx, draft, source, "validation.txt", download.ConflictFail); err != nil {
		return result, fmt.Errorf("upload Zenodo publication validation file: %w", err)
	}

	reserved, err := publication.Execute(ctx, zenodopublish.Request{
		RecordID: draft.ID, State: zenodopublish.StateDraft, Action: zenodopublish.ActionReserveDOI,
		Authorized: true, Metadata: metadata,
	}, now)
	if err != nil {
		return result, fmt.Errorf("reserve sandbox DOI: %w", err)
	}
	result.ReservedDOI = reserved.DOI
	publishRequest := zenodopublish.Request{
		RecordID: draft.ID, State: zenodopublish.StateDOIReserved, Action: zenodopublish.ActionPublish,
		Authorized: true, DryRun: true, Metadata: metadata,
	}
	publishPreview, err := publication.Execute(ctx, publishRequest, now)
	if err != nil {
		return result, fmt.Errorf("preview sandbox publication: %w", err)
	}
	publishRequest.DryRun = false
	publishRequest.Confirmation = publishPreview.Plan.Confirmation
	publishedResult, err := publication.Execute(ctx, publishRequest, now)
	if err != nil {
		return result, fmt.Errorf("publish sandbox record: %w", err)
	}
	published = true
	result.DraftCleanup = "not-applicable-published"
	result.PublishedDOI = publishedResult.DOI

	version, err := publication.Execute(ctx, zenodopublish.Request{
		RecordID: draft.ID, State: zenodopublish.StatePublished, Action: zenodopublish.ActionNewVersion,
		Authorized: true,
	}, now)
	if err != nil {
		return result, fmt.Errorf("create sandbox version draft: %w", err)
	}
	result.VersionDraftID = version.RecordID
	versionDraftID = version.RecordID
	if err := discardVersionDraft(ctx, publication, version.RecordID, now); err != nil {
		return result, fmt.Errorf("discard sandbox version draft: %w", err)
	}
	result.VersionDiscarded = true
	result.ValidationLevel = "live-validated"
	for _, audit := range audits {
		result.AuditOutcomes = append(result.AuditOutcomes, string(audit.Action)+":"+audit.Outcome)
	}
	return result, nil
}

func discardVersionDraft(ctx context.Context, client *zenodopublish.Client, recordID string, now time.Time) error {
	request := zenodopublish.Request{
		RecordID: recordID, State: zenodopublish.StateVersionDraft, Action: zenodopublish.ActionDiscard,
		Authorized: true, DryRun: true,
	}
	preview, err := client.Execute(ctx, request, now)
	if err != nil {
		return err
	}
	request.DryRun = false
	request.Confirmation = preview.Plan.Confirmation
	_, err = client.Execute(ctx, request, now)
	return err
}

func validationMetadata(now time.Time) zenodopublish.Metadata {
	return zenodopublish.Metadata{
		Title:       "OSF CLI Go sandbox lifecycle validation " + now.Format("20060102-150405"),
		Description: "Automated validation of explicit DOI reservation, publication, version creation, and draft discard safety gates.",
		UploadType:  "dataset", Creators: []zenodopublish.Creator{{Name: "osf-cli-go validation"}},
		Access: zenodopublish.AccessOpen, License: "cc-by-4.0",
	}
}

func (result evidence) markdown() string {
	audits, _ := json.Marshal(result.AuditOutcomes)
	return fmt.Sprintf(`# Zenodo publication validation evidence

- Date: %s
- Validation level: %s
- Host: %s
- Published sandbox record ID: %s
- Reserved DOI: %s
- Published DOI: %s
- Validation file: %d bytes, %s
- New-version draft ID: %s
- New-version draft discarded: %t
- Original draft cleanup: %s
- Redacted audit outcomes: %s
- Token scope policy: deposit:write and deposit:actions; user:email excluded

The live proof is opt-in and sandbox-only. Publication is irreversible: the
published sandbox record remains available and is never described as disposable
or deleted. Only the unpublished new-version draft is discarded. Evidence omits
the token, confirmation challenges, metadata description, and authenticated
response bodies.
`, result.Date, result.ValidationLevel, result.SandboxHost, result.RecordID,
		result.ReservedDOI, result.PublishedDOI, result.FileBytes, result.FileSHA256,
		result.VersionDraftID, result.VersionDiscarded, result.DraftCleanup, audits)
}
