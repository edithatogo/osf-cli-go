package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
)

const version = "0.0.0-dev"

const (
	outputModeTable = "table"
	outputModeJSON  = "json"
)

var errPlannedCommand = errors.New("planned command")

type rootContract struct {
	Name          string          `json:"name"`
	Version       string          `json:"version"`
	DefaultOutput string          `json:"default_output"`
	OutputModes   []string        `json:"output_modes"`
	ExitCodes     map[string]int  `json:"exit_codes"`
	Commands      []contractEntry `json:"commands"`
}

type contractEntry struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

// Run executes the osf CLI and returns a process exit code.
func Run(args []string, stdout io.Writer, stderr io.Writer) int {
	root := newRootCommand(stdout, stderr)
	root.SetArgs(args)

	if err := root.Execute(); err != nil {
		fmt.Fprintln(stderr, err)
		return exitCodeForError(err)
	}

	return 0
}

func newRootCommand(stdout, stderr io.Writer) *cobra.Command {
	root := &cobra.Command{
		Use:           "osf",
		Short:         "A command-line client for the Open Science Framework.",
		Long:          "osf is a command-line client for the Open Science Framework.\n\nThe root command prints human help by default and can emit a machine-readable contract with --output json.",
		Version:       version,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return unknownCommandError(cmd.CommandPath(), args[0])
			}

			outputMode, err := resolveOutputMode(cmd)
			if err != nil {
				return err
			}

			if outputMode == outputModeJSON {
				return writeRootContract(cmd.OutOrStdout())
			}

			return cmd.Help()
		},
	}

	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetVersionTemplate("{{.Version}}\n")
	root.PersistentFlags().String("output", outputModeTable, "output mode: table or json")
	root.PersistentFlags().Bool("json", false, "shorthand for --output json")
	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		_, err := resolveOutputMode(cmd)
		return err
	}

	root.AddCommand(
		newPlannedCommand("auth", "Manage OSF personal access tokens"),
		newPlannedCommand("projects", "List and inspect OSF projects and components"),
		newPlannedCommand("components", "List project components"),
		newPlannedCommand("files", "List, download, and upload OSF Storage files"),
	)

	return root
}

func newPlannedCommand(name, description string) *cobra.Command {
	return &cobra.Command{
		Use:   name,
		Short: description + " (planned)",
		Long:  description + "\n\nThis command is planned for a later track and is not available in this build.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("%w: %s", errPlannedCommand, cmd.CommandPath())
		},
	}
}

func resolveOutputMode(cmd *cobra.Command) (string, error) {
	outputMode, err := cmd.Flags().GetString("output")
	if err != nil {
		return "", err
	}

	jsonMode, err := cmd.Flags().GetBool("json")
	if err != nil {
		return "", err
	}

	outputChanged := cmd.Flags().Changed("output")
	jsonChanged := cmd.Flags().Changed("json")

	if jsonChanged && outputChanged && outputMode != outputModeJSON {
		return "", fmt.Errorf("cannot combine --json with --output=%s", outputMode)
	}

	if jsonMode {
		outputMode = outputModeJSON
	}

	switch outputMode {
	case outputModeTable, outputModeJSON:
		return outputMode, nil
	default:
		return "", fmt.Errorf("invalid output mode %q", outputMode)
	}
}

func writeRootContract(w io.Writer) error {
	return json.NewEncoder(w).Encode(rootContract{
		Name:          "osf",
		Version:       version,
		DefaultOutput: outputModeTable,
		OutputModes:   []string{outputModeTable, outputModeJSON},
		ExitCodes: map[string]int{
			"success":           0,
			"planned_command":   1,
			"usage_or_argument": 2,
		},
		Commands: []contractEntry{
			{Name: "auth", Status: "pending", Description: "Manage OSF personal access tokens"},
			{Name: "projects", Status: "pending", Description: "List and inspect OSF projects and components"},
			{Name: "components", Status: "pending", Description: "List project components"},
			{Name: "files", Status: "pending", Description: "List, download, and upload OSF Storage files"},
		},
	})
}

func unknownCommandError(commandPath, name string) error {
	return fmt.Errorf("unknown command %q for %q", name, commandPath)
}

func exitCodeForError(err error) int {
	switch {
	case errors.Is(err, errPlannedCommand):
		return 1
	case isUsageError(err):
		return 2
	default:
		return 1
	}
}

func isUsageError(err error) bool {
	if err == nil {
		return false
	}

	message := err.Error()
	return strings.Contains(message, "unknown command") ||
		strings.Contains(message, "invalid output mode") ||
		strings.Contains(message, "cannot combine --json") ||
		strings.Contains(message, "unknown flag") ||
		strings.Contains(message, "flag needs an argument") ||
		strings.Contains(message, "unknown shorthand flag")
}
