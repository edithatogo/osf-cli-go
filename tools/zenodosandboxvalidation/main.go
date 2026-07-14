package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/edithatogo/osf-cli-go/internal/download"
	"github.com/edithatogo/osf-cli-go/internal/zenodotransfer"
)

const (
	defaultEvidencePath = "docs/zenodo-sandbox-validation-evidence.md"
	defaultBaseURL      = "https://sandbox.zenodo.org/api/"
	defaultPayloadBytes = 1 << 20
)

var ErrLiveNotConfigured = errors.New("live Zenodo sandbox validation is not configured")

type validationEnv struct {
	enabled bool
	token   string
	baseURL string
}

type validationStep struct {
	Name   string
	Status string
	Detail string
}

type validationReport struct {
	GeneratedAt time.Time
	Mode        string
	SkipReason  string
	TokenSet    bool
	BaseHost    string
	Steps       []validationStep
}

func main() {
	live := flag.Bool("live", false, "run disposable Zenodo sandbox validation")
	evidence := flag.String("evidence", defaultEvidencePath, "write sanitized markdown evidence")
	timeout := flag.Duration("timeout", 5*time.Minute, "overall live validation timeout")
	payloadBytes := flag.Int("bytes", defaultPayloadBytes, "disposable validation file size")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	env := loadValidationEnv()
	report, runErr := runValidation(ctx, env, *live, *payloadBytes)
	if err := writeEvidence(*evidence, report, env.token); err != nil {
		fmt.Fprintf(os.Stderr, "zenodosandboxvalidation: write evidence: %v\n", err)
		os.Exit(1)
	}
	if runErr != nil {
		fmt.Fprintf(os.Stderr, "zenodosandboxvalidation: %s\n", exactRedact(runErr.Error(), env.token))
		os.Exit(1)
	}
	if report.Mode != "live" {
		_, _ = fmt.Fprintln(os.Stdout, report.SkipReason)
		return
	}
	_, _ = fmt.Fprintln(os.Stdout, "Zenodo sandbox transfer validation passed and disposable draft cleanup succeeded")
}

func loadValidationEnv() validationEnv {
	baseURL := strings.TrimSpace(os.Getenv("ZENODO_BASE_URL"))
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return validationEnv{
		enabled: truthy(os.Getenv("ZENODO_SANDBOX_VALIDATION")),
		token:   strings.TrimSpace(os.Getenv("ZENODO_TOKEN")),
		baseURL: baseURL,
	}
}

func truthy(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func runValidation(ctx context.Context, env validationEnv, live bool, payloadSize int) (report validationReport, err error) {
	report = validationReport{GeneratedAt: time.Now().UTC(), TokenSet: env.token != "", BaseHost: safeHost(env.baseURL)}
	planned := []string{"create disposable draft", "upload and checksum", "verified download", "resumable download", "cleanup"}
	if !live {
		report.Mode = "dry-run"
		report.SkipReason = "dry-run only; pass -live and set ZENODO_SANDBOX_VALIDATION=1 to opt in"
		for _, name := range planned {
			report.Steps = append(report.Steps, validationStep{Name: name, Status: "planned", Detail: "not executed"})
		}
		return report, nil
	}
	if !env.enabled {
		report.Mode = "skipped"
		report.SkipReason = "live Zenodo sandbox validation skipped: set ZENODO_SANDBOX_VALIDATION=1"
		for _, name := range planned {
			report.Steps = append(report.Steps, validationStep{Name: name, Status: "skipped", Detail: "explicit opt-in is absent"})
		}
		return report, fmt.Errorf("%w: set ZENODO_SANDBOX_VALIDATION=1", ErrLiveNotConfigured)
	}
	if env.token == "" {
		report.Mode = "skipped"
		report.SkipReason = "live Zenodo sandbox validation skipped: ZENODO_TOKEN is not set"
		for _, name := range planned {
			report.Steps = append(report.Steps, validationStep{Name: name, Status: "skipped", Detail: "sandbox credential is absent"})
		}
		return report, fmt.Errorf("%w: ZENODO_TOKEN is not set", ErrLiveNotConfigured)
	}
	if payloadSize < 3 || payloadSize > 50<<20 {
		return report, errors.New("validation payload must be between 3 bytes and 50 MiB")
	}
	client, err := zenodotransfer.New(env.baseURL, env.token)
	if err != nil {
		return report, fmt.Errorf("configure sandbox transfer client: %w", err)
	}
	report.Mode = "live"

	tempDir, err := os.MkdirTemp("", "osf-cli-go-zenodo-validation-*")
	if err != nil {
		return report, fmt.Errorf("create validation directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()
	content := validationPayload(payloadSize)
	source := filepath.Join(tempDir, "validation.bin")
	if err := os.WriteFile(source, content, 0o600); err != nil {
		return report, fmt.Errorf("write validation payload: %w", err)
	}

	draft, createErr := client.CreateDraft(ctx)
	if createErr != nil {
		report.Steps = append(report.Steps, validationStep{Name: planned[0], Status: "failed", Detail: "provider did not acknowledge draft creation"})
		return report, createErr
	}
	report.Steps = append(report.Steps, validationStep{Name: planned[0], Status: "passed", Detail: "disposable unpublished draft acknowledged"})
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		cleanupErr := client.DeleteDraft(cleanupCtx, draft.ID)
		if cleanupErr != nil {
			report.Steps = append(report.Steps, validationStep{Name: "cleanup", Status: "failed", Detail: "disposable draft deletion was not acknowledged"})
			err = errors.Join(err, fmt.Errorf("cleanup disposable Zenodo draft: %w", cleanupErr))
			return
		}
		report.Steps = append(report.Steps, validationStep{Name: "cleanup", Status: "passed", Detail: "disposable unpublished draft deleted"})
	}()

	upload, uploadErr := client.UploadFile(ctx, draft, source, "validation.bin", download.ConflictFail)
	if uploadErr != nil {
		report.Steps = append(report.Steps, validationStep{Name: planned[1], Status: "failed", Detail: "upload size or checksum was not verified"})
		return report, uploadErr
	}
	report.Steps = append(report.Steps, validationStep{Name: planned[1], Status: "passed", Detail: fmt.Sprintf("bytes=%d checksum=%s retries=%d", upload.Bytes, upload.Remote.Checksum, upload.RetryCount)})

	destination := filepath.Join(tempDir, "download.bin")
	downloadResult, downloadErr := client.DownloadFile(ctx, upload.Remote, destination, download.ConflictFail)
	if downloadErr != nil {
		report.Steps = append(report.Steps, validationStep{Name: planned[2], Status: "failed", Detail: "download bytes or checksum were not verified"})
		return report, downloadErr
	}
	downloaded, readErr := os.ReadFile(destination)
	if readErr != nil || !bytes.Equal(downloaded, content) {
		report.Steps = append(report.Steps, validationStep{Name: planned[2], Status: "failed", Detail: "downloaded bytes differ from source"})
		return report, errors.New("downloaded Zenodo sandbox bytes differ from source")
	}
	report.Steps = append(report.Steps, validationStep{Name: planned[2], Status: "passed", Detail: fmt.Sprintf("bytes=%d checksum=%s", downloadResult.Bytes, downloadResult.Checksum)})

	resumeDestination := filepath.Join(tempDir, "resume.bin")
	resumeResult, resumeErr := client.ValidateResumableDownload(ctx, upload.Remote, resumeDestination, int64(payloadSize/3))
	if resumeErr != nil {
		report.Steps = append(report.Steps, validationStep{Name: planned[3], Status: "failed", Detail: "checkpoint continuation was not verified"})
		return report, resumeErr
	}
	report.Steps = append(report.Steps, validationStep{Name: planned[3], Status: "passed", Detail: fmt.Sprintf("bytes=%d checksum=%s resumed=%t", resumeResult.Bytes, resumeResult.Checksum, resumeResult.Resumed)})
	return report, nil
}

func validationPayload(size int) []byte {
	payload := make([]byte, size)
	for index := range payload {
		payload[index] = byte((index*31 + 17) % 251)
	}
	return payload
}

func writeEvidence(path string, report validationReport, token string) error {
	var builder strings.Builder
	fmt.Fprintf(&builder, "# Zenodo sandbox transfer evidence\n\n")
	fmt.Fprintf(&builder, "- Generated: %s\n", report.GeneratedAt.Format(time.RFC3339))
	fmt.Fprintf(&builder, "- Mode: %s\n", report.Mode)
	fmt.Fprintf(&builder, "- Base host: %s\n", report.BaseHost)
	fmt.Fprintf(&builder, "- ZENODO_TOKEN: %s\n", setState(report.TokenSet))
	if report.SkipReason != "" {
		fmt.Fprintf(&builder, "- Boundary: %s\n", report.SkipReason)
	}
	builder.WriteString("\n## Checks\n\n")
	for _, step := range report.Steps {
		detail := exactRedact(step.Detail, token)
		fmt.Fprintf(&builder, "- %s: %s", step.Name, step.Status)
		if detail != "" {
			fmt.Fprintf(&builder, " (%s)", detail)
		}
		builder.WriteByte('\n')
	}
	content := []byte(exactRedact(builder.String(), token))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".zenodo-evidence-*.tmp")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	defer func() { _ = os.Remove(tempName) }()
	if err := temp.Chmod(0o644); err != nil {
		_ = temp.Close()
		return err
	}
	if _, err := temp.Write(content); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(tempName, path)
}

func safeHost(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		return "invalid"
	}
	return parsed.Host
}

func setState(set bool) string {
	if set {
		return "set"
	}
	return "unset"
}

func exactRedact(text, secret string) string {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return text
	}
	return strings.ReplaceAll(text, secret, "[REDACTED]")
}

func stepStatus(report validationReport, name string) string {
	for _, step := range report.Steps {
		if step.Name == name {
			return step.Status
		}
	}
	return "missing"
}
