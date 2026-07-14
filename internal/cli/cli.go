package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/edithatogo/osf-cli-go/internal/auth"
	"github.com/edithatogo/osf-cli-go/internal/buildinfo"
	"github.com/edithatogo/osf-cli-go/internal/observability"
	"github.com/spf13/cobra"
)

var version = "0.0.0-dev"
var buildCommit = ""
var buildDate = ""

const (
	outputModeTable = "table"
	outputModeJSON  = "json"
)

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
	emitter, closer, err := observability.OpenFromEnv(stderr)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 2
	}
	defer func() { _ = closer.Close() }()
	root := newRootCommandWithProviders(stdout, stderr, newDefaultReadonlyClientWithObserver(auth.EnvSource{}, emitter), newDefaultZenodoRESTClient(emitter), newDefaultZenodoOAIClient(emitter))
	root.SetArgs(args)
	ctx := observability.WithEmitter(observability.WithOperationID(WithSignal(context.Background()), observability.NewID("op")), emitter)
	root.SetContext(ctx)
	started := time.Now()
	observability.Emit(ctx, emitter, observability.Event{
		Name:       "cli.command",
		Fields:     map[string]any{"argumentCount": len(args)},
		RetryCount: 0,
	})

	if err := root.Execute(); err != nil {
		err = auth.RedactError(err)
		observability.Emit(ctx, emitter, observability.Event{
			Level:      observability.LevelError,
			Name:       "cli.command.result",
			DurationMS: time.Since(started).Milliseconds(),
			Outcome:    observability.OutcomeError,
			Error:      observability.RedactedError(err),
		})
		_, _ = fmt.Fprintln(stderr, err)
		return exitCodeForError(err)
	}
	observability.Emit(ctx, emitter, observability.Event{
		Name:       "cli.command.result",
		DurationMS: time.Since(started).Milliseconds(),
		Outcome:    observability.OutcomeOK,
	})

	return 0
}

func newRootCommandWithClient(stdout, stderr io.Writer, client readonlyClient) *cobra.Command {
	return newRootCommandWithClients(stdout, stderr, client, newDefaultZenodoOAIClient(nil))
}

func newRootCommandWithClients(stdout, stderr io.Writer, client readonlyClient, oai zenodoOAIClient) *cobra.Command {
	return newRootCommandWithProviders(stdout, stderr, client, newDefaultZenodoRESTClient(nil), oai)
}

func newRootCommandWithProviders(stdout, stderr io.Writer, client readonlyClient, rest zenodoRESTClient, oai zenodoOAIClient) *cobra.Command {
	if client == nil {
		client = newDefaultReadonlyClient()
	}
	if oai == nil {
		oai = newDefaultZenodoOAIClient(nil)
	}
	if rest == nil {
		rest = newDefaultZenodoRESTClient(nil)
	}

	root := &cobra.Command{
		Use:           "osf",
		Short:         "A command-line client for the Open Science Framework.",
		Long:          "osf is a command-line client for the Open Science Framework.\n\nThe root command prints human help by default and can emit a machine-readable contract with --output json.",
		Version:       versionString(),
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
		newAuthCommand(client),
		newProjectsCommand(client),
		newComponentsCommand(client),
		newFilesCommand(client),
		newNodeRelationsCommand(client),
		newExportCommand(client),
		newValidateCommand(client),
		newSearchCommand(client),
		newPreprintsCommand(client),
		newRegistrationsCommand(client),
		newResolveCommand(),
		newOpenCommand(),
		newWhoamiCommand(client),
		newZenodoCommand(rest, oai),
		newCompletionCommand(root),
	)

	return root
}

func versionString() string {
	effectiveVersion := buildinfo.Version(version)
	metadata := make([]string, 0, 2)
	if buildCommit != "" {
		metadata = append(metadata, "commit "+buildCommit)
	}
	if buildDate != "" {
		metadata = append(metadata, "built "+buildDate)
	}

	if len(metadata) == 0 {
		return effectiveVersion
	}

	return fmt.Sprintf("%s (%s)", effectiveVersion, strings.Join(metadata, ", "))
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
		Version:       versionString(),
		DefaultOutput: outputModeTable,
		OutputModes:   []string{outputModeTable, outputModeJSON},
		ExitCodes: map[string]int{
			"success":           0,
			"runtime_error":     1,
			"usage_or_argument": 2,
		},
		Commands: []contractEntry{
			{Name: "auth", Status: "implemented", Description: "Manage OSF authentication and token bootstrap guidance"},
			{Name: "projects", Status: "implemented", Description: "List, inspect, create, update, and delete OSF projects and components"},
			{Name: "components", Status: "implemented", Description: "List project components"},
			{Name: "files", Status: "implemented", Description: "List, download, upload, create folders, and delete OSF Storage files"},
			{Name: "nodes", Status: "implemented", Description: "List OSF wiki pages, comments, logs, and identifiers"},
			{Name: "export", Status: "implemented", Description: "Export a node snapshot"},
			{Name: "validate", Status: "implemented", Description: "Validate OSF node metadata against a deterministic profile"},
			{Name: "search", Status: "implemented", Description: "Search OSF projects and components"},
			{Name: "preprints", Status: "implemented", Description: "List OSF preprints"},
			{Name: "registrations", Status: "implemented", Description: "Create OSF draft registrations"},
			{Name: "resolve", Status: "implemented", Description: "Resolve an OSF DOI or DOI URL"},
			{Name: "open", Status: "implemented", Description: "Open an OSF node in the default browser"},
			{Name: "whoami", Status: "implemented", Description: "Show the active OSF account (alias for auth whoami)"},
			{Name: "zenodo", Status: "implemented", Description: "Search and inspect public Zenodo records or harvest OAI-PMH metadata"},
			{Name: "completion", Status: "implemented", Description: "Generate shell completion scripts"},
		},
	})
}

func unknownCommandError(commandPath, name string) error {
	return fmt.Errorf("unknown command %q for %q", name, commandPath)
}

func exitCodeForError(err error) int {
	if isUsageError(err) {
		return 2
	}
	return 1
}

func isUsageError(err error) bool {
	if err == nil {
		return false
	}
	message := err.Error()
	return strings.Contains(message, "unknown command") ||
		strings.Contains(message, "invalid output mode") ||
		strings.Contains(message, "cannot combine --json") ||
		strings.Contains(message, "cannot combine --file with --tree") ||
		strings.Contains(message, "unknown flag") ||
		strings.Contains(message, "flag needs an argument") ||
		strings.Contains(message, "unknown shorthand flag") ||
		strings.Contains(message, "accepts") ||
		strings.Contains(message, "unsupported conflict policy") ||
		strings.Contains(message, "unsupported file source URL")
}
