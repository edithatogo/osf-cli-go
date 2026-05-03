package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/edithatogo/osf-cli-go/internal/auth"
)

const (
	defaultEvidencePath = "conductor/tracks/live-osf-validation_20260502/live-validation-evidence.md"
	osfCommandPath      = "./cmd/osf"
)

type validationEnv struct {
	liveEnabled bool
	token       string
	projectRef  string
	downloadRef string
}

type validationStep struct {
	Name       string
	Command    string
	Executable bool
}

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
	report, err := runValidation(context.Background(), env, liveMode, timeout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "livevalidation: %v\n", err)
		os.Exit(1)
	}

	if err := writeEvidence(evidencePath, report); err != nil {
		fmt.Fprintf(os.Stderr, "livevalidation: write evidence: %v\n", err)
		os.Exit(1)
	}

	if report.Skipped {
		fmt.Fprintln(os.Stdout, report.SkipReason)
		return
	}

	if report.Mode == "dry-run" {
		fmt.Fprintln(os.Stdout, "dry-run validation report written")
		return
	}

	if hasFailures(report.Steps) {
		os.Exit(1)
	}
}

func runValidation(ctx context.Context, env validationEnv, liveMode bool, timeout time.Duration) (validationReport, error) {
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

	if env.token == "" || env.projectRef == "" {
		report.Mode = "skipped"
		report.Skipped = true
		report.SkipReason = missingEnvMessage(env)
		report.Steps = plannedSteps(env)
		return report, nil
	}

	report.Mode = "live"
	steps := executableSteps(env)
	results := make([]validationResult, 0, len(steps))
	var stepErr error
	for _, step := range steps {
		result := validationResult{Step: step}
		if !step.Executable {
			result.Status = "pending"
			result.Output = "command not yet available in this repository"
			results = append(results, result)
			continue
		}

		start := time.Now()
		output, runErr := runOSFCommand(ctx, timeout, env, step.Command)
		result.Elapsed = time.Since(start)
		if runErr != nil {
			result.Status = "failed"
			result.Output = output
			stepErr = combineErrors(stepErr, fmt.Errorf("%s: %w", step.Name, runErr))
		} else {
			result.Status = "passed"
			result.Output = output
		}
		results = append(results, result)
	}
	report.Steps = results
	if stepErr != nil {
		return report, stepErr
	}
	return report, nil
}

func loadValidationEnv(source auth.Source) validationEnv {
	token, _ := source.Lookup(auth.TokenEnv)
	projectRef, _ := source.Lookup("OSF_VALIDATE_PROJECT")
	downloadRef, _ := source.Lookup("OSF_VALIDATE_DOWNLOAD")

	return validationEnv{
		liveEnabled: truthyLookup(source, "OSF_LIVE_VALIDATION"),
		token:       strings.TrimSpace(token),
		projectRef:  strings.TrimSpace(projectRef),
		downloadRef: strings.TrimSpace(downloadRef),
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
	if env.token == "" {
		missing = append(missing, auth.TokenEnv)
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
			result.Output = "command not yet available in this repository"
		}
		results = append(results, result)
	}
	return results
}

func executableSteps(env validationEnv) []validationStep {
	project := "<project>"
	if env.projectRef != "" {
		project = "<redacted-project>"
	}

	downloadRef := "<download-ref>"
	if env.downloadRef != "" {
		downloadRef = "<redacted-download-ref>"
	}

	return []validationStep{
		{Name: "auth whoami", Command: "auth whoami", Executable: true},
		{Name: "projects list", Command: "projects list", Executable: true},
		{Name: "projects get", Command: "projects get " + project, Executable: true},
		{Name: "components list", Command: "components list " + project, Executable: true},
		{Name: "files list", Command: "files list " + project, Executable: true},
		{Name: "files download", Command: fmt.Sprintf("files download --file %s <temp-dir>", downloadRef), Executable: true},
	}
}

func runOSFCommand(parent context.Context, timeout time.Duration, env validationEnv, command string) (string, error) {
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	args := strings.Fields(command)
	if len(args) == 0 {
		return "", fmt.Errorf("empty command")
	}

	cmdArgs := append([]string{"run", osfCommandPath}, args...)
	cmd := exec.CommandContext(ctx, "go", cmdArgs...)
	cmd.Env = append(os.Environ(), auth.TokenEnv+"="+env.token)
	if env.projectRef != "" {
		cmd.Env = append(cmd.Env, "OSF_VALIDATE_PROJECT="+env.projectRef)
	}
	if env.downloadRef != "" {
		cmd.Env = append(cmd.Env, "OSF_VALIDATE_DOWNLOAD="+env.downloadRef)
	}

	out, err := cmd.CombinedOutput()
	output := auth.Redact(string(out), env.token, env.projectRef, env.downloadRef)
	if ctx.Err() != nil {
		return output, ctx.Err()
	}
	if err != nil {
		return output, auth.RedactError(err, env.token, env.projectRef, env.downloadRef)
	}
	return output, nil
}

func hasFailures(results []validationResult) bool {
	for _, result := range results {
		if result.Status == "failed" {
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
	builder.WriteString(fmt.Sprintf("- Generated: %s\n", report.GeneratedAt.Format(time.RFC3339)))
	builder.WriteString(fmt.Sprintf("- Mode: %s\n", report.Mode))
	if report.Skipped {
		builder.WriteString(fmt.Sprintf("- Skip reason: %s\n", report.SkipReason))
	}
	builder.WriteString("- Environment:\n")
	builder.WriteString(fmt.Sprintf("  - %s: %s\n", auth.TokenEnv, presence(report.Env.token)))
	builder.WriteString(fmt.Sprintf("  - OSF_VALIDATE_PROJECT: %s\n", presence(report.Env.projectRef)))
	builder.WriteString(fmt.Sprintf("  - OSF_LIVE_VALIDATION: %t\n", report.Env.liveEnabled))
	builder.WriteString("- Planned coverage:\n")
	for _, step := range plannedSteps(report.Env) {
		builder.WriteString(fmt.Sprintf("  - %s: %s\n", step.Step.Name, step.Status))
	}
	builder.WriteString("- Results:\n")
	for _, result := range report.Steps {
		builder.WriteString(fmt.Sprintf("  - %s: %s\n", result.Step.Name, result.Status))
		if result.Output != "" {
			builder.WriteString(fmt.Sprintf("    - Output: %s\n", sanitizeOutput(result.Output)))
		}
		if result.Elapsed > 0 {
			builder.WriteString(fmt.Sprintf("    - Elapsed: %s\n", result.Elapsed.Round(time.Millisecond)))
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
