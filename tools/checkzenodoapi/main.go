// Command checkzenodoapi validates the pinned Zenodo API capability snapshot.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"
)

type manifest struct {
	SchemaVersion   int      `json:"schemaVersion"`
	ReviewedDate    string   `json:"reviewedDate"`
	VersionPolicy   string   `json:"versionPolicy"`
	VersionDecision string   `json:"versionDecision"`
	Sources         []source `json:"sources"`
	Snapshot        snapshot `json:"snapshot"`
	SnapshotSHA256  string   `json:"snapshotSha256"`
}

type source struct {
	ID              string   `json:"id"`
	Kind            string   `json:"kind"`
	URL             string   `json:"url"`
	RetrievedDate   string   `json:"retrievedDate"`
	RequiredMarkers []string `json:"requiredMarkers"`
}

type snapshot struct {
	APIGeneration     string         `json:"apiGeneration"`
	ProductionBaseURL string         `json:"productionBaseUrl"`
	SandboxBaseURL    string         `json:"sandboxBaseUrl"`
	OAIPMHBaseURL     string         `json:"oaiPmhBaseUrl"`
	Authentication    authentication `json:"authentication"`
	Limits            limits         `json:"limits"`
	Capabilities      []capability   `json:"capabilities"`
}

type authentication struct {
	PreferredTransport        string   `json:"preferredTransport"`
	PublicRecordsRequireToken bool     `json:"publicRecordsRequireToken"`
	DepositionsRequireToken   bool     `json:"depositionsRequireToken"`
	Scopes                    []string `json:"scopes"`
}

type limits struct {
	GuestRequestsPerMinute         int64 `json:"guestRequestsPerMinute"`
	GuestRequestsPerHour           int64 `json:"guestRequestsPerHour"`
	AuthenticatedRequestsPerMinute int64 `json:"authenticatedRequestsPerMinute"`
	AuthenticatedRequestsPerHour   int64 `json:"authenticatedRequestsPerHour"`
	SearchRequestsPerMinute        int64 `json:"searchRequestsPerMinute"`
	OAIRequestsPerMinute           int64 `json:"oaiRequestsPerMinute"`
	OAIPageSize                    int64 `json:"oaiPageSize"`
	OAIResumptionTokenSeconds      int64 `json:"oaiResumptionTokenSeconds"`
	RecordFileLimit                int64 `json:"recordFileLimit"`
	RecordBytesLimit               int64 `json:"recordBytesLimit"`
}

type capability struct {
	ID        string `json:"id"`
	Protocol  string `json:"protocol"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	Access    string `json:"access"`
	Scope     string `json:"scope"`
	Lifecycle string `json:"lifecycle"`
	Risk      string `json:"risk"`
	Source    string `json:"source"`
}

var tagPattern = regexp.MustCompile(`<[^>]+>`)

func main() {
	if err := execute(os.Args[1:], os.Stdout, &http.Client{Timeout: 30 * time.Second}); err != nil {
		fmt.Fprintf(os.Stderr, "checkzenodoapi: %v\n", err)
		os.Exit(1)
	}
}

func execute(args []string, output io.Writer, client *http.Client) error {
	flags := flag.NewFlagSet("checkzenodoapi", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	online := flags.Bool("online", false, "fetch official sources and verify reviewed evidence markers")
	printDigest := flags.Bool("print-digest", false, "print the canonical snapshot digest")
	manifestPath := flags.String("manifest", "docs/zenodo-api-source.json", "path to the capability manifest")
	if err := flags.Parse(args); err != nil {
		return err
	}
	m, err := run(*manifestPath, *online, client)
	if err != nil {
		return err
	}
	if *printDigest {
		_, err = fmt.Fprintln(output, snapshotDigest(m.Snapshot))
	} else {
		_, err = fmt.Fprintln(output, "Zenodo API capability snapshot: valid")
	}
	return err
}

func run(path string, online bool, client *http.Client) (manifest, error) {
	m, err := loadManifest(path)
	if err != nil {
		return manifest{}, err
	}
	if err := validateManifest(m); err != nil {
		return manifest{}, err
	}
	if online {
		if client == nil {
			return manifest{}, errors.New("online validation requires an HTTP client")
		}
		if err := validateOnline(m, client); err != nil {
			return manifest{}, err
		}
	}
	return m, nil
}

func loadManifest(path string) (manifest, error) {
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return manifest{}, err
	}
	defer func() { _ = f.Close() }()
	decoder := json.NewDecoder(f)
	decoder.DisallowUnknownFields()
	var m manifest
	if err := decoder.Decode(&m); err != nil {
		return manifest{}, err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return manifest{}, errors.New("manifest must contain exactly one JSON value")
	}
	return m, nil
}

func validateManifest(m manifest) error {
	if m.SchemaVersion != 1 {
		return fmt.Errorf("schemaVersion = %d, want 1", m.SchemaVersion)
	}
	if _, err := time.Parse("2006-01-02", m.ReviewedDate); err != nil {
		return fmt.Errorf("reviewedDate: %w", err)
	}
	if m.VersionPolicy != "documentation-date" || strings.TrimSpace(m.VersionDecision) == "" {
		return errors.New("version policy must document the documentation-date decision")
	}
	if err := validateSources(m); err != nil {
		return err
	}
	if err := validateSnapshot(m.Snapshot); err != nil {
		return err
	}
	want := snapshotDigest(m.Snapshot)
	if m.SnapshotSHA256 != want {
		return fmt.Errorf("snapshotSha256 = %q, want %q; review the contract change and update the digest", m.SnapshotSHA256, want)
	}
	return nil
}

func validateSources(m manifest) error {
	wantIDs := []string{"developers", "sandbox", "terms", "policies"}
	if len(m.Sources) != len(wantIDs) {
		return fmt.Errorf("sources count = %d, want %d", len(m.Sources), len(wantIDs))
	}
	seen := make(map[string]bool, len(m.Sources))
	for i, source := range m.Sources {
		if source.ID != wantIDs[i] {
			return fmt.Errorf("source %d id = %q, want %q", i+1, source.ID, wantIDs[i])
		}
		if seen[source.ID] {
			return fmt.Errorf("duplicate source id %q", source.ID)
		}
		seen[source.ID] = true
		if strings.TrimSpace(source.Kind) == "" || len(source.RequiredMarkers) < 2 {
			return fmt.Errorf("source %q requires a kind and at least two evidence markers", source.ID)
		}
		if source.RetrievedDate != m.ReviewedDate {
			return fmt.Errorf("source %q retrievedDate must match reviewedDate", source.ID)
		}
		if err := validateOfficialURL(source.URL); err != nil {
			return fmt.Errorf("source %q: %w", source.ID, err)
		}
	}
	return nil
}

func validateOfficialURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return errors.New("source must be a plain HTTPS URL")
	}
	allowed := map[string]bool{
		"developers.zenodo.org": true,
		"help.zenodo.org":       true,
		"about.zenodo.org":      true,
	}
	if !allowed[strings.ToLower(u.Hostname())] || u.Port() != "" {
		return fmt.Errorf("unapproved official host %q", u.Host)
	}
	return nil
}

func validateSnapshot(s snapshot) error {
	if s.APIGeneration != "documented-depositions-rest" {
		return errors.New("apiGeneration must identify the reviewed depositions REST surface")
	}
	if s.ProductionBaseURL != "https://zenodo.org/api/" || s.SandboxBaseURL != "https://sandbox.zenodo.org/api/" || s.OAIPMHBaseURL != "https://zenodo.org/oai2d" {
		return errors.New("production, sandbox, or OAI-PMH base URL changed")
	}
	if s.Authentication.PreferredTransport != "authorization-bearer-header" || s.Authentication.PublicRecordsRequireToken || !s.Authentication.DepositionsRequireToken {
		return errors.New("authentication policy is invalid")
	}
	if !slices.Equal(s.Authentication.Scopes, []string{"deposit:actions", "deposit:write"}) {
		return errors.New("authentication scopes must be sorted and complete")
	}
	if s.Limits.GuestRequestsPerMinute != 60 || s.Limits.GuestRequestsPerHour != 2000 ||
		s.Limits.AuthenticatedRequestsPerMinute != 100 || s.Limits.AuthenticatedRequestsPerHour != 5000 ||
		s.Limits.SearchRequestsPerMinute != 30 || s.Limits.OAIRequestsPerMinute != 30 ||
		s.Limits.OAIPageSize != 50 || s.Limits.OAIResumptionTokenSeconds != 120 ||
		s.Limits.RecordFileLimit != 100 || s.Limits.RecordBytesLimit != 50*1024*1024*1024 {
		return errors.New("reviewed Zenodo limit snapshot changed")
	}
	return validateCapabilities(s.Capabilities)
}

func validateCapabilities(capabilities []capability) error {
	required := map[string]bool{
		"deposition-create": false, "deposition-publish": false, "file-upload-bucket": false,
		"oai-harvest": false, "record-retrieve": false, "record-search": false,
	}
	seen := make(map[string]bool, len(capabilities))
	previous := ""
	validProtocols := map[string]bool{"rest": true, "oai-pmh": true}
	validRisks := map[string]bool{"read": true, "write": true, "destructive": true, "irreversible": true}
	for _, capability := range capabilities {
		if capability.ID <= previous {
			return errors.New("capabilities must be uniquely sorted by id")
		}
		previous = capability.ID
		if seen[capability.ID] {
			return fmt.Errorf("duplicate capability id %q", capability.ID)
		}
		seen[capability.ID] = true
		if !validProtocols[capability.Protocol] || !validRisks[capability.Risk] || capability.Source != "developers" {
			return fmt.Errorf("capability %q has an invalid protocol, risk, or source", capability.ID)
		}
		if capability.Method == "" || capability.Path == "" || capability.Access == "" || capability.Lifecycle == "" {
			return fmt.Errorf("capability %q has incomplete endpoint semantics", capability.ID)
		}
		if capability.Risk != "read" && capability.Access != "authenticated" {
			return fmt.Errorf("write capability %q must require authentication", capability.ID)
		}
		if _, ok := required[capability.ID]; ok {
			required[capability.ID] = true
		}
	}
	for id, found := range required {
		if !found {
			return fmt.Errorf("required capability %q is missing", id)
		}
	}
	return nil
}

func snapshotDigest(s snapshot) string {
	data, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func validateOnline(m manifest, client *http.Client) error {
	for _, source := range m.Sources {
		request, err := http.NewRequest(http.MethodGet, source.URL, nil)
		if err != nil {
			return err
		}
		request.Header.Set("User-Agent", "osf-cli-go-zenodo-contract-check/1")
		response, err := client.Do(request)
		if err != nil {
			return fmt.Errorf("source %q fetch: %w", source.ID, err)
		}
		body, readErr := io.ReadAll(io.LimitReader(response.Body, 4<<20))
		closeErr := response.Body.Close()
		if readErr != nil {
			return fmt.Errorf("source %q read: %w", source.ID, readErr)
		}
		if closeErr != nil {
			return fmt.Errorf("source %q close: %w", source.ID, closeErr)
		}
		if response.StatusCode != http.StatusOK {
			return fmt.Errorf("source %q returned HTTP %d", source.ID, response.StatusCode)
		}
		if err := validateMarkers(source, body); err != nil {
			return err
		}
	}
	return nil
}

func validateMarkers(source source, body []byte) error {
	text := strings.ToLower(html.UnescapeString(tagPattern.ReplaceAllString(string(body), " ")))
	text = strings.Join(strings.Fields(text), " ")
	for _, marker := range source.RequiredMarkers {
		want := strings.ToLower(strings.Join(strings.Fields(marker), " "))
		if !strings.Contains(text, want) {
			return fmt.Errorf("source %q no longer contains evidence marker %q; review upstream documentation", source.ID, marker)
		}
	}
	return nil
}
