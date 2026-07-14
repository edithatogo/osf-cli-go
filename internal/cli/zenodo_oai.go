package cli

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/edithatogo/osf-cli-go/internal/observability"
	"github.com/edithatogo/osf-cli-go/internal/output"
	"github.com/edithatogo/osf-cli-go/internal/zenodooai"
	"github.com/spf13/cobra"
)

type zenodoOAIClient interface {
	ListRecords(context.Context, zenodooai.Request) (zenodooai.Page, error)
	Harvest(context.Context, zenodooai.Request) ([]zenodooai.Record, error)
	ListSets(context.Context) ([]zenodooai.Set, error)
	ListMetadataFormats(context.Context, string) ([]zenodooai.MetadataFormat, error)
}

func newDefaultZenodoOAIClient(emitter observability.Emitter) zenodoOAIClient {
	client, err := zenodooai.New("", zenodooai.WithObserver(emitter))
	if err != nil {
		panic(err)
	}
	return client
}

func newZenodoCommand(rest zenodoRESTClient, client zenodoOAIClient) *cobra.Command {
	command := &cobra.Command{Use: "zenodo", Short: "Use provider-specific Zenodo workflows"}
	oai := &cobra.Command{Use: "oai", Short: "Harvest public Zenodo OAI-PMH metadata"}
	oai.AddCommand(newZenodoOAIHarvestCommand(client), newZenodoOAISetsCommand(client), newZenodoOAIFormatsCommand(client))
	command.AddCommand(newZenodoRecordsCommand(rest), newZenodoFilesCommand(rest), newZenodoCapabilitiesCommand(), newZenodoPublishCommand(), oai)
	return command
}

func newZenodoOAIHarvestCommand(client zenodoOAIClient) *cobra.Command {
	var prefix, setName, fromValue, untilValue, resumeToken string
	var all bool
	command := &cobra.Command{
		Use:   "harvest",
		Short: "Harvest one page or all bounded pages of OAI-PMH records",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			request, err := oaiRequest(prefix, setName, fromValue, untilValue, resumeToken)
			if err != nil {
				return err
			}
			mode, err := resolveOutputMode(cmd)
			if err != nil {
				return err
			}
			if all {
				records, err := client.Harvest(cmd.Context(), request)
				if err != nil {
					return err
				}
				return writeOAIRecords(cmd, mode, records, zenodooai.ResumptionToken{})
			}
			page, err := client.ListRecords(cmd.Context(), request)
			if err != nil {
				return err
			}
			return writeOAIRecords(cmd, mode, page.Records, page.Next)
		},
	}
	command.Flags().StringVar(&prefix, "metadata-prefix", "oai_dc", "OAI metadata prefix")
	command.Flags().StringVar(&setName, "set", "", "optional OAI set spec")
	command.Flags().StringVar(&fromValue, "from", "", "inclusive RFC3339 or YYYY-MM-DD start")
	command.Flags().StringVar(&untilValue, "until", "", "inclusive RFC3339 or YYYY-MM-DD end")
	command.Flags().StringVar(&resumeToken, "resume-token", "", "opaque token returned by a prior page")
	command.Flags().BoolVar(&all, "all", false, "follow bounded resumption tokens until complete")
	return command
}

func oaiRequest(prefix, setName, fromValue, untilValue, token string) (zenodooai.Request, error) {
	if token = strings.TrimSpace(token); token != "" {
		if strings.TrimSpace(setName) != "" || strings.TrimSpace(fromValue) != "" || strings.TrimSpace(untilValue) != "" {
			return zenodooai.Request{}, errors.New("--resume-token cannot be combined with --set, --from, or --until")
		}
		return zenodooai.Request{Token: zenodooai.ResumptionToken{Value: token, MetadataPrefix: strings.TrimSpace(prefix)}}, nil
	}
	from, err := parseOAIDate(fromValue)
	if err != nil {
		return zenodooai.Request{}, fmt.Errorf("invalid --from: %w", err)
	}
	until, err := parseOAIDate(untilValue)
	if err != nil {
		return zenodooai.Request{}, fmt.Errorf("invalid --until: %w", err)
	}
	return zenodooai.Request{MetadataPrefix: strings.TrimSpace(prefix), Set: strings.TrimSpace(setName), From: from, Until: until}, nil
}

func parseOAIDate(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}
	for _, layout := range []string{time.RFC3339, time.DateOnly} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, errors.New("expected RFC3339 or YYYY-MM-DD")
}

type oaiCLIResult struct {
	Records []zenodooai.Record         `json:"records"`
	Next    *zenodooai.ResumptionToken `json:"next,omitempty"`
}

func writeOAIRecords(cmd *cobra.Command, mode string, records []zenodooai.Record, next zenodooai.ResumptionToken) error {
	if mode == outputModeJSON {
		result := oaiCLIResult{Records: records}
		if !next.Empty() {
			result.Next = &next
		}
		return output.WriteJSON(cmd.OutOrStdout(), result)
	}
	rows := make([][]string, 0, len(records))
	for _, record := range records {
		rows = append(rows, []string{record.Header.Identifier, record.Header.Datestamp, strconv.FormatBool(record.Header.Deleted), strings.Join(record.Header.SetSpecs, ",")})
	}
	if err := output.WriteTable(cmd.OutOrStdout(), []string{"IDENTIFIER", "DATESTAMP", "DELETED", "SETS"}, rows); err != nil {
		return err
	}
	if !next.Empty() {
		expiry := "expiry not advertised"
		if !next.ExpiresAt.IsZero() {
			expiry = "expires " + next.ExpiresAt.Format(time.RFC3339)
		}
		_, err := fmt.Fprintf(cmd.OutOrStdout(), "Next resumption token: %s (%s)\n", next.Value, expiry)
		return err
	}
	return nil
}

func newZenodoOAISetsCommand(client zenodoOAIClient) *cobra.Command {
	return &cobra.Command{Use: "sets", Short: "List Zenodo OAI-PMH sets", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		mode, err := resolveOutputMode(cmd)
		if err != nil {
			return err
		}
		sets, err := client.ListSets(cmd.Context())
		if err != nil {
			return err
		}
		if mode == outputModeJSON {
			return output.WriteJSON(cmd.OutOrStdout(), sets)
		}
		rows := make([][]string, 0, len(sets))
		for _, set := range sets {
			rows = append(rows, []string{set.Spec, set.Name})
		}
		return output.WriteTable(cmd.OutOrStdout(), []string{"SPEC", "NAME"}, rows)
	}}
}

func newZenodoOAIFormatsCommand(client zenodoOAIClient) *cobra.Command {
	var identifier string
	command := &cobra.Command{Use: "formats", Short: "List Zenodo OAI-PMH metadata formats", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		mode, err := resolveOutputMode(cmd)
		if err != nil {
			return err
		}
		formats, err := client.ListMetadataFormats(cmd.Context(), identifier)
		if err != nil {
			return err
		}
		if mode == outputModeJSON {
			return output.WriteJSON(cmd.OutOrStdout(), formats)
		}
		rows := make([][]string, 0, len(formats))
		for _, format := range formats {
			rows = append(rows, []string{format.Prefix, format.Schema, format.NamespaceURL})
		}
		return output.WriteTable(cmd.OutOrStdout(), []string{"PREFIX", "SCHEMA", "NAMESPACE"}, rows)
	}}
	command.Flags().StringVar(&identifier, "identifier", "", "optional OAI identifier")
	return command
}
