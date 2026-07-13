// Command checkfeaturematrix validates and renders the canonical feature matrix.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type matrix struct {
	Version           string            `json:"version"`
	Reviewed          string            `json:"reviewed"`
	StatusDefinitions map[string]string `json:"status_definitions"`
	Rows              []row             `json:"rows"`
}

type row struct {
	Area     string `json:"area"`
	CLI      string `json:"cli"`
	API      string `json:"api"`
	MCP      string `json:"mcp"`
	Contract string `json:"contract"`
	Status   string `json:"status"`
	Next     string `json:"next,omitempty"`
	Track    string `json:"track,omitempty"`
	Issue    string `json:"issue,omitempty"`
}

func main() {
	write := flag.Bool("write", false, "write the generated Markdown presentation")
	flag.Parse()
	if err := run(*write); err != nil {
		fmt.Fprintln(os.Stderr, "feature matrix:", err)
		os.Exit(1)
	}
}

func run(write bool) error {
	var m matrix
	if err := readJSON("docs/feature-matrix.json", &m); err != nil {
		return err
	}
	if err := validate(m); err != nil {
		return err
	}
	generated, err := render(m)
	if err != nil {
		return err
	}
	path := "docs/feature-matrix.md"
	if write {
		return os.WriteFile(path, generated, 0o644)
	}
	actual, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if !bytes.Equal(actual, generated) {
		return errors.New("docs/feature-matrix.md is stale; run go run ./tools/checkfeaturematrix -write")
	}
	return nil
}

func validate(m matrix) error {
	if strings.TrimSpace(m.Version) == "" || strings.TrimSpace(m.Reviewed) == "" {
		return errors.New("version and reviewed are required")
	}
	if _, err := time.Parse("2006-01-02", m.Reviewed); err != nil {
		return fmt.Errorf("reviewed must be YYYY-MM-DD: %w", err)
	}
	var server struct {
		Version string `json:"version"`
	}
	if err := readJSON("server.json", &server); err != nil {
		return err
	}
	if m.Version != server.Version {
		return fmt.Errorf("matrix version %q does not match server version %q", m.Version, server.Version)
	}
	allowed := map[string]bool{"implemented": true, "prepared": true, "track": true, "external-gate": true}
	seen := map[string]bool{}
	for i, item := range m.Rows {
		if strings.TrimSpace(item.Area) == "" || seen[item.Area] {
			return fmt.Errorf("row %d has a missing or duplicate area", i+1)
		}
		seen[item.Area] = true
		for name, value := range map[string]string{"cli": item.CLI, "api": item.API, "mcp": item.MCP, "contract": item.Contract, "status": item.Status} {
			if strings.TrimSpace(value) == "" {
				return fmt.Errorf("row %q is missing %s", item.Area, name)
			}
		}
		if !allowed[item.Status] {
			return fmt.Errorf("row %q has invalid status %q", item.Area, item.Status)
		}
		if item.Status == "track" && (strings.TrimSpace(item.Next) == "" || strings.TrimSpace(item.Track) == "") {
			return fmt.Errorf("track row %q requires next and track", item.Area)
		}
		if item.Track != "" {
			if _, err := os.Stat(item.Track); err != nil {
				return fmt.Errorf("row %q references missing track %q: %w", item.Area, item.Track, err)
			}
		}
	}
	return nil
}

func render(m matrix) ([]byte, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "# OSF CLI Go feature matrix\n\nLast reviewed: %s\n\n", m.Reviewed)
	b.WriteString("This generated matrix is backed by `docs/feature-matrix.json`. Status meanings: ")
	statuses := make([]string, 0, len(m.StatusDefinitions))
	for key := range m.StatusDefinitions {
		statuses = append(statuses, key)
	}
	sort.Strings(statuses)
	for i, key := range statuses {
		if i > 0 {
			b.WriteString("; ")
		}
		fmt.Fprintf(&b, "**%s** is %s", key, m.StatusDefinitions[key])
	}
	b.WriteString("\n\n| Area | CLI | API client | MCP | Safety/quality contract | Status | Next action | Track | Issue |\n|---|---|---|---|---|---|---|---|---|\n")
	for _, item := range m.Rows {
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %s | %s | `%s` | %s |\n", item.Area, item.CLI, item.API, item.MCP, item.Contract, item.Status, item.Next, item.Track, item.Issue)
	}
	b.WriteString("\n## Matrix rules\n\n1. A row cannot be marked implemented from documentation alone; it needs code and deterministic validation.\n2. A registry cannot be marked published from a local packet; provider-side receipt or public listing evidence is required.\n3. A feature that performs a write must document authorization, confirmation, rollback, and live-test cleanup before MCP exposure.\n4. Every release updates this matrix, the feature inventory, and the registry scorecard together.\n")
	return []byte(b.String()), nil
}

func readJSON(path string, target any) error {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}
