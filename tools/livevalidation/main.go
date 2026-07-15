package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/edithatogo/osf-cli-go/internal/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	defaultEvidencePath      = "docs/live-osf-validation-evidence.md"
	osfCommandPath           = "./cmd/osf"
	cancellationProbeTimeout = time.Nanosecond
)

type validationEnv struct {
	liveEnabled   bool
	token         string
	username      string
	password      string
	projectRef    string
	downloadRef   string
	writesEnabled bool
	fixturePath   string
	fixtureName   string
	downloadPath  string
}

type validationStep struct {
	Name       string
	Command    string
	Args       []string
	Executable bool
	Mode       stepMode
}

type stepMode int

const (
	stepNormal stepMode = iota
	stepExpectedFailure
	stepCancellation
	stepMCP
	stepCleanup
)

type osfStepRunner func(context.Context, time.Duration, validationEnv, validationStep) (string, error)
type mcpStepRunner func(context.Context, time.Duration, validationEnv) (string, error)

type validationResult struct {
	Step    validationStep
	Status  string
	Output  string
	Elapsed time.Duration
}

type validationReport struct {
	GeneratedAt time.Time
	Mode        string
	Skipped     bool
	SkipReason  string
	Env         validationEnv
	Steps       []validationResult
}

func main() {
	var (
		liveMode     bool
		evidencePath string
		timeout      time.Duration
	)

	flag.BoolVar(&liveMode, "live", false, "run live OSF validation when explicit env vars are present")
	flag.StringVar(&evidencePath, "evidence", defaultEvidencePath, "write a markdown evidence report to this path")
	flag.DurationVar(&timeout, "timeout", 2*time.Minute, "timeout for each live OSF command")
	flag.Parse()

	env := loadValidationEnv(auth.EnvSource{})
	report, runErr := runValidation(context.Background(), env, liveMode, timeout)
	if err := writeEvidence(evidencePath, report); err != nil {
		fmt.Fprintf(os.Stderr, "livevalidation: write evidence: %v\n", err)
		os.Exit(1)
	}
	if runErr != nil {
		fmt.Fprintf(os.Stderr, "livevalidation: %v\n", runErr)
		os.Exit(1)
	}

	if report.Skipped {
		_, _ = fmt.Fprintln(os.Stdout, report.SkipReason)
		return
	}

	if report.Mode == "dry-run" {
		_, _ = fmt.Fprintln(os.Stdout, "dry-run validation report written")
		return
	}

	if hasFailures(report.Steps) {
		os.Exit(1)
	}
}

func runValidation(ctx context.Context, env validationEnv, liveMode bool, timeout time.Duration) (validationReport, error) {
	return runValidationWithRunners(ctx, env, liveMode, timeout, runOSFStep, runMCPCommand)
}

func runValidationWithRunners(ctx context.Context, env validationEnv, liveMode bool, timeout time.Duration, osfRunner osfStepRunner, mcpRunner mcpStepRunner) (validationReport, error) {
	report := validationReport{
		GeneratedAt: time.Now().UTC(),
		Env:         env,
	}

	if !liveMode {
		report.Mode = "dry-run"
		report.Steps = plannedSteps(env)
		return report, nil
	}

	if !env.liveEnabled {
		report.Mode = "skipped"
		report.Skipped = true
		report.SkipReason = "live OSF validation skipped: set OSF_LIVE_VALIDATION=1 to opt in"
		report.Steps = plannedSteps(env)
		return report, nil
	}

	if (env.token == "" && (env.username == "" || env.password == "")) || env.projectRef == "" {
		report.Mode = "skipped"
		report.Skipped = true
		report.SkipReason = missingEnvMessage(env)
		report.Steps = plannedSteps(env)
		return report, nil
	}
	fixture, err := os.CreateTemp("", "osf-cli-go-livevalidation-*.txt")
	if err != nil {
		return report, fmt.Errorf("create validation fixture: %w", err)
	}
	fixturePath := fixture.Name()
	if _, err := fixture.WriteString("OSF CLI Go disposable live-validation fixture\n"); err != nil {
		_ = fixture.Close()
		_ = os.Remove(fixturePath)
		return report, fmt.Errorf("write validation fixture: %w", err)
	}
	if err := fixture.Close(); err != nil {
		_ = os.Remove(fixturePath)
		return report, fmt.Errorf("close validation fixture: %w", err)
	}
	defer func() { _ = os.Remove(fixturePath) }()
	env.fixturePath = fixturePath
	env.fixtureName = filepath.Base(fixturePath)
	if env.downloadRef != "" {
		downloadDir, err := os.MkdirTemp("", "osf-cli-go-livevalidation-download-*")
		if err != nil {
			return report, fmt.Errorf("create validation download directory: %w", err)
		}
		defer func() { _ = os.RemoveAll(downloadDir) }()
		env.downloadPath = filepath.Join(downloadDir, "downloaded-fixture")
	}
	report.Env = env

	report.Mode = "live"
	steps := executableSteps(env)
	results := make([]validationResult, 0, len(steps))
	var stepErr error
	uploadPassed := false
	for _, step := range steps {
		result := validationResult{Step: step}
		if !step.Executable {
			result.Status = "pending"
			result.Output = pendingReason(step)
			results = append(results, result)
			continue
		}

		if step.Mode == stepExpectedFailure && !uploadPassed {
			result.Status = "skipped"
			result.Output = "conflict check skipped because the initial upload failed"
			results = append(results, result)
			continue
		}
		start := time.Now()
		runContext := ctx
		runTimeout := timeout
		if step.Mode == stepCleanup {
			runContext = context.WithoutCancel(ctx)
		}
		if step.Mode == stepCancellation {
			runTimeout = cancellationProbeTimeout
		}
		var output string
		var runErr error
		if step.Mode == stepMCP {
			output, runErr = mcpRunner(runContext, runTimeout, env)
		} else {
			output, runErr = osfRunner(runContext, runTimeout, env, step)
		}
		result.Elapsed = time.Since(start)
		switch {
		case step.Mode == stepExpectedFailure && isExpectedConflict(output, runErr):
			result.Status = "passed"
			result.Output = "existing-file conflict rejected as expected"
		case step.Mode == stepExpectedFailure && runErr != nil:
			result.Status = "failed"
			result.Output = "upload failed without a recognized existing-file conflict"
			stepErr = combineErrors(stepErr, fmt.Errorf("%s: %w", step.Name, runErr))
		case step.Mode == stepExpectedFailure:
			result.Status = "failed"
			result.Output = output
			runErr = errors.New("expected existing-file conflict was accepted")
			stepErr = combineErrors(stepErr, fmt.Errorf("%s: %w", step.Name, runErr))
		case step.Mode == stepCancellation && (errors.Is(runErr, context.DeadlineExceeded) || errors.Is(runErr, context.Canceled)):
			result.Status = "passed"
			result.Output = "command stopped at the cancellation deadline"
		case runErr != nil:
			result.Status = "failed"
			result.Output = "command failed; inspect the local validation error"
			stepErr = combineErrors(stepErr, fmt.Errorf("%s: %w", step.Name, runErr))
		default:
			result.Status = "passed"
			result.Output = "command completed successfully"
		}
		if step.Name == "files upload" && result.Status == "passed" {
			uploadPassed = true
		}
		results = append(results, result)
	}
	report.Steps = results
	if hasIncomplete(results) {
		stepErr = combineErrors(stepErr, errors.New("incomplete live validation: pending, skipped, or failed scenarios remain"))
	}
	if stepErr != nil {
		return report, stepErr
	}
	return report, nil
}

func loadValidationEnv(source auth.Source) validationEnv {
	token, _ := source.Lookup(auth.TokenEnv)
	username, _ := source.Lookup(auth.UsernameEnv)
	password, _ := source.Lookup(auth.PasswordEnv)
	projectRef, _ := source.Lookup("OSF_VALIDATE_PROJECT")
	downloadRef, _ := source.Lookup("OSF_VALIDATE_DOWNLOAD")

	return validationEnv{
		liveEnabled:   truthyLookup(source, "OSF_LIVE_VALIDATION"),
		writesEnabled: truthyLookup(source, "OSF_VALIDATE_WRITES"),
		token:         strings.TrimSpace(token),
		username:      strings.TrimSpace(username),
		password:      strings.TrimSpace(password),
		projectRef:    strings.TrimSpace(projectRef),
		downloadRef:   strings.TrimSpace(downloadRef),
	}
}

func truthyLookup(source auth.Source, name string) bool {
	raw, ok := source.Lookup(name)
	if !ok {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func missingEnvMessage(env validationEnv) string {
	var missing []string
	if env.token == "" && (env.username == "" || env.password == "") {
		missing = append(missing, auth.TokenEnv+" or "+auth.UsernameEnv+"/"+auth.PasswordEnv)
	}
	if env.projectRef == "" {
		missing = append(missing, "OSF_VALIDATE_PROJECT")
	}
	if len(missing) == 0 {
		return "live OSF validation skipped: required environment variables are present"
	}
	return fmt.Sprintf("live OSF validation skipped: missing %s", strings.Join(missing, ", "))
}

func plannedSteps(env validationEnv) []validationResult {
	steps := executableSteps(env)
	results := make([]validationResult, 0, len(steps))
	for _, step := range steps {
		result := validationResult{Step: step}
		if step.Executable {
			result.Status = "planned"
			result.Output = "not executed in dry-run mode"
		} else {
			result.Status = "pending"
			result.Output = pendingReason(step)
		}
		results = append(results, result)
	}
	return results
}

func executableSteps(env validationEnv) []validationStep {
	project := "<project>"
	if env.projectRef != "" {
		project = env.projectRef
	}

	downloadRef := "<download-ref>"
	if env.downloadRef != "" {
		downloadRef = env.downloadRef
	}

	downloadArgs := []string{"files", "download", "--file", downloadRef, "<temp-dir>"}
	if env.downloadRef != "" {
		destination := env.downloadPath
		if destination == "" {
			destination = "<generated-temp-dir>"
		}
		downloadArgs = []string{"files", "download", "--file", downloadRef, destination}
	}
	fixturePath := env.fixturePath
	fixtureName := env.fixtureName
	if fixturePath == "" {
		fixturePath = "<generated-fixture>"
		fixtureName = "<generated-fixture-name>"
	}

	return []validationStep{
		newValidationStep("auth whoami", stepNormal, true, "auth", "whoami"),
		newValidationStep("projects list", stepNormal, true, "projects", "list"),
		newValidationStep("projects get", stepNormal, true, "projects", "get", project),
		newValidationStep("components list", stepNormal, true, "components", "list", project),
		newValidationStep("files list", stepNormal, true, "files", "list", project),
		newValidationStep("files addons", stepNormal, true, "files", "addons", project),
		newValidationStep("export", stepNormal, true, "export", project, "--json"),
		newValidationStep("search", stepNormal, true, "search", "open", "--limit", "5", "--json"),
		newValidationStep("preprints list", stepNormal, true, "preprints", "list", "--limit", "5", "--json"),
		newValidationStep("files upload", stepNormal, env.writesEnabled, "files", "upload", "--node", project, fixturePath, "--conflict", "overwrite"),
		newValidationStep("files upload conflict", stepExpectedFailure, env.writesEnabled, "files", "upload", "--node", project, fixturePath, "--conflict", "fail"),
		newValidationStep("cancellation", stepCancellation, true, "projects", "list", "--json"),
		{Name: "MCP project get", Command: "osf_project_get", Executable: true, Mode: stepMCP},
		newValidationStep("files cleanup", stepCleanup, env.writesEnabled, "files", "rm", "--node", project, fixtureName, "--yes"),
		newValidationStep("files download", stepNormal, env.downloadRef != "", downloadArgs...),
	}
}

func newValidationStep(name string, mode stepMode, executable bool, args ...string) validationStep {
	return validationStep{Name: name, Command: strings.Join(args, " "), Args: args, Executable: executable, Mode: mode}
}

func pendingReason(step validationStep) string {
	switch step.Name {
	case "files upload", "files upload conflict", "files cleanup":
		return "set OSF_VALIDATE_WRITES=1 to opt in to disposable writes"
	case "files download":
		return "set OSF_VALIDATE_DOWNLOAD to a disposable fixture file reference"
	default:
		return "scenario is not enabled"
	}
}

func isExpectedConflict(output string, err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(output + " " + err.Error())
	for _, marker := range []string{"already exists", "exists in this location", "conflict", "status 409", "status code 409"} {
		if strings.Contains(message, marker) {
			return true
		}
	}
	return false
}

func runOSFStep(parent context.Context, timeout time.Duration, env validationEnv, step validationStep) (string, error) {
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	if len(step.Args) == 0 {
		return "", fmt.Errorf("empty command")
	}

	cmdArgs := append([]string{"run", osfCommandPath}, step.Args...)
	cmd := exec.CommandContext(ctx, "go", cmdArgs...)
	cmd.Env = commandEnvironment(env)

	out, err := cmd.CombinedOutput()
	output := auth.Redact(string(out), env.token, env.username, env.password, env.projectRef, env.downloadRef)
	if ctx.Err() != nil {
		return output, ctx.Err()
	}
	if err != nil {
		return output, auth.RedactError(err, env.token, env.username, env.password, env.projectRef, env.downloadRef)
	}
	return output, nil
}

func runMCPCommand(parent context.Context, timeout time.Duration, env validationEnv) (string, error) {
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "run", "./cmd/osf-mcp")
	command.Env = commandEnvironment(env)
	client := mcp.NewClient(&mcp.Implementation{Name: "osf-cli-go-livevalidation", Version: "1.0.0"}, nil)
	session, err := client.Connect(ctx, &mcp.CommandTransport{Command: command}, nil)
	if err != nil {
		return "", auth.RedactError(err, env.token, env.username, env.password, env.projectRef)
	}
	defer func() { _ = session.Close() }()
	arguments, err := json.Marshal(map[string]string{"id": env.projectRef})
	if err != nil {
		return "", err
	}
	result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "osf_project_get", Arguments: json.RawMessage(arguments)})
	if err != nil {
		return "", auth.RedactError(err, env.token, env.username, env.password, env.projectRef)
	}
	if result.IsError {
		return "", errors.New("MCP osf_project_get returned an error result")
	}
	if result.StructuredContent == nil && len(result.Content) == 0 {
		return "", errors.New("MCP osf_project_get returned no content")
	}
	return "MCP osf_project_get returned structured content", nil
}

func commandEnvironment(env validationEnv) []string {
	values := append(os.Environ(), auth.TokenEnv+"="+env.token, auth.UsernameEnv+"="+env.username, auth.PasswordEnv+"="+env.password)
	if env.projectRef != "" {
		values = append(values, "OSF_VALIDATE_PROJECT="+env.projectRef)
	}
	if env.downloadRef != "" {
		values = append(values, "OSF_VALIDATE_DOWNLOAD="+env.downloadRef)
	}
	return values
}

func hasFailures(results []validationResult) bool {
	for _, result := range results {
		if result.Status == "failed" {
			return true
		}
	}
	return false
}

func hasIncomplete(results []validationResult) bool {
	for _, result := range results {
		if result.Status != "passed" {
			return true
		}
	}
	return false
}

func combineErrors(existing error, next error) error {
	if existing == nil {
		return next
	}
	return fmt.Errorf("%w; %v", existing, next)
}

func writeEvidence(path string, report validationReport) error {
	if path == "" {
		return nil
	}

	var builder strings.Builder
	builder.WriteString("# Live OSF Validation Evidence\n\n")
	fmt.Fprintf(&builder, "- Generated: %s\n", report.GeneratedAt.Format(time.RFC3339))
	fmt.Fprintf(&builder, "- Mode: %s\n", report.Mode)
	if report.Skipped {
		fmt.Fprintf(&builder, "- Skip reason: %s\n", report.SkipReason)
	}
	builder.WriteString("- Environment:\n")
	fmt.Fprintf(&builder, "  - %s: %s\n", auth.TokenEnv, presence(report.Env.token))
	fmt.Fprintf(&builder, "  - %s: %s\n", auth.UsernameEnv, presence(report.Env.username))
	fmt.Fprintf(&builder, "  - %s: %s\n", auth.PasswordEnv, presence(report.Env.password))
	fmt.Fprintf(&builder, "  - OSF_VALIDATE_PROJECT: %s\n", presence(report.Env.projectRef))
	fmt.Fprintf(&builder, "  - OSF_LIVE_VALIDATION: %t\n", report.Env.liveEnabled)
	fmt.Fprintf(&builder, "  - OSF_VALIDATE_WRITES: %t\n", report.Env.writesEnabled)
	builder.WriteString("- Planned coverage:\n")
	for _, step := range plannedSteps(report.Env) {
		fmt.Fprintf(&builder, "  - %s: %s\n", step.Step.Name, step.Status)
	}
	builder.WriteString("- Results:\n")
	for _, result := range report.Steps {
		fmt.Fprintf(&builder, "  - %s: %s\n", result.Step.Name, result.Status)
		if result.Output != "" {
			output := sanitizeOutput(result.Output)
			if report.Mode == "skipped" && result.Status == "planned" {
				output = "not executed; live validation was skipped"
			}
			fmt.Fprintf(&builder, "    - Output: %s\n", output)
		}
		if result.Elapsed > 0 {
			fmt.Fprintf(&builder, "    - Elapsed: %s\n", result.Elapsed.Round(time.Millisecond))
		}
	}
	if !containsStep(report.Steps, "files download") {
		builder.WriteString("- Future coverage:\n")
		builder.WriteString("  - files download: pending until the download command lands\n")
	}
	builder.WriteString("\n")

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(builder.String()), 0o644)
}

func sanitizeOutput(text string) string {
	return strings.TrimSpace(strings.ReplaceAll(text, "\r\n", "\n"))
}

func presence(value string) string {
	if strings.TrimSpace(value) == "" {
		return "missing"
	}
	return "set (redacted)"
}

func containsStep(results []validationResult, name string) bool {
	for _, result := range results {
		if result.Step.Name == name {
			return true
		}
	}
	return false
}
