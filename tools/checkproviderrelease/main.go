// Command checkproviderrelease validates dated, digest-bound provider claims.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const defaultManifest = "docs/multi-provider-validation.json"

type manifest struct {
	SchemaVersion int     `json:"schemaVersion"`
	GeneratedAt   string  `json:"generatedAt"`
	OptInWorkflow string  `json:"optInWorkflow"`
	Claims        []claim `json:"claims"`
}

type claim struct {
	ID                  string     `json:"id"`
	Provider            string     `json:"provider"`
	Capability          string     `json:"capability"`
	Level               string     `json:"level"`
	ValidatedAt         string     `json:"validatedAt"`
	Evidence            []evidence `json:"evidence"`
	ProductionRecord    string     `json:"productionRecord,omitempty"`
	ResourceDisposition string     `json:"resourceDisposition,omitempty"`
	ResourceRecord      string     `json:"resourceRecord,omitempty"`
}

type evidence struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

func main() {
	manifestPath := flag.String("manifest", defaultManifest, "provider validation manifest")
	root := flag.String("root", ".", "repository root")
	report := flag.String("report", "", "write a reproducible Markdown validation report")
	flag.Parse()
	if err := run(*root, *manifestPath, time.Now().UTC()); err != nil {
		fmt.Fprintf(os.Stderr, "provider release contract: %v\n", err)
		os.Exit(1)
	}
	if *report != "" {
		if err := writeReport(*root, *manifestPath, *report); err != nil {
			fmt.Fprintf(os.Stderr, "provider release report: %v\n", err)
			os.Exit(1)
		}
	}
}

func run(root, manifestPath string, now time.Time) error {
	root, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	manifestFile, err := confinedPath(root, manifestPath)
	if err != nil {
		return err
	}
	payload, err := os.ReadFile(manifestFile)
	if err != nil {
		return fmt.Errorf("read manifest: %w", err)
	}
	var document manifest
	decoder := json.NewDecoder(strings.NewReader(string(payload)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&document); err != nil {
		return fmt.Errorf("decode manifest: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("manifest must contain exactly one JSON document")
	}
	if document.SchemaVersion != 1 || len(document.Claims) == 0 {
		return errors.New("schema version 1 and at least one claim are required")
	}
	if err := validateOptInWorkflow(root, document.OptInWorkflow); err != nil {
		return err
	}
	generatedAt, err := time.Parse(time.RFC3339, document.GeneratedAt)
	if err != nil || generatedAt.After(now.Add(time.Minute)) || generatedAt.Before(now.Add(-30*24*time.Hour)) {
		return errors.New("generatedAt must be a recent, non-future RFC3339 timestamp")
	}
	seen := make(map[string]bool, len(document.Claims))
	seenCapabilities := make(map[string]bool, len(document.Claims))
	providers := map[string]bool{"osf": true, "zenodo": true, "cross-provider": true}
	levels := map[string]bool{"offline-tested": true, "sandbox-validated": true, "production-validated": true}
	for _, claim := range document.Claims {
		if claim.ID == "" || seen[claim.ID] {
			return fmt.Errorf("claim id %q is empty or duplicated", claim.ID)
		}
		seen[claim.ID] = true
		if !providers[claim.Provider] || !levels[claim.Level] || strings.TrimSpace(claim.Capability) == "" {
			return fmt.Errorf("claim %s has an invalid provider, capability, or validation level", claim.ID)
		}
		capabilityKey := claim.Provider + "\x00" + claim.Capability
		if seenCapabilities[capabilityKey] {
			return fmt.Errorf("claim %s duplicates provider capability %s", claim.ID, claim.Capability)
		}
		seenCapabilities[capabilityKey] = true
		validatedAt, err := time.Parse("2006-01-02", claim.ValidatedAt)
		if err != nil || validatedAt.After(now.Add(24*time.Hour)) || validatedAt.Before(now.Add(-90*24*time.Hour)) {
			return fmt.Errorf("claim %s has an invalid, stale, or future validation date", claim.ID)
		}
		if len(claim.Evidence) == 0 {
			return fmt.Errorf("claim %s has no evidence", claim.ID)
		}
		if claim.Level == "production-validated" {
			if err := validateProductionRecord(claim.ProductionRecord); err != nil {
				return fmt.Errorf("claim %s: %w", claim.ID, err)
			}
		} else if claim.ProductionRecord != "" {
			return fmt.Errorf("claim %s cannot attach a production record to level %s", claim.ID, claim.Level)
		}
		if err := validateResourceDisposition(claim); err != nil {
			return fmt.Errorf("claim %s: %w", claim.ID, err)
		}
		for _, item := range claim.Evidence {
			if err := validateEvidence(root, claim, item); err != nil {
				return fmt.Errorf("claim %s: %w", claim.ID, err)
			}
		}
	}
	return nil
}

func validateOptInWorkflow(root, name string) error {
	filename, err := confinedPath(root, name)
	if err != nil {
		return fmt.Errorf("opt-in workflow: %w", err)
	}
	payload, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("read opt-in workflow: %w", err)
	}
	content := string(payload)
	if !strings.Contains(content, "workflow_dispatch:") {
		return errors.New("opt-in workflow must use workflow_dispatch")
	}
	for _, forbidden := range []string{"pull_request:", "push:", "schedule:"} {
		if strings.Contains(content, forbidden) {
			return fmt.Errorf("opt-in workflow cannot declare %s", forbidden)
		}
	}
	for _, required := range []string{
		"ZENODO_SANDBOX_VALIDATION", "ZENODO_PUBLICATION_VALIDATION", "CROSS_PROVIDER_SANDBOX_VALIDATION", "OSF_LIVE_VALIDATION",
		"OSF_VALIDATE_WRITES",
		"ZENODO_SANDBOX_TOKEN", "ZENODO_SANDBOX_PUBLICATION_TOKEN", "OSF_VALIDATION_TOKEN", "OSF_VALIDATE_PROJECT",
	} {
		if !strings.Contains(content, required) {
			return fmt.Errorf("opt-in workflow is missing %s", required)
		}
	}
	return nil
}

func writeReport(root, manifestPath, reportPath string) error {
	root, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	manifestFile, err := confinedPath(root, manifestPath)
	if err != nil {
		return err
	}
	payload, err := os.ReadFile(manifestFile)
	if err != nil {
		return err
	}
	var document manifest
	if err := json.Unmarshal(payload, &document); err != nil {
		return err
	}
	var builder strings.Builder
	builder.WriteString("# Multi-provider validation report\n\n")
	fmt.Fprintf(&builder, "- Schema: %d\n- Generated: %s\n- Opt-in workflow: `%s`\n- Production-validated claims: %d\n\n", document.SchemaVersion, document.GeneratedAt, document.OptInWorkflow, countLevel(document.Claims, "production-validated"))
	builder.WriteString("| Provider | Capability | Level | Validated | Resource disposition | Evidence |\n|---|---|---|---|---|---|\n")
	for _, claim := range document.Claims {
		paths := make([]string, 0, len(claim.Evidence))
		for _, item := range claim.Evidence {
			paths = append(paths, "`"+item.Path+"` ("+shortDigest(item.SHA256)+")")
		}
		disposition := claim.ResourceDisposition
		if disposition == "" {
			disposition = "not-applicable"
		}
		if claim.ResourceRecord != "" {
			disposition += " (" + claim.ResourceRecord + ")"
		}
		fmt.Fprintf(&builder, "| %s | %s | %s | %s | %s | %s |\n", claim.Provider, claim.Capability, claim.Level, claim.ValidatedAt, disposition, strings.Join(paths, "; "))
	}
	filename, err := confinedPath(root, reportPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		return err
	}
	return os.WriteFile(filename, []byte(builder.String()), 0o644)
}

func countLevel(claims []claim, level string) int {
	count := 0
	for _, claim := range claims {
		if claim.Level == level {
			count++
		}
	}
	return count
}

func shortDigest(value string) string {
	value = strings.TrimPrefix(value, "sha256:")
	if len(value) > 12 {
		value = value[:12]
	}
	return "sha256:" + value
}

func validateEvidence(root string, claim claim, item evidence) error {
	filename, err := confinedPath(root, item.Path)
	if err != nil {
		return err
	}
	payload, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("read evidence %s: %w", item.Path, err)
	}
	digest := sha256.Sum256(payload)
	want := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(item.SHA256)), "sha256:")
	if len(want) != 64 || hex.EncodeToString(digest[:]) != want {
		return fmt.Errorf("evidence %s digest does not match", item.Path)
	}
	if claim.Level == "sandbox-validated" && !strings.Contains(strings.ToLower(string(payload)), "sandbox") {
		return fmt.Errorf("sandbox evidence %s does not identify a sandbox", item.Path)
	}
	return nil
}

func confinedPath(root, name string) (string, error) {
	if filepath.IsAbs(name) || strings.TrimSpace(name) == "" {
		return "", errors.New("evidence paths must be non-empty and repository-relative")
	}
	filename := filepath.Join(root, filepath.Clean(name))
	relative, err := filepath.Rel(root, filename)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", errors.New("evidence path escapes repository root")
	}
	return filename, nil
}

func validateProductionRecord(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "https" || parsed.Hostname() == "" || parsed.User != nil {
		return errors.New("production-validated claims require a public HTTPS productionRecord")
	}
	host := strings.ToLower(parsed.Hostname())
	if strings.Contains(host, "sandbox") || host == "localhost" || host == "127.0.0.1" {
		return errors.New("productionRecord cannot target a sandbox or loopback host")
	}
	return nil
}

func validateResourceDisposition(value claim) error {
	if value.Level != "sandbox-validated" {
		if value.ResourceDisposition != "" || value.ResourceRecord != "" {
			return errors.New("resource disposition is only valid for sandbox claims")
		}
		return nil
	}
	switch value.ResourceDisposition {
	case "deleted":
		if value.ResourceRecord != "" {
			return errors.New("deleted sandbox resources cannot retain a resource record")
		}
	case "published-retained":
		parsed, err := url.Parse(value.ResourceRecord)
		if err != nil || parsed.Scheme != "https" || parsed.Hostname() != "sandbox.zenodo.org" || parsed.User != nil {
			return errors.New("retained sandbox publications require a public sandbox.zenodo.org resource record")
		}
	default:
		return errors.New("sandbox claims require resourceDisposition deleted or published-retained")
	}
	return nil
}
