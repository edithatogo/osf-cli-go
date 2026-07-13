// Command checkapischema validates the pinned official OSF API schema manifest.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type sourceManifest struct {
	SchemaVersion    int      `json:"schemaVersion"`
	SourceRepository string   `json:"sourceRepository"`
	SourcePath       string   `json:"sourcePath"`
	SourceRef        string   `json:"sourceRef"`
	SourceCommit     string   `json:"sourceCommit"`
	SourceURL        string   `json:"sourceUrl"`
	RetrievedDate    string   `json:"retrievedDate"`
	License          string   `json:"license"`
	Handling         string   `json:"handling"`
	Decision         string   `json:"decision"`
	DeferIssue       string   `json:"deferIssue"`
	TrackedTags      []string `json:"trackedTags"`
	Rationale        string   `json:"rationale"`
}

var commitPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

func main() {
	online := flag.Bool("online", false, "fetch and validate the pinned source URL")
	flag.Parse()
	manifest, err := loadManifest("docs/osf-api-schema-source.json")
	if err == nil {
		err = validateManifest(manifest)
	}
	if err == nil && *online {
		err = validateOnline(manifest)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "checkapischema: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("OSF API schema manifest: valid")
}

func loadManifest(path string) (sourceManifest, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return sourceManifest{}, err
	}
	var manifest sourceManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return sourceManifest{}, err
	}
	return manifest, nil
}

func validateManifest(manifest sourceManifest) error {
	if manifest.SchemaVersion != 1 {
		return fmt.Errorf("schemaVersion = %d, want 1", manifest.SchemaVersion)
	}
	if manifest.SourceRepository != "https://github.com/CenterForOpenScience/developer.osf.io" {
		return errors.New("sourceRepository must be the official OSF developer repository")
	}
	if manifest.SourcePath != "swagger-spec/swagger.yaml" || manifest.SourceRef != "master" {
		return errors.New("sourcePath/sourceRef do not identify the official Swagger source")
	}
	if !commitPattern.MatchString(manifest.SourceCommit) {
		return errors.New("sourceCommit must be a lowercase 40-character commit SHA")
	}
	wantURL := "https://raw.githubusercontent.com/CenterForOpenScience/developer.osf.io/" + manifest.SourceCommit + "/" + manifest.SourcePath
	if manifest.SourceURL != wantURL {
		return fmt.Errorf("sourceUrl = %q, want pinned raw source URL", manifest.SourceURL)
	}
	if _, err := time.Parse("2006-01-02", manifest.RetrievedDate); err != nil {
		return fmt.Errorf("retrievedDate: %w", err)
	}
	if manifest.License != "CC BY-NC 4.0" || manifest.Handling != "remote-source-manifest" {
		return errors.New("license or handling policy is invalid")
	}
	if manifest.Decision != "deferred" || manifest.DeferIssue != "#46" || strings.TrimSpace(manifest.Rationale) == "" {
		return errors.New("deferred decision must include issue #46 and rationale")
	}
	if len(manifest.TrackedTags) == 0 {
		return errors.New("trackedTags must not be empty")
	}
	return nil
}

func validateOnline(manifest sourceManifest) error {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(manifest.SourceURL)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("source returned HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return err
	}
	if len(body) == 0 || !strings.Contains(string(body), "swagger:") {
		return errors.New("source is empty or does not look like Swagger YAML")
	}
	return nil
}
