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
		".github/plugin/marketplace.json",
		"gemini-extension.json",
		"docs/compatibility-policy.md",
		"docs/support-policy.md",
		"docs/live-validation-matrix.md",
		"docs/release-candidate-evidence.md",
		"docs/v1-launch-roadmap.md",
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
		return fmt.Errorf("Copilot marketplace name or version is invalid")
	}
	if len(copilotMarketplace.Plugins) != 1 {
		return fmt.Errorf("Copilot marketplace must contain exactly one plugin")
	}
	copilotPlugin := copilotMarketplace.Plugins[0]
	if copilotPlugin.Name == "" || copilotPlugin.Source == "" || copilotPlugin.Version != server.Version {
		return fmt.Errorf("Copilot marketplace plugin name, source, or version is invalid")
	}
	if _, err := os.Stat(filepath.Join(copilotPlugin.Source, "plugin.json")); err != nil {
		return fmt.Errorf("Copilot marketplace plugin source %s is invalid: %w", copilotPlugin.Source, err)
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
