// Command checkreleasecontract validates the repository's release-readiness contract.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type manifest struct {
	Version string `json:"version"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "release contract: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	required := []string{
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
