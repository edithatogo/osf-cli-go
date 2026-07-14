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
	defaultEvidencePath = "conductor/tracks/live-osf-release-validation_20260714/live-validation-evidence.md"
	osfCommandPath      = "./cmd/osf"
)

type validationEnv struct {
	liveEnabled bool
	token       string
	username    string
	password    string
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
	username, _ := source.Lookup(auth.UsernameEnv)
	password, _ := source.Lookup(auth.PasswordEnv)
	projectRef, _ := source.Lookup("OSF_VALIDATE_PROJECT")
	downloadRef, _ := source.Lookup("OSF_VALIDATE_DOWNLOAD")

	return validationEnv{
		liveEnabled: truthyLookup(source, "OSF_LIVE_VALIDATION"),
		token:       strings.TrimSpace(token),
		username:    strings.TrimSpace(username),
		password:    strings.TrimSpace(password),
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
			result.Output = "command not yet available in this repository"
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

	downloadCommand := fmt.Sprintf("files download --file %s <temp-dir>", downloadRef)
	if env.downloadRef != "" {
		downloadCommand = fmt.Sprintf("files download --file %s %s", downloadRef, filepath.Join(os.TempDir(), "osf-cli-go-livevalidation-download"))
	}

	return []validationStep{
		{Name: "auth whoami", Command: "auth whoami", Executable: true},
		{Name: "projects list", Command: "projects list", Executable: true},
		{Name: "projects get", Command: "projects get " + project, Executable: true},
		{Name: "components list", Command: "components list " + project, Executable: true},
		{Name: "files list", Command: "files list " + project, Executable: true},
		{Name: "files addons", Command: "files addons " + project, Executable: true},
		{Name: "export", Command: "export " + project + " --json", Executable: true},
		{Name: "search", Command: "search open --limit 5 --json", Executable: true},
		{Name: "preprints list", Command: "preprints list --limit 5 --json", Executable: true},
		{Name: "files download", Command: downloadCommand, Executable: env.downloadRef != ""},
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
	cmd.Env = append(os.Environ(), auth.TokenEnv+"="+env.token, auth.UsernameEnv+"="+env.username, auth.PasswordEnv+"="+env.password)
	if env.projectRef != "" {
		cmd.Env = append(cmd.Env, "OSF_VALIDATE_PROJECT="+env.projectRef)
	}
	if env.downloadRef != "" {
		cmd.Env = append(cmd.Env, "OSF_VALIDATE_DOWNLOAD="+env.downloadRef)
	}

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
