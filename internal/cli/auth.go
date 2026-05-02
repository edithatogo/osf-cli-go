package cli

import (
	"osf-cli-go/internal/output"

	"github.com/spf13/cobra"
)

type authUserRecord struct {
	ID         string `json:"id"`
	FullName   string `json:"full_name"`
	GivenName  string `json:"given_name,omitempty"`
	FamilyName string `json:"family_name,omitempty"`
}

func newAuthCommand(client readonlyClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage OSF personal access tokens",
		Long:  "Manage OSF personal access tokens.",
	}
	cmd.AddCommand(newAuthWhoamiCommand(client))
	return cmd
}

func newAuthWhoamiCommand(client readonlyClient) *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show the active authenticated OSF account",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			outputMode, err := resolveOutputMode(cmd)
			if err != nil {
				return err
			}

			user, err := client.CurrentUser(cmd.Context())
			if err != nil {
				return err
			}

			record := authUserRecord{
				ID:         user.ID,
				FullName:   user.Attributes.FullName,
				GivenName:  user.Attributes.GivenName,
				FamilyName: user.Attributes.FamilyName,
			}

			if outputMode == outputModeJSON {
				return output.WriteJSON(cmd.OutOrStdout(), record)
			}

			rows := [][]string{
				{"ID", record.ID},
				{"Full Name", record.FullName},
				{"Given Name", record.GivenName},
				{"Family Name", record.FamilyName},
			}
			return output.WriteTable(cmd.OutOrStdout(), []string{"FIELD", "VALUE"}, rows)
		},
	}
}
