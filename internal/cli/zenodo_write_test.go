package cli

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/edithatogo/osf-cli-go/internal/download"
	"github.com/edithatogo/osf-cli-go/internal/zenodopublish"
	"github.com/edithatogo/osf-cli-go/internal/zenodotransfer"
)

type fakeZenodoTransfer struct {
	draft       zenodotransfer.Draft
	created     bool
	uploaded    bool
	deletedFile string
}

func (fake *fakeZenodoTransfer) CreateDraft(context.Context) (zenodotransfer.Draft, error) {
	fake.created = true
	return fake.draft, nil
}

func (fake *fakeZenodoTransfer) GetDraft(_ context.Context, id string) (zenodotransfer.Draft, error) {
	if id != fake.draft.ID {
		return zenodotransfer.Draft{}, errors.New("unknown draft")
	}
	return fake.draft, nil
}

func (*fakeZenodoTransfer) DeleteDraft(context.Context, string) error { return nil }

func (fake *fakeZenodoTransfer) DeleteFile(_ context.Context, draftID, fileID string) error {
	fake.deletedFile = draftID + "/" + fileID
	return nil
}

func (*fakeZenodoTransfer) ListDraftFiles(context.Context, string) ([]zenodotransfer.RemoteFile, error) {
	return nil, nil
}

func (fake *fakeZenodoTransfer) UploadFile(_ context.Context, _ zenodotransfer.Draft, source, name string, policy download.ConflictPolicy) (zenodotransfer.UploadResult, error) {
	if name == "" || source == "" || policy != download.ConflictFail {
		return zenodotransfer.UploadResult{}, errors.New("invalid upload")
	}
	fake.uploaded = true
	return zenodotransfer.UploadResult{Remote: zenodotransfer.RemoteFile{ID: "file-1", Name: name}, Completed: true}, nil
}

type fakeZenodoPublication struct {
	request  zenodopublish.Request
	metadata zenodopublish.Metadata
}

func (fake *fakeZenodoPublication) Execute(_ context.Context, request zenodopublish.Request, now time.Time) (zenodopublish.Result, error) {
	fake.request = request
	request.Scopes = []zenodopublish.Scope{zenodopublish.ScopeDepositWrite, zenodopublish.ScopeDepositActions}
	plan, err := zenodopublish.BuildPlan(request, now)
	return zenodopublish.Result{Plan: plan, RecordID: request.RecordID, Executed: err == nil}, err
}

func (fake *fakeZenodoPublication) ApplyDraftMetadata(_ context.Context, _ string, metadata zenodopublish.Metadata, _ time.Time) error {
	fake.metadata = metadata
	return nil
}

func TestZenodoWriteCommandsAreGuardedAndExecutable(t *testing.T) {
	transfer := &fakeZenodoTransfer{draft: zenodotransfer.Draft{ID: "123", BucketURL: "https://sandbox.zenodo.org/api/files/bucket"}}
	publication := &fakeZenodoPublication{}
	oldTransfer, oldPublication := newZenodoTransferClient, newZenodoPublicationClient
	newZenodoTransferClient = func(bool) (zenodoTransferClient, error) { return transfer, nil }
	newZenodoPublicationClient = func(bool) (zenodoPublicationClient, error) { return publication, nil }
	t.Cleanup(func() { newZenodoTransferClient, newZenodoPublicationClient = oldTransfer, oldPublication })

	if _, err := executeZenodoWriteCommand("zenodo", "deposits", "create"); err == nil || transfer.created {
		t.Fatal("unguarded draft creation succeeded")
	}
	if output, err := executeZenodoWriteCommand("zenodo", "deposits", "create", "--execute", "--json"); err != nil || !transfer.created || !strings.Contains(output, `"id":"123"`) {
		t.Fatalf("create output=%q created=%v err=%v", output, transfer.created, err)
	}

	source := filepath.Join(t.TempDir(), "data.csv")
	if err := os.WriteFile(source, []byte("a,b\n1,2\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if output, err := executeZenodoWriteCommand("zenodo", "files", "upload", "123", source, "--execute"); err != nil || !transfer.uploaded || !strings.Contains(output, `"completed":true`) {
		t.Fatalf("upload output=%q uploaded=%v err=%v", output, transfer.uploaded, err)
	}

	if _, err := executeZenodoWriteCommand("zenodo", "files", "delete", "123", "file-1"); err == nil || transfer.deletedFile != "" {
		t.Fatal("unconfirmed file deletion succeeded")
	}
	if _, err := executeZenodoWriteCommand("zenodo", "files", "delete", "123", "file-1", "--confirm", "zenodo:delete-file:123:file-1"); err != nil || transfer.deletedFile != "123/file-1" {
		t.Fatalf("delete=%q err=%v", transfer.deletedFile, err)
	}
}

func TestZenodoProductionWritesRequireTargetAndConfirmations(t *testing.T) {
	transfer := &fakeZenodoTransfer{draft: zenodotransfer.Draft{ID: "123", BucketURL: "https://zenodo.org/api/files/bucket"}}
	oldTransfer := newZenodoTransferClient
	productionRequested := false
	newZenodoTransferClient = func(production bool) (zenodoTransferClient, error) {
		productionRequested = production
		return transfer, nil
	}
	t.Cleanup(func() { newZenodoTransferClient = oldTransfer })

	if _, err := executeZenodoWriteCommand("zenodo", "deposits", "create", "--production", "--execute"); err == nil || transfer.created {
		t.Fatal("production draft creation succeeded without its confirmation")
	}
	if _, err := executeZenodoWriteCommand("zenodo", "deposits", "create", "--production", "--execute", "--confirm", "zenodo:production:create-draft"); err != nil || !transfer.created {
		t.Fatalf("production draft creation err=%v created=%v", err, transfer.created)
	}
	if !productionRequested {
		t.Fatal("production draft creation did not select the production client")
	}

	productionRequested = false
	if _, err := executeZenodoWriteCommand("zenodo", "deposits", "get", "123", "--production"); err != nil || !productionRequested {
		t.Fatalf("production draft get err=%v production=%v", err, productionRequested)
	}

	productionRequested = false
	if _, err := executeZenodoWriteCommand("zenodo", "files", "draft-list", "123", "--production"); err != nil || !productionRequested {
		t.Fatalf("production draft file list err=%v production=%v", err, productionRequested)
	}
}

func TestZenodoMetadataAndPublishDryRun(t *testing.T) {
	metadataPath := filepath.Join(t.TempDir(), "metadata.json")
	metadata := `{"title":"Dataset","description":"Description","uploadType":"dataset","creators":[{"name":"Doe, Jane"}],"access":"open","license":"cc-by-4.0"}`
	if err := os.WriteFile(metadataPath, []byte(metadata), 0o600); err != nil {
		t.Fatal(err)
	}

	output, err := executeZenodoWriteCommand("zenodo", "publish", "123", "--metadata", metadataPath)
	if err != nil || !strings.Contains(output, `"confirmation":"zenodo:publish:123:published"`) || !strings.Contains(output, `"dryRun":true`) {
		t.Fatalf("dry-run output=%q err=%v", output, err)
	}

	publication := &fakeZenodoPublication{}
	oldPublication := newZenodoPublicationClient
	newZenodoPublicationClient = func(bool) (zenodoPublicationClient, error) { return publication, nil }
	t.Cleanup(func() { newZenodoPublicationClient = oldPublication })
	output, err = executeZenodoWriteCommand("zenodo", "publish", "123", "--metadata", metadataPath, "--execute", "--confirm", "zenodo:publish:123:published")
	if err != nil || !publication.request.Authorized || publication.request.DryRun || !strings.Contains(output, `"executed":true`) {
		t.Fatalf("execute output=%q request=%+v err=%v", output, publication.request, err)
	}
}

func TestZenodoMetadataRejectsUnknownAndTrailingJSON(t *testing.T) {
	now := time.Date(2026, 7, 23, 0, 0, 0, 0, time.UTC)
	valid := `{"title":"Dataset","description":"Description","uploadType":"dataset","creators":[{"name":"Doe, Jane"}],"access":"open","license":"cc-by-4.0"}`
	for name, payload := range map[string]string{
		"unknown field":  strings.TrimSuffix(valid, "}") + `,"unexpected":true}`,
		"trailing value": valid + ` {}`,
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "metadata.json")
			if err := os.WriteFile(path, []byte(payload), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := readZenodoMetadata(path, now); err == nil {
				t.Fatalf("metadata accepted: %s", payload)
			}
		})
	}
}

func executeZenodoWriteCommand(args ...string) (string, error) {
	var stdout, stderr bytes.Buffer
	root := newRootCommandWithProviders(&stdout, &stderr, &fakeReadonlyClient{}, &fakeZenodoREST{}, &fakeOAIClient{})
	root.SetArgs(args)
	err := root.Execute()
	return stdout.String(), err
}
