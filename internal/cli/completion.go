package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newCompletionCommand(root *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion",
		Short: "Generate shell completion scripts",
		Long:  "Generate shell completion scripts for supported shells.",
	}

	cmd.AddCommand(
		newShellCompletionCommand(root, "bash", "Generate a Bash completion script", func(cmd *cobra.Command) error {
			return root.GenBashCompletionV2(cmd.OutOrStdout(), true)
		}),
		newShellCompletionCommand(root, "zsh", "Generate a Zsh completion script", func(cmd *cobra.Command) error {
			return root.GenZshCompletion(cmd.OutOrStdout())
		}),
		newShellCompletionCommand(root, "fish", "Generate a Fish completion script", func(cmd *cobra.Command) error {
			return root.GenFishCompletion(cmd.OutOrStdout(), true)
		}),
		newShellCompletionCommand(root, "powershell", "Generate a PowerShell completion script", func(cmd *cobra.Command) error {
			return root.GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
		}),
	)

	return cmd
}

func newShellCompletionCommand(root *cobra.Command, name, description string, run func(*cobra.Command) error) *cobra.Command {
	return &cobra.Command{
		Use:   name,
		Short: description,
		Long:  description + ".",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := run(cmd); err != nil {
				return fmt.Errorf("generate %s completion: %w", name, err)
			}
			return nil
		},
	}
}
