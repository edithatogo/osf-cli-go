package cli

import (
	"fmt"
	"strings"

	"github.com/edithatogo/osf-cli-go/internal/auth"
	"github.com/edithatogo/osf-cli-go/internal/output"

	"github.com/spf13/cobra"
)

type authUserRecord struct {
	ID         string `json:"id"`
	FullName   string `json:"full_name"`
	GivenName  string `json:"given_name,omitempty"`
	FamilyName string `json:"family_name,omitempty"`
	AuthMode   string `json:"auth_mode,omitempty"`
}

type authModeProvider interface {
	AuthMode() auth.Mode
}

func newAuthCommand(client readonlyClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage OSF authentication",
		Long:  "Manage OSF authentication. Personal access tokens are preferred for automation; username/password can be used as a fallback credential source or guided token-bootstrap input.",
	}
	cmd.AddCommand(newAuthWhoamiCommand(client))
	cmd.AddCommand(newAuthLoginCommand())
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
				AuthMode:   currentAuthMode(client),
			}

			if outputMode == outputModeJSON {
				return output.WriteJSON(cmd.OutOrStdout(), record)
			}

			rows := [][]string{
				{"ID", record.ID},
				{"Full Name", record.FullName},
				{"Given Name", record.GivenName},
				{"Family Name", record.FamilyName},
				{"Auth Mode", record.AuthMode},
			}
			return output.WriteTable(cmd.OutOrStdout(), []string{"FIELD", "VALUE"}, rows)
		},
	}
}

func newAuthLoginCommand() *cobra.Command {
	var token string
	var printEnv bool
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Prepare token-backed OSF authentication",
		Long: strings.Join([]string{
			"Prepare token-backed OSF authentication.",
			"",
			"OSF does not currently document a supported API endpoint for minting personal access tokens from a username and password.",
			"This command therefore implements a guided bootstrap workflow: use OSF_USERNAME and OSF_PASSWORD to identify the account context, then create a PAT in OSF Account Settings and pass it with --token or set OSF_TOKEN yourself.",
		}, "\n"),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(token) != "" {
				if printEnv {
					_, err := fmt.Fprintf(cmd.OutOrStdout(), "export %s=%q\n", auth.TokenEnv, strings.TrimSpace(token))
					return err
				}
				rows := [][]string{
					{"Token", "provided (redacted)"},
					{"Next Step", "set OSF_TOKEN in your shell or secret store"},
				}
				return output.WriteTable(cmd.OutOrStdout(), []string{"FIELD", "VALUE"}, rows)
			}

			credentials, _ := auth.LoadCredentials(auth.EnvSource{})
			rows := [][]string{
				{"Token URL", "https://osf.io/settings/tokens/"},
				{"Preferred Mode", string(auth.ModeBearerToken)},
				{"Username Password Env", usernamePasswordStatus(credentials)},
				{"Required Next Step", "create a PAT in OSF Account Settings, then set OSF_TOKEN"},
				{"Suggested Scopes", "read commands: osf.full_read; write commands and WaterButler writes: osf.full_write"},
			}
			return output.WriteTable(cmd.OutOrStdout(), []string{"FIELD", "VALUE"}, rows)
		},
	}
	cmd.Flags().StringVar(&token, "token", "", "personal access token to format for export or acknowledge as supplied")
	cmd.Flags().BoolVar(&printEnv, "print-env", false, "print an OSF_TOKEN shell export command for the supplied --token")
	return cmd
}

func currentAuthMode(client readonlyClient) string {
	if provider, ok := client.(authModeProvider); ok {
		return string(provider.AuthMode())
	}
	return "injected-client"
}

func usernamePasswordStatus(credentials auth.Credentials) string {
	if credentials.Mode == auth.ModeUsernamePassword && credentials.Authenticated() {
		return "set (redacted)"
	}
	return "not active"
}
