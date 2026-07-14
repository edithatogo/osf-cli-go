package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/edithatogo/osf-cli-go/internal/observability"
	"github.com/edithatogo/osf-cli-go/internal/output"
	"github.com/edithatogo/osf-cli-go/internal/repository"
	"github.com/edithatogo/osf-cli-go/internal/zenodoapi"
	"github.com/spf13/cobra"
)

type zenodoRESTClient interface {
	SearchRecords(context.Context, string, int) ([]zenodoapi.Record, error)
	GetRecord(context.Context, string) (zenodoapi.Record, error)
	ListRecordFiles(context.Context, string) ([]zenodoapi.File, error)
}

func newDefaultZenodoRESTClient(emitter observability.Emitter) zenodoRESTClient {
	client, err := zenodoapi.New("", zenodoapi.WithObserver(emitter))
	if err != nil {
		panic(err)
	}
	return client
}

type zenodoRecordOutput struct {
	QualifiedID    string                  `json:"qualifiedId"`
	Provider       repository.Provider     `json:"provider"`
	Kind           repository.ResourceKind `json:"kind"`
	ID             string                  `json:"id"`
	ConceptID      string                  `json:"conceptId,omitempty"`
	Title          string                  `json:"title"`
	Description    string                  `json:"description,omitempty"`
	DOI            string                  `json:"doi,omitempty"`
	ConceptDOI     string                  `json:"conceptDoi,omitempty"`
	Creators       []zenodoapi.Creator     `json:"creators,omitempty"`
	Keywords       []string                `json:"keywords,omitempty"`
	AccessRight    string                  `json:"accessRight,omitempty"`
	License        string                  `json:"license,omitempty"`
	Created        string                  `json:"created,omitempty"`
	Updated        string                  `json:"updated,omitempty"`
	Links          map[string]string       `json:"links,omitempty"`
	NativeMetadata json.RawMessage         `json:"nativeMetadata,omitempty"`
}

type zenodoFileOutput struct {
	RecordQualifiedID string            `json:"recordQualifiedId"`
	ID                string            `json:"id,omitempty"`
	Key               string            `json:"key"`
	Size              int64             `json:"size,omitempty"`
	Checksum          string            `json:"checksum,omitempty"`
	DownloadURL       string            `json:"downloadUrl,omitempty"`
	Links             map[string]string `json:"links,omitempty"`
}

func newZenodoRecordsCommand(client zenodoRESTClient) *cobra.Command {
	command := &cobra.Command{Use: "records", Short: "Search and inspect published Zenodo records"}
	command.AddCommand(newZenodoRecordsSearchCommand(client), newZenodoRecordsGetCommand(client))
	command.AddCommand(newUnsupportedZenodoCommand("create", repository.CapabilityRecordCreate), newUnsupportedZenodoCommand("update <record>", repository.CapabilityRecordUpdate), newUnsupportedZenodoCommand("delete <record>", repository.CapabilityRecordDelete))
	return command
}

func newZenodoRecordsSearchCommand(client zenodoRESTClient) *cobra.Command {
	var limit int
	command := &cobra.Command{Use: "search [query]", Short: "Search public Zenodo records", Args: cobra.MaximumNArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		mode, err := resolveOutputMode(cmd)
		if err != nil {
			return err
		}
		if limit < 1 || limit > 100 {
			return errors.New("--limit must be between 1 and 100")
		}
		query := ""
		if len(args) == 1 {
			query = strings.TrimSpace(args[0])
		}
		records, err := client.SearchRecords(cmd.Context(), query, limit)
		if err != nil {
			return err
		}
		rows, err := zenodoRecordOutputs(records)
		if err != nil {
			return err
		}
		if mode == outputModeJSON {
			return output.WriteJSON(cmd.OutOrStdout(), rows)
		}
		table := make([][]string, 0, len(rows))
		for _, row := range rows {
			table = append(table, []string{row.QualifiedID, row.Title, row.DOI, row.AccessRight})
		}
		return output.WriteTable(cmd.OutOrStdout(), []string{"QUALIFIED ID", "TITLE", "DOI", "ACCESS"}, table)
	}}
	command.Flags().IntVar(&limit, "limit", 10, "maximum records (1-100)")
	return command
}

func newZenodoRecordsGetCommand(client zenodoRESTClient) *cobra.Command {
	return &cobra.Command{Use: "get <id-or-url>", Short: "Inspect a published Zenodo record", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		mode, err := resolveOutputMode(cmd)
		if err != nil {
			return err
		}
		id, err := parseZenodoRecordID(args[0])
		if err != nil {
			return err
		}
		record, err := client.GetRecord(cmd.Context(), id)
		if err != nil {
			return err
		}
		row, err := zenodoRecordOutputFor(record)
		if err != nil {
			return err
		}
		if mode == outputModeJSON {
			return output.WriteJSON(cmd.OutOrStdout(), row)
		}
		return output.WriteTable(cmd.OutOrStdout(), []string{"FIELD", "VALUE"}, [][]string{{"Qualified ID", row.QualifiedID}, {"Title", row.Title}, {"DOI", row.DOI}, {"Concept DOI", row.ConceptDOI}, {"Access", row.AccessRight}, {"License", row.License}, {"Updated", row.Updated}})
	}}
}

func newZenodoFilesCommand(client zenodoRESTClient) *cobra.Command {
	command := &cobra.Command{Use: "files", Short: "Inspect files attached to published Zenodo records"}
	command.AddCommand(&cobra.Command{Use: "list <record-id-or-url>", Short: "List files in a published Zenodo record", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		mode, err := resolveOutputMode(cmd)
		if err != nil {
			return err
		}
		id, err := parseZenodoRecordID(args[0])
		if err != nil {
			return err
		}
		files, err := client.ListRecordFiles(cmd.Context(), id)
		if err != nil {
			return err
		}
		qualified, err := (repository.QualifiedID{Provider: repository.ProviderZenodo, Kind: repository.KindRecord, NativeID: id}).Key()
		if err != nil {
			return err
		}
		rows := make([]zenodoFileOutput, 0, len(files))
		for _, file := range files {
			rows = append(rows, zenodoFileOutput{RecordQualifiedID: qualified, ID: file.ID, Key: file.Key, Size: file.Size, Checksum: file.Checksum, DownloadURL: file.ContentURL(), Links: file.Links})
		}
		if mode == outputModeJSON {
			return output.WriteJSON(cmd.OutOrStdout(), rows)
		}
		table := make([][]string, 0, len(rows))
		for _, row := range rows {
			table = append(table, []string{row.Key, strconv.FormatInt(row.Size, 10), row.Checksum, row.DownloadURL})
		}
		return output.WriteTable(cmd.OutOrStdout(), []string{"KEY", "SIZE", "CHECKSUM", "DOWNLOAD URL"}, table)
	}})
	command.AddCommand(newUnsupportedZenodoCommand("upload <record>", repository.CapabilityFileUpload), newUnsupportedZenodoCommand("delete <record>", repository.CapabilityFileDelete))
	return command
}

func newZenodoCapabilitiesCommand() *cobra.Command {
	return &cobra.Command{Use: "capabilities", Short: "Show the reviewed Zenodo provider capability contract", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		contract := repository.ZenodoContract()
		if err := contract.Validate(); err != nil {
			return err
		}
		mode, err := resolveOutputMode(cmd)
		if err != nil {
			return err
		}
		if mode == outputModeJSON {
			return output.WriteJSON(cmd.OutOrStdout(), contract)
		}
		rows := make([][]string, 0, len(contract.Capabilities))
		for _, detail := range contract.Capabilities {
			rows = append(rows, []string{string(detail.Capability), string(detail.Level), strings.Join(detail.Constraints, "; ")})
		}
		return output.WriteTable(cmd.OutOrStdout(), []string{"CAPABILITY", "SUPPORT", "CONSTRAINTS"}, rows)
	}}
}

func newZenodoPublishCommand() *cobra.Command {
	return newUnsupportedZenodoCommand("publish <record>", repository.CapabilityPublish)
}

func newUnsupportedZenodoCommand(use string, capability repository.Capability) *cobra.Command {
	return &cobra.Command{Use: use, Short: "Report the reviewed availability of " + string(capability), Args: cobra.ArbitraryArgs, RunE: func(*cobra.Command, []string) error { return repository.ZenodoContract().Require(capability) }}
}

func parseZenodoRecordID(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", errors.New("zenodo record id is required")
	}
	if strings.HasPrefix(value, "zenodo:") {
		id, err := repository.ParseQualifiedID(value)
		if err != nil {
			return "", fmt.Errorf("invalid Zenodo qualified id: %w", err)
		}
		if id.Provider != repository.ProviderZenodo || id.Kind != repository.KindRecord {
			return "", errors.New("qualified id must identify a Zenodo record")
		}
		value = id.NativeID
	} else if parsed, err := url.Parse(value); err == nil && parsed.Scheme != "" {
		host := strings.ToLower(parsed.Hostname())
		if parsed.Scheme != "https" || parsed.RawQuery != "" || parsed.Fragment != "" || (host != "zenodo.org" && host != "sandbox.zenodo.org") {
			return "", errors.New("zenodo record URL must use https://zenodo.org or https://sandbox.zenodo.org")
		}
		parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
		if len(parts) != 2 || parts[0] != "records" || strings.TrimSpace(parts[1]) == "" {
			return "", errors.New("zenodo record URL must have /records/<id> form")
		}
		value = parts[1]
	}
	if strings.ContainsAny(value, "/?# \\") {
		return "", errors.New("zenodo record id must not contain paths, spaces, query, or fragment")
	}
	for _, character := range value {
		if character < '0' || character > '9' {
			return "", errors.New("zenodo record id must contain decimal digits only")
		}
	}
	return value, nil
}

func zenodoRecordOutputs(records []zenodoapi.Record) ([]zenodoRecordOutput, error) {
	rows := make([]zenodoRecordOutput, 0, len(records))
	for _, record := range records {
		row, err := zenodoRecordOutputFor(record)
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func zenodoRecordOutputFor(record zenodoapi.Record) (zenodoRecordOutput, error) {
	qualified, err := (repository.QualifiedID{Provider: repository.ProviderZenodo, Kind: repository.KindRecord, NativeID: record.ID}).Key()
	if err != nil {
		return zenodoRecordOutput{}, err
	}
	return zenodoRecordOutput{QualifiedID: qualified, Provider: repository.ProviderZenodo, Kind: repository.KindRecord, ID: record.ID, ConceptID: record.ConceptRecID, Title: record.Metadata.Title, Description: record.Metadata.Description, DOI: record.DOI, ConceptDOI: record.ConceptDOI, Creators: append([]zenodoapi.Creator(nil), record.Metadata.Creators...), Keywords: append([]string(nil), record.Metadata.Keywords...), AccessRight: record.Metadata.AccessRight, License: record.Metadata.License.ID, Created: record.Created, Updated: record.Updated, Links: record.Links, NativeMetadata: json.RawMessage(record.NativeJSON())}, nil
}
