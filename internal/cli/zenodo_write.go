package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/edithatogo/osf-cli-go/internal/download"
	"github.com/edithatogo/osf-cli-go/internal/output"
	"github.com/edithatogo/osf-cli-go/internal/zenodopublish"
	"github.com/edithatogo/osf-cli-go/internal/zenodotransfer"
	"github.com/spf13/cobra"
)

type zenodoTransferClient interface {
	CreateDraft(context.Context) (zenodotransfer.Draft, error)
	GetDraft(context.Context, string) (zenodotransfer.Draft, error)
	DeleteDraft(context.Context, string) error
	DeleteFile(context.Context, string, string) error
	ListDraftFiles(context.Context, string) ([]zenodotransfer.RemoteFile, error)
	UploadFile(context.Context, zenodotransfer.Draft, string, string, download.ConflictPolicy) (zenodotransfer.UploadResult, error)
}

type zenodoPublicationClient interface {
	Execute(context.Context, zenodopublish.Request, time.Time) (zenodopublish.Result, error)
	ApplyDraftMetadata(context.Context, string, zenodopublish.Metadata, time.Time) error
}

const productionZenodoBaseURL = "https://zenodo.org/api/"

var zenodoProduction bool

var newZenodoTransferClient = func() (zenodoTransferClient, error) {
	baseURL := os.Getenv("ZENODO_BASE_URL")
	if zenodoProduction {
		if baseURL != productionZenodoBaseURL {
			return nil, fmt.Errorf("production writes require ZENODO_BASE_URL=%q", productionZenodoBaseURL)
		}
		return zenodotransfer.New(baseURL, os.Getenv("ZENODO_TOKEN"), zenodotransfer.WithProductionWrites())
	}
	return zenodotransfer.New(baseURL, os.Getenv("ZENODO_TOKEN"))
}

var newZenodoPublicationClient = func() (zenodoPublicationClient, error) {
	baseURL := os.Getenv("ZENODO_BASE_URL")
	scopes := []zenodopublish.Scope{
		zenodopublish.ScopeDepositWrite,
		zenodopublish.ScopeDepositActions,
	}
	if zenodoProduction {
		if baseURL != productionZenodoBaseURL {
			return nil, fmt.Errorf("production writes require ZENODO_BASE_URL=%q", productionZenodoBaseURL)
		}
		return zenodopublish.New(baseURL, os.Getenv("ZENODO_TOKEN"), scopes, zenodopublish.WithProductionWrites())
	}
	return zenodopublish.New(baseURL, os.Getenv("ZENODO_TOKEN"), scopes)
}

func newZenodoDepositsCommand() *cobra.Command {
	command := &cobra.Command{Use: "deposits", Short: "Manage explicitly confirmed Zenodo deposits"}
	command.AddCommand(
		newZenodoDraftCreateCommand(),
		newZenodoDraftGetCommand(),
		newZenodoDraftMetadataCommand(),
		newZenodoLifecycleCommand("reserve-doi", "Inspect the automatically reserved sandbox DOI", "draft", "reserve_doi", false),
		newZenodoLifecycleCommand("new-version", "Create a new sandbox version draft", "published", "new_version", false),
		newZenodoLifecycleCommand("discard", "Discard an unpublished Zenodo sandbox draft", "draft", "discard", false),
	)
	command.PersistentFlags().BoolVar(&zenodoProduction, "production", false, "allow writes to https://zenodo.org/api/ with additional confirmations")
	return command
}

func requireProductionConfirmation(provided, expected string) error {
	if zenodoProduction && provided != expected {
		return fmt.Errorf("production write requires --confirm %q", expected)
	}
	return nil
}

func newZenodoDraftCreateCommand() *cobra.Command {
	var execute bool
	var confirmation string
	command := &cobra.Command{Use: "create", Short: "Create an empty Zenodo sandbox draft", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		if !execute {
			return errors.New("draft creation requires --execute")
		}
		if err := requireProductionConfirmation(confirmation, "zenodo:production:create-draft"); err != nil {
			return err
		}
		client, err := newZenodoTransferClient()
		if err != nil {
			return err
		}
		draft, err := client.CreateDraft(cmd.Context())
		if err != nil {
			return err
		}
		return writeZenodoValue(cmd, draft, []string{"ID", "BUCKET URL"}, []string{draft.ID, draft.BucketURL})
	}}
	command.Flags().BoolVar(&execute, "execute", false, "perform the sandbox write")
	command.Flags().StringVar(&confirmation, "confirm", "", "exact production-action confirmation")
	return command
}

func newZenodoDraftGetCommand() *cobra.Command {
	return &cobra.Command{Use: "get <draft-id>", Short: "Inspect a Zenodo sandbox draft", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newZenodoTransferClient()
		if err != nil {
			return err
		}
		draft, err := client.GetDraft(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		return writeZenodoValue(cmd, draft, []string{"ID", "BUCKET URL"}, []string{draft.ID, draft.BucketURL})
	}}
}

func newZenodoDraftMetadataCommand() *cobra.Command {
	var metadataPath, confirmation string
	var execute bool
	command := &cobra.Command{Use: "metadata <draft-id>", Short: "Apply validated metadata to a Zenodo sandbox draft", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		metadata, err := readZenodoMetadata(metadataPath, time.Now())
		if err != nil {
			return err
		}
		if !execute {
			return output.WriteJSON(cmd.OutOrStdout(), struct {
				RecordID string                 `json:"recordId"`
				Metadata zenodopublish.Metadata `json:"metadata"`
				Executed bool                   `json:"executed"`
			}{RecordID: strings.TrimSpace(args[0]), Metadata: metadata})
		}
		if err := requireProductionConfirmation(confirmation, "zenodo:production:metadata:"+strings.TrimSpace(args[0])); err != nil {
			return err
		}
		client, err := newZenodoPublicationClient()
		if err != nil {
			return err
		}
		if err := client.ApplyDraftMetadata(cmd.Context(), args[0], metadata, time.Now()); err != nil {
			return err
		}
		return output.WriteJSON(cmd.OutOrStdout(), map[string]any{"recordId": strings.TrimSpace(args[0]), "executed": true})
	}}
	command.Flags().StringVar(&metadataPath, "metadata", "", "path to a Zenodo metadata JSON file")
	command.Flags().BoolVar(&execute, "execute", false, "perform the sandbox write")
	command.Flags().StringVar(&confirmation, "confirm", "", "exact production-action confirmation")
	_ = command.MarkFlagRequired("metadata")
	return command
}

func newZenodoDraftUploadCommand() *cobra.Command {
	var remoteName, conflict, confirmation string
	var execute bool
	command := &cobra.Command{Use: "upload <draft-id> <local-path>", Short: "Upload a file to a Zenodo sandbox draft", Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
		if !execute {
			return errors.New("draft file upload requires --execute")
		}
		policy := download.ConflictPolicy(conflict)
		if err := policy.Validate(); err != nil {
			return err
		}
		if remoteName == "" {
			remoteName = filepath.Base(args[1])
		}
		if err := requireProductionConfirmation(confirmation, "zenodo:production:upload:"+strings.TrimSpace(args[0])+":"+remoteName); err != nil {
			return err
		}
		client, err := newZenodoTransferClient()
		if err != nil {
			return err
		}
		draft, err := client.GetDraft(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		result, err := client.UploadFile(cmd.Context(), draft, args[1], remoteName, policy)
		if err != nil {
			return err
		}
		return output.WriteJSON(cmd.OutOrStdout(), result)
	}}
	command.Flags().StringVar(&remoteName, "name", "", "remote filename (defaults to the local basename)")
	command.Flags().StringVar(&conflict, "conflict", string(download.ConflictFail), "conflict policy: fail, skip, or overwrite")
	command.Flags().BoolVar(&execute, "execute", false, "perform the sandbox write")
	command.Flags().StringVar(&confirmation, "confirm", "", "exact production-action confirmation")
	return command
}

func newZenodoDraftFilesListCommand() *cobra.Command {
	return &cobra.Command{Use: "draft-list <draft-id>", Short: "List files in a Zenodo sandbox draft", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newZenodoTransferClient()
		if err != nil {
			return err
		}
		files, err := client.ListDraftFiles(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		mode, err := resolveOutputMode(cmd)
		if err != nil {
			return err
		}
		if mode == outputModeJSON {
			return output.WriteJSON(cmd.OutOrStdout(), files)
		}
		rows := make([][]string, 0, len(files))
		for _, file := range files {
			rows = append(rows, []string{file.ID, file.Name, fmt.Sprintf("%d", file.Size), file.Checksum})
		}
		return output.WriteTable(cmd.OutOrStdout(), []string{"ID", "NAME", "SIZE", "CHECKSUM"}, rows)
	}}
}

func newZenodoDraftFileDeleteCommand() *cobra.Command {
	var confirmation string
	command := &cobra.Command{Use: "delete <draft-id> <file-id>", Short: "Delete a file from a Zenodo sandbox draft", Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
		expected := fmt.Sprintf("zenodo:delete-file:%s:%s", strings.TrimSpace(args[0]), strings.TrimSpace(args[1]))
		if confirmation != expected {
			return fmt.Errorf("exact confirmation required: supply --confirm %q", expected)
		}
		client, err := newZenodoTransferClient()
		if err != nil {
			return err
		}
		if err := client.DeleteFile(cmd.Context(), args[0], args[1]); err != nil {
			return err
		}
		return output.WriteJSON(cmd.OutOrStdout(), map[string]any{"recordId": strings.TrimSpace(args[0]), "fileId": strings.TrimSpace(args[1]), "executed": true})
	}}
	command.Flags().StringVar(&confirmation, "confirm", "", "exact destructive-action confirmation")
	return command
}

func newZenodoLifecycleCommand(use, short, defaultState, action string, needsMetadata bool) *cobra.Command {
	var state, metadataPath, confirmation string
	var execute bool
	command := &cobra.Command{Use: use + " <record-id>", Short: short, Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		request := zenodopublish.Request{
			RecordID: strings.TrimSpace(args[0]), State: zenodopublish.State(state), Action: zenodopublish.Action(action),
			Authorized: true, DryRun: !execute, Confirmation: confirmation,
		}
		if needsMetadata {
			metadata, err := readZenodoMetadata(metadataPath, time.Now())
			if err != nil {
				return err
			}
			request.Metadata = metadata
		}
		if !execute {
			request.Scopes = []zenodopublish.Scope{zenodopublish.ScopeDepositWrite, zenodopublish.ScopeDepositActions}
			plan, err := zenodopublish.BuildPlan(request, time.Now())
			if err != nil {
				return err
			}
			return output.WriteJSON(cmd.OutOrStdout(), zenodopublish.Result{Plan: plan, RecordID: plan.RecordID})
		}
		client, err := newZenodoPublicationClient()
		if err != nil {
			return err
		}
		result, err := client.Execute(cmd.Context(), request, time.Now())
		if err != nil {
			return err
		}
		return output.WriteJSON(cmd.OutOrStdout(), result)
	}}
	command.Flags().StringVar(&state, "state", defaultState, "current lifecycle state")
	command.Flags().BoolVar(&execute, "execute", false, "perform the sandbox action instead of returning a dry-run plan")
	command.Flags().StringVar(&confirmation, "confirm", "", "exact confirmation emitted by the dry-run plan")
	if needsMetadata {
		command.Flags().StringVar(&metadataPath, "metadata", "", "path to a Zenodo metadata JSON file")
		_ = command.MarkFlagRequired("metadata")
	}
	return command
}

func readZenodoMetadata(path string, now time.Time) (zenodopublish.Metadata, error) {
	payload, err := os.ReadFile(path)
	if err != nil {
		return zenodopublish.Metadata{}, fmt.Errorf("read Zenodo metadata: %w", err)
	}
	var metadata zenodopublish.Metadata
	decoder := json.NewDecoder(strings.NewReader(string(payload)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&metadata); err != nil {
		return zenodopublish.Metadata{}, fmt.Errorf("decode Zenodo metadata: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return zenodopublish.Metadata{}, errors.New("decode Zenodo metadata: multiple JSON values are not allowed")
		}
		return zenodopublish.Metadata{}, fmt.Errorf("decode Zenodo metadata: %w", err)
	}
	if err := zenodopublish.ValidateMetadata(metadata, now); err != nil {
		return zenodopublish.Metadata{}, err
	}
	return metadata, nil
}

func writeZenodoValue(cmd *cobra.Command, value any, headers, row []string) error {
	mode, err := resolveOutputMode(cmd)
	if err != nil {
		return err
	}
	if mode == outputModeJSON {
		return output.WriteJSON(cmd.OutOrStdout(), value)
	}
	return output.WriteTable(cmd.OutOrStdout(), headers, [][]string{row})
}
