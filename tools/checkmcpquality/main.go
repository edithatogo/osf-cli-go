// Command checkmcpquality evaluates the offline MCP quality and compatibility contract.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var expectedTools = []string{
	"osf_whoami", "osf_projects_list", "osf_project_get", "osf_components_list",
	"osf_files_list", "osf_contributors_list", "osf_search", "osf_preprints_list",
	"osf_preprints_search", "osf_doi_resolve",
}

type qualityReport struct {
	SchemaVersion int            `json:"schemaVersion"`
	GeneratedDate string         `json:"generatedDate"`
	Version       string         `json:"version"`
	Mode          string         `json:"mode"`
	Summary       qualitySummary `json:"summary"`
	Checks        []qualityCheck `json:"checks"`
}

type qualitySummary struct {
	Passed  int    `json:"passed"`
	Failed  int    `json:"failed"`
	Skipped int    `json:"skipped"`
	Status  string `json:"status"`
}

type qualityCheck struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Detail string `json:"detail"`
}

type serverMetadata struct {
	Version string `json:"version"`
}

type packageManifest struct {
	Version string `json:"version"`
	Tools   []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"tools"`
	Server struct {
		MCPConfig struct {
			Env map[string]string `json:"env"`
		} `json:"mcp_config"`
	} `json:"server"`
	UserConfig map[string]struct {
		Sensitive bool `json:"sensitive"`
	} `json:"user_config"`
}

func main() {
	output := flag.String("output", "docs/mcp-quality-report.json", "report output path")
	live := flag.Bool("live", false, "run opt-in live OSF validation when credentials and project are configured")
	flag.Parse()

	report, err := evaluate(*live)
	if writeErr := writeReport(*output, report); writeErr != nil {
		fmt.Fprintf(os.Stderr, "checkmcpquality: write report: %v\n", writeErr)
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "checkmcpquality: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("MCP quality report: %s (%d passed, %d skipped)\n", report.Summary.Status, report.Summary.Passed, report.Summary.Skipped)
}

func evaluate(live bool) (qualityReport, error) {
	report := qualityReport{
		SchemaVersion: 1,
		GeneratedDate: time.Now().UTC().Format("2006-01-02"),
		Mode:          "offline",
	}
	var server serverMetadata
	if err := readJSON("server.json", &server); err != nil {
		return report, err
	}
	report.Version = server.Version
	checks := []qualityCheck{}
	add := func(name string, err error) {
		check := qualityCheck{Name: name, Status: "passed", Detail: "validated"}
		if err != nil {
			check.Status = "failed"
			check.Detail = err.Error()
		}
		checks = append(checks, check)
	}

	add("server-metadata", checkServerMetadata(server))
	var manifest packageManifest
	if err := readJSON("packaging/mcpb/manifest.json", &manifest); err != nil {
		add("mcpb-manifest", err)
	} else {
		add("mcpb-manifest", checkPackageManifest(manifest, server.Version))
	}
	add("registry-tool-inventory", checkRegistryToolInventory())
	add("client-configurations", checkClientConfigurations())
	add("security-and-bounds-source", checkSecurityAndBoundsSource())
	add("mcp-server-tests", runCommand("go", "test", "./internal/mcpserver"))
	if live {
		report.Mode = "offline-and-live"
		add("live-osf-validation", runCommand("go", "run", "./tools/livevalidation", "-live"))
	} else {
		checks = append(checks, qualityCheck{Name: "live-osf-validation", Status: "skipped", Detail: "opt-in; rerun with -live and OSF_LIVE_VALIDATION=1 plus an OSF project reference"})
	}
	report.Checks = checks
	for _, check := range checks {
		switch check.Status {
		case "passed":
			report.Summary.Passed++
		case "failed":
			report.Summary.Failed++
		case "skipped":
			report.Summary.Skipped++
		}
	}
	report.Summary.Status = "passed"
	if report.Summary.Failed > 0 {
		report.Summary.Status = "failed"
		return report, errors.New("one or more MCP quality checks failed")
	}
	return report, nil
}

func checkServerMetadata(server serverMetadata) error {
	if strings.TrimSpace(server.Version) == "" {
		return errors.New("server version is required")
	}
	return nil
}

func checkPackageManifest(manifest packageManifest, version string) error {
	if manifest.Version != version {
		return fmt.Errorf("MCPB version %q does not match server version %q", manifest.Version, version)
	}
	seen := map[string]bool{}
	for _, tool := range manifest.Tools {
		if !contains(expectedTools, tool.Name) || seen[tool.Name] {
			return fmt.Errorf("MCPB tool inventory contains unexpected or duplicate tool %q", tool.Name)
		}
		if strings.TrimSpace(tool.Description) == "" {
			return fmt.Errorf("MCPB tool %q has no description", tool.Name)
		}
		seen[tool.Name] = true
	}
	for _, name := range expectedTools {
		if !seen[name] {
			return fmt.Errorf("MCPB tool inventory is missing %q", name)
		}
	}
	for _, name := range []string{"osf_token", "osf_username", "osf_password"} {
		config, ok := manifest.UserConfig[name]
		if !ok || !config.Sensitive {
			return fmt.Errorf("MCPB user config %q must exist and be sensitive", name)
		}
	}
	return nil
}

func checkRegistryToolInventory() error {
	var packet struct {
		Capabilities struct {
			Tools []string `json:"tools"`
		} `json:"capabilities"`
	}
	if err := readJSON("registry/directory-submissions.json", &packet); err != nil {
		return err
	}
	if !sameNames(packet.Capabilities.Tools, expectedTools) {
		return fmt.Errorf("directory tool inventory does not match MCP contract")
	}
	return nil
}

func checkClientConfigurations() error {
	configs := map[string]string{
		".github/mcp.json": "mcpServers", ".cursor/mcp.json": "mcpServers", ".roo/mcp.json": "mcpServers",
		".vscode/mcp.json": "servers", "integrations/cline/cline_mcp_settings.json": "mcpServers",
		"integrations/windsurf/mcp_config.json": "mcpServers", "integrations/zed/settings.json": "context_servers",
	}
	for path, key := range configs {
		var value map[string]json.RawMessage
		if err := readJSON(path, &value); err != nil {
			return err
		}
		var entries map[string]json.RawMessage
		data, ok := value[key]
		if !ok || json.Unmarshal(data, &entries) != nil || len(entries) == 0 {
			return fmt.Errorf("%s must define a non-empty %s map", path, key)
		}
	}
	return nil
}

func checkSecurityAndBoundsSource() error {
	data, err := os.ReadFile("internal/mcpserver/server.go")
	if err != nil {
		return err
	}
	text := string(data)
	for _, marker := range []string{"mcpError", "boundedLimit", "boundedSearchLimit"} {
		if !strings.Contains(text, marker) {
			return fmt.Errorf("MCP server source missing safety marker %q", marker)
		}
	}
	return nil
}

func runCommand(name string, args ...string) error {
	command := exec.Command(name, args...)
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return nil
}

func writeReport(path string, report qualityReport) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(path, data, 0o644)
}

func readJSON(path string, target any) error {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func sameNames(got, want []string) bool {
	gotCopy, wantCopy := append([]string(nil), got...), append([]string(nil), want...)
	sort.Strings(gotCopy)
	sort.Strings(wantCopy)
	return bytes.Equal(mustJSON(gotCopy), mustJSON(wantCopy))
}

func mustJSON(value any) []byte { data, _ := json.Marshal(value); return data }
