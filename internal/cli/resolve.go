package cli

import (
	"fmt"

	"github.com/edithatogo/osf-cli-go/internal/osfapi"
	"github.com/edithatogo/osf-cli-go/internal/output"
	"github.com/spf13/cobra"
)

type doiResolutionRecord struct {
	DOI         string `json:"doi"`
	ResolvedURL string `json:"resolved_url"`
}

func newResolveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "resolve <doi-or-url>",
		Short: "Resolve an OSF DOI or DOI URL",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			outputMode, err := resolveOutputMode(cmd)
			if err != nil {
				return err
			}
			resolution, err := osfapi.ResolveDOI(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("resolve OSF DOI: %w", err)
			}
			record := doiResolutionRecord{DOI: resolution.DOI, ResolvedURL: resolution.ResolvedURL}
			if outputMode == outputModeJSON {
				return output.WriteJSON(cmd.OutOrStdout(), record)
			}
			return output.WriteTable(cmd.OutOrStdout(), []string{"FIELD", "VALUE"}, [][]string{
				{"DOI", record.DOI},
				{"Resolved URL", record.ResolvedURL},
			})
		},
	}
}
