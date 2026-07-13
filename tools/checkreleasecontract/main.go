// Command checkreleasecontract validates the repository's release-readiness contract.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type manifest struct {
	Version string `json:"version"`
}

type qualityReport struct {
	SchemaVersion int    `json:"schemaVersion"`
	Version       string `json:"version"`
	Summary       struct {
		Failed int    `json:"failed"`
		Status string `json:"status"`
	} `json:"summary"`
}

type marketplace struct {
	Name     string `json:"name"`
	Metadata struct {
		Version string `json:"version"`
	} `json:"metadata"`
	Plugins []struct {
		Name    string `json:"name"`
		Source  string `json:"source"`
		Version string `json:"version"`
	} `json:"plugins"`
}

type codexMarketplace struct {
	Name    string `json:"name"`
	Plugins []struct {
		Name   string `json:"name"`
		Source struct {
			Source string `json:"source"`
			Path   string `json:"path"`
		} `json:"source"`
	} `json:"plugins"`
}

type claudeMarketplace struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Plugins []struct {
		Name    string `json:"name"`
		Source  string `json:"source"`
		Version string `json:"version"`
	} `json:"plugins"`
}

type geminiManifest struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Settings []struct {
		Name   string `json:"name"`
		EnvVar string `json:"envVar"`
		Secret bool   `json:"sensitive"`
	} `json:"settings"`
	MCPServers map[string]struct {
		Command string `json:"command"`
	} `json:"mcpServers"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "release contract: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	required := []string{
		".plugin/plugin.json",
		".github/plugin/marketplace.json",
		"gemini-extension.json",
		"qwen-extension.json",
		".cursor/mcp.json",
		".roo/mcp.json",
		".vscode/mcp.json",
		"integrations/cline/cline_mcp_settings.json",
		"integrations/windsurf/mcp_config.json",
		"integrations/zed/settings.json",
		"docs/compatibility-policy.md",
		"docs/cli-json-contract.md",
		"docs/mcp-schema-contract.md",
		"docs/migration-v1.md",
		"docs/threat-model.md",
		"docs/operations-runbook.md",
		"docs/support-policy.md",
		"docs/live-validation-matrix.md",
		"docs/release-candidate-evidence.md",
		"docs/v1-launch-roadmap.md",
		"docs/v1-launch-review.md",
		"docs/mcp-quality-report.json",
	}
	for _, path := range required {
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("required readiness artifact %s: %w", path, err)
		}
	}

	var server manifest
	if err := readJSON("server.json", &server); err != nil {
		return err
	}
	var openPlugin manifest
	if err := readJSON(".plugin/plugin.json", &openPlugin); err != nil {
		return err
	}
	if openPlugin.Version != server.Version {
		return fmt.Errorf(".plugin/plugin.json version %q does not match server version %q", openPlugin.Version, server.Version)
	}
	var quality qualityReport
	if err := readJSON("docs/mcp-quality-report.json", &quality); err != nil {
		return err
	}
	if quality.SchemaVersion != 1 || quality.Version != server.Version || quality.Summary.Status != "passed" || quality.Summary.Failed != 0 {
		return fmt.Errorf("MCP quality report is not a passing report for server version %q", server.Version)
	}
	pluginPaths := []string{
		"plugins/codex-osf/.codex-plugin/plugin.json",
		"plugins/claude-osf/.claude-plugin/plugin.json",
		"plugins/github-copilot-osf/plugin.json",
		"plugins/gemini-osf/gemini-extension.json",
		"plugins/qwen-osf/qwen-extension.json",
	}
	for _, path := range pluginPaths {
		var plugin manifest
		if err := readJSON(path, &plugin); err != nil {
			return err
		}
		if plugin.Version != server.Version {
			return fmt.Errorf("%s version %q does not match server version %q", path, plugin.Version, server.Version)
		}
	}

	var copilotMarketplace marketplace
	if err := readJSON(".github/plugin/marketplace.json", &copilotMarketplace); err != nil {
		return err
	}
	if copilotMarketplace.Name == "" || copilotMarketplace.Metadata.Version != server.Version {
		return fmt.Errorf("copilot marketplace name or version is invalid")
	}
	if len(copilotMarketplace.Plugins) != 1 {
		return fmt.Errorf("copilot marketplace must contain exactly one plugin")
	}
	copilotPlugin := copilotMarketplace.Plugins[0]
	if copilotPlugin.Name == "" || copilotPlugin.Source == "" || copilotPlugin.Version != server.Version {
		return fmt.Errorf("copilot marketplace plugin name, source, or version is invalid")
	}
	if _, err := os.Stat(filepath.Join(copilotPlugin.Source, "plugin.json")); err != nil {
		return fmt.Errorf("copilot marketplace plugin source %s is invalid: %w", copilotPlugin.Source, err)
	}

	var codex codexMarketplace
	if err := readJSON(".agents/plugins/marketplace.json", &codex); err != nil {
		return err
	}
	if codex.Name == "" || len(codex.Plugins) != 1 {
		return fmt.Errorf("codex marketplace name or plugin count is invalid")
	}
	codexPlugin := codex.Plugins[0]
	if codexPlugin.Name == "" || codexPlugin.Source.Source != "local" || codexPlugin.Source.Path == "" {
		return fmt.Errorf("codex marketplace plugin source is invalid")
	}
	codexPath := filepath.Clean(codexPlugin.Source.Path)
	if _, err := os.Stat(filepath.Join(codexPath, ".codex-plugin/plugin.json")); err != nil {
		return fmt.Errorf("codex marketplace plugin source %s is invalid: %w", codexPlugin.Source.Path, err)
	}

	var claudeCatalog claudeMarketplace
	if err := readJSON(".agents/plugins/.claude-plugin/marketplace.json", &claudeCatalog); err != nil {
		return err
	}
	if claudeCatalog.Name == "" || claudeCatalog.Version != server.Version || len(claudeCatalog.Plugins) != 1 {
		return fmt.Errorf("claude-compatible marketplace name, version, or plugin count is invalid")
	}
	claudePlugin := claudeCatalog.Plugins[0]
	claudePath := filepath.Clean(filepath.Join(".agents/plugins/.claude-plugin", claudePlugin.Source))
	if claudePlugin.Name == "" || claudePlugin.Version != server.Version {
		return fmt.Errorf("claude-compatible marketplace plugin name or version is invalid")
	}
	if _, err := os.Stat(filepath.Join(claudePath, ".codex-plugin/plugin.json")); err != nil {
		return fmt.Errorf("claude-compatible marketplace plugin source %s is invalid: %w", claudePlugin.Source, err)
	}

	var rootGemini geminiManifest
	if err := readJSON("gemini-extension.json", &rootGemini); err != nil {
		return err
	}
	if err := validateGeminiManifest("gemini-extension.json", rootGemini, server.Version, "go"); err != nil {
		return err
	}
	var packagedGemini geminiManifest
	if err := readJSON("plugins/gemini-osf/gemini-extension.json", &packagedGemini); err != nil {
		return err
	}
	if err := validateGeminiManifest("plugins/gemini-osf/gemini-extension.json", packagedGemini, server.Version, "${extensionPath}"); err != nil {
		return err
	}
	var rootQwen geminiManifest
	if err := readJSON("qwen-extension.json", &rootQwen); err != nil {
		return err
	}
	if err := validateGeminiManifest("qwen-extension.json", rootQwen, server.Version, "go"); err != nil {
		return err
	}
	var packagedQwen geminiManifest
	if err := readJSON("plugins/qwen-osf/qwen-extension.json", &packagedQwen); err != nil {
		return err
	}
	if err := validateGeminiManifest("plugins/qwen-osf/qwen-extension.json", packagedQwen, server.Version, "${extensionPath}"); err != nil {
		return err
	}
	for path, key := range map[string]string{
		".cursor/mcp.json": "mcpServers",
		".roo/mcp.json":    "mcpServers",
		".vscode/mcp.json": "servers",
		"integrations/cline/cline_mcp_settings.json": "mcpServers",
		"integrations/windsurf/mcp_config.json":      "mcpServers",
		"integrations/zed/settings.json":             "context_servers",
	} {
		if err := validateIntegrationConfig(path, key); err != nil {
			return err
		}
	}

	return nil
}

func validateIntegrationConfig(path, key string) error {
	var config map[string]map[string]struct {
		Command string            `json:"command"`
		Args    []string          `json:"args"`
		Env     map[string]string `json:"env"`
	}
	if err := readJSON(path, &config); err != nil {
		return err
	}
	servers, ok := config[key]
	if !ok {
		return fmt.Errorf("%s is missing %s", path, key)
	}
	osf, ok := servers["osf"]
	if !ok || osf.Command != "go" || len(osf.Args) != 2 || osf.Args[0] != "run" || osf.Args[1] != "./cmd/osf-mcp" {
		return fmt.Errorf("%s has an invalid osf server command", path)
	}
	requiredEnv := map[string]bool{
		"OSF_TOKEN":    false,
		"OSF_USERNAME": false,
		"OSF_PASSWORD": false,
	}
	for name, value := range osf.Env {
		if name != "OSF_TOKEN" && name != "OSF_USERNAME" && name != "OSF_PASSWORD" {
			return fmt.Errorf("%s contains an unexpected environment variable %s", path, name)
		}
		if value != "${env:"+name+"}" {
			return fmt.Errorf("%s does not reference %s through the environment", path, name)
		}
		requiredEnv[name] = true
	}
	for name, present := range requiredEnv {
		if !present {
			return fmt.Errorf("%s is missing environment variable %s", path, name)
		}
	}
	return nil
}

func validateGeminiManifest(path string, manifest geminiManifest, version string, command string) error {
	if manifest.Name != "osf-cli-go" || manifest.Version != version {
		return fmt.Errorf("%s name or version is invalid", path)
	}
	if len(manifest.Settings) != 3 {
		return fmt.Errorf("%s must declare three credential settings", path)
	}
	seen := map[string]bool{}
	for _, setting := range manifest.Settings {
		if setting.Name == "" || setting.EnvVar == "" || !setting.Secret {
			return fmt.Errorf("%s contains an invalid credential setting", path)
		}
		seen[setting.EnvVar] = true
	}
	for _, envVar := range []string{"OSF_TOKEN", "OSF_USERNAME", "OSF_PASSWORD"} {
		if !seen[envVar] {
			return fmt.Errorf("%s does not allowlist %s", path, envVar)
		}
	}
	server, ok := manifest.MCPServers["osf"]
	if !ok || (server.Command != command && !strings.HasPrefix(server.Command, command)) {
		return fmt.Errorf("%s has an invalid osf MCP server command", path)
	}
	return nil
}

func readJSON(path string, target any) error {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	return nil
}
