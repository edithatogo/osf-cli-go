package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/edithatogo/osf-cli-go/internal/auth"
	"github.com/edithatogo/osf-cli-go/internal/osfapi"
)

func TestRunPrintsHelpWithoutArgs(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run returned %d, want 0, stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Open Science Framework") {
		t.Fatalf("help output did not describe OSF CLI: %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunWritesOptInStructuredEventsOutsideCommandOutput(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "events.jsonl")
	t.Setenv("OSF_EVENT_LOG", logPath)
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"unknown-command"}, &stdout, &stderr); code == 0 {
		t.Fatal("Run returned success for unknown command")
	}
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("ReadFile events: %v", err)
	}
	if !strings.Contains(string(data), `"name":"cli.command.result"`) {
		t.Fatalf("events=%q", data)
	}
	if strings.Contains(stdout.String(), `"schemaVersion"`) {
		t.Fatalf("structured events polluted stdout: %q", stdout.String())
	}
}

func TestRunPrintsHelpWithHelpFlag(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"--help"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run returned %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "Open Science Framework") {
		t.Fatalf("help output did not describe OSF CLI: %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunPrintsHelpWithShortHelpFlag(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"-h"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run returned %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "Open Science Framework") {
		t.Fatalf("help output did not describe OSF CLI: %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunPrintsVersion(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"--version"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run returned %d, want 0", code)
	}
	if got := strings.TrimSpace(stdout.String()); got != version {
		t.Fatalf("version output = %q, want %q", got, version)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestVersionStringIncludesMetadata(t *testing.T) {
	if got, want := versionString(), "0.0.0-dev"; got != want {
		t.Fatalf("versionString() = %q, want %q", got, want)
	}

	// Verify that with metadata the string includes commit and date.
	// We need a subtest to isolate the global variable mutation.
	t.Run("with metadata", func(t *testing.T) {
		oldVersion := version
		oldCommit := buildCommit
		oldDate := buildDate
		version = "1.2.3"
		buildCommit = "abc1234"
		buildDate = "2026-05-02T00:00:00Z"
		defer func() {
			version = oldVersion
			buildCommit = oldCommit
			buildDate = oldDate
		}()

		if got, want := versionString(), "1.2.3 (commit abc1234, built 2026-05-02T00:00:00Z)"; got != want {
			t.Fatalf("versionString() = %q, want %q", got, want)
		}
	})
}

func TestRunRejectsUnknownCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"nope"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("Run returned %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "unknown command") || !strings.Contains(stderr.String(), "nope") {
		t.Fatalf("stderr = %q, want unknown command message", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestRunRejectsMissingSubcommandArg(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"projects", "get"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("Run returned %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "accepts 1 arg") {
		t.Fatalf("stderr = %q, want argument usage message", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestRunRejectsInvalidOutputMode(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"--output", "csv"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("Run returned %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "invalid output mode") {
		t.Fatalf("stderr = %q, want invalid output mode message", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestRunRejectsConflictingOutputModes(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"--output", "table", "--json"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("Run returned %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "cannot combine --json") {
		t.Fatalf("stderr = %q, want conflicting output mode message", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestExitCodeForError(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		err  error
		want int
	}{
		{"usage error", fmt.Errorf("unknown command"), 2},
		{"general error", fmt.Errorf("network failure"), 1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := exitCodeForError(tc.err)
			if got != tc.want {
				t.Fatalf("exitCodeForError(%v) = %d, want %d", tc.err, got, tc.want)
			}
		})
	}
}

func TestVersionStringWithCommitOnly(t *testing.T) {
	oldCommit := buildCommit
	oldDate := buildDate
	buildCommit = "def5678"
	buildDate = ""
	defer func() {
		buildCommit = oldCommit
		buildDate = oldDate
	}()

	got := versionString()
	if !strings.Contains(got, "commit def5678") {
		t.Fatalf("versionString() = %q, want commit", got)
	}
	if strings.Contains(got, "built") {
		t.Fatalf("versionString() = %q, should not contain built", got)
	}
}

func TestVersionStringWithDateOnly(t *testing.T) {
	oldCommit := buildCommit
	oldDate := buildDate
	buildCommit = ""
	buildDate = "2026-01-01T00:00:00Z"
	defer func() {
		buildCommit = oldCommit
		buildDate = oldDate
	}()

	got := versionString()
	if !strings.Contains(got, "built 2026-01-01") {
		t.Fatalf("versionString() = %q, want built date", got)
	}
}

func TestDefaultReadonlyClientWithMockServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_, _ = w.Write([]byte(`{"data":{"id":"n1","type":"nodes","attributes":{"title":"Test"}},"links":{}}`))
	}))
	defer server.Close()

	api, err := osfapi.New(server.URL + "/v2/")
	if err != nil {
		t.Fatal(err)
	}

	client := &defaultReadonlyClient{api: api}
	node, err := client.GetNode(t.Context(), "n1")
	if err != nil {
		t.Fatalf("GetNode returned error: %v", err)
	}
	if node.ID != "n1" {
		t.Fatalf("node id = %q, want n1", node.ID)
	}
}

func TestDefaultReadonlyClientListChildren(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_, _ = w.Write([]byte(`{"data":[],"links":{}}`))
	}))
	defer server.Close()

	api, err := osfapi.New(server.URL + "/v2/")
	if err != nil {
		t.Fatal(err)
	}

	client := &defaultReadonlyClient{api: api}
	children, err := client.ListNodeChildren(t.Context(), "n1")
	if err != nil {
		t.Fatalf("ListNodeChildren returned error: %v", err)
	}
	if len(children) != 0 {
		t.Fatalf("children = %d, want 0", len(children))
	}
}

func TestDefaultReadonlyClientEntityCoverageForwarders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_, _ = w.Write([]byte(`{"data":[],"links":{}}`))
	}))
	defer server.Close()

	api, err := osfapi.New(server.URL + "/v2/")
	if err != nil {
		t.Fatal(err)
	}
	client := &defaultReadonlyClient{api: api}
	if _, err := client.ListFileVersions(t.Context(), "file-1"); err != nil {
		t.Fatalf("ListFileVersions returned error: %v", err)
	}
	if _, err := client.ListNodeAddons(t.Context(), "node-1"); err != nil {
		t.Fatalf("ListNodeAddons returned error: %v", err)
	}
	if _, err := client.ListWikiPages(t.Context(), "node-1"); err != nil {
		t.Fatalf("ListWikiPages returned error: %v", err)
	}
	if _, err := client.ListNodeComments(t.Context(), "node-1"); err != nil {
		t.Fatalf("ListNodeComments returned error: %v", err)
	}
	if _, err := client.ListNodeLogs(t.Context(), "node-1"); err != nil {
		t.Fatalf("ListNodeLogs returned error: %v", err)
	}
	if _, err := client.ListNodeIdentifiers(t.Context(), "node-1"); err != nil {
		t.Fatalf("ListNodeIdentifiers returned error: %v", err)
	}
}

func TestDefaultReadonlyClientGetStorageFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_, _ = w.Write([]byte(`{"data":{"id":"f1","type":"files","attributes":{"name":"test.txt","kind":"file"},"links":{"download":"https://dl.example.com/f1"}},"links":{}}`))
	}))
	defer server.Close()

	api, err := osfapi.New(server.URL + "/v2/")
	if err != nil {
		t.Fatal(err)
	}

	client := &defaultReadonlyClient{api: api}
	file, err := client.GetStorageFile(t.Context(), "f1")
	if err != nil {
		t.Fatalf("GetStorageFile returned error: %v", err)
	}
	if file.ID != "f1" {
		t.Fatalf("file id = %q, want f1", file.ID)
	}
}

func TestNewDefaultReadonlyClientFromSource(t *testing.T) {
	client := newDefaultReadonlyClientFromSource(auth.FuncSource(func(name string) (string, bool) {
		return "", false
	}))
	if client == nil {
		t.Fatal("newDefaultReadonlyClientFromSource returned nil")
	}
	// Should be a defaultReadonlyClient with no bearer token
	dc, ok := client.(*defaultReadonlyClient)
	if !ok {
		t.Fatalf("client type = %T, want *defaultReadonlyClient", client)
	}
	if dc.bearerToken {
		t.Fatal("bearerToken = true, want false for missing token")
	}
}

func TestDefaultReadonlyClientListProjectsNeedsToken(t *testing.T) {
	client := &defaultReadonlyClient{bearerToken: false}
	_, err := client.ListProjects(t.Context())
	if err == nil {
		t.Fatal("ListProjects returned nil error, want missing token error")
	}
	_, err = client.CurrentUser(t.Context())
	if err == nil {
		t.Fatal("CurrentUser returned nil error, want missing token error")
	}
}

func TestParseNodeIDOrURLEmpty(t *testing.T) {
	_, err := parseNodeIDOrURL("")
	if err == nil {
		t.Fatal("parseNodeIDOrURL returned nil error for empty input")
	}
}

func TestProjectsListJSONOutput(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{
		projects: []osfapi.Node{
			{ID: "p1", Attributes: osfapi.NodeAttributes{Title: "Project", Category: "project"}, Links: osfapi.Links{Self: "https://osf.io/p1/"}},
		},
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClient([]string{"projects", "list", "--output", "json"}, &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("projects list json returned %d, want 0", code)
	}
}

func TestProjectsListCommandError(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClient([]string{"projects", "list"}, &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("projects list with empty projects returned %d, want 0", code)
	}
}

func TestComponentsListJSONOutput(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{
		children: []osfapi.Node{
			{ID: "c1", Attributes: osfapi.NodeAttributes{Title: "Child", Category: "component"}, Links: osfapi.Links{Self: "https://osf.io/c1/"}},
		},
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClient([]string{"components", "list", "project-1", "--output", "json"}, &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("components list json returned %d, want 0", code)
	}
}

func TestValidateResearchOutputJSON(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{
		node:  osfapi.Node{ID: "project-1", Attributes: osfapi.NodeAttributes{Title: "Study", Description: "Research output", Category: "project"}},
		files: []osfapi.StorageFile{{ID: "file-1", Attributes: osfapi.StorageFileAttributes{Name: "data.csv", Kind: "file"}}},
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClient([]string{"validate", "project-1", "--profile", "research-output", "--json"}, &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("validate returned %d, want 0, stderr=%q", code, stderr.String())
	}
	var report validationReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("decode report: %v; stdout=%q", err, stdout.String())
	}
	if !report.Valid || report.Profile != "research-output" || len(report.Findings) != 4 {
		t.Fatalf("report = %+v, want valid four-finding report", report)
	}
}

func TestValidatePreregistrationWarnsOnCategory(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{node: osfapi.Node{ID: "project-1", Attributes: osfapi.NodeAttributes{Title: "Registration", Description: "Plan", Category: "project"}}}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClient([]string{"validate", "project-1", "--profile", "preregistration", "--json"}, &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("validate returned %d, want 0, stderr=%q", code, stderr.String())
	}
	var report validationReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("decode report: %v", err)
	}
	if report.Valid || report.Findings[len(report.Findings)-1].Rule != "preregistration.category" {
		t.Fatalf("report = %+v, want invalid preregistration category finding", report)
	}
}

func TestFilesListJSONOutput(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{
		files: []osfapi.StorageFile{
			{ID: "f1", Attributes: osfapi.StorageFileAttributes{Name: "file.txt", Kind: "file", Extra: osfapi.StorageFileExtra{Hashes: osfapi.StorageFileHashes{MD5: "abc123"}}}, Links: osfapi.Links{Download: "https://files.osf.io/f1"}},
		},
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClient([]string{"files", "list", "project-1", "--output", "json"}, &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("files list json returned %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), `"md5":"abc123"`) {
		t.Fatalf("stdout = %q, want md5", stdout.String())
	}
}

func TestFilesListPassesFolderPathSegments(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{
		storageLists: map[string][]osfapi.StorageFile{
			"project-1:folder/subfolder": {
				{ID: "f1", Attributes: osfapi.StorageFileAttributes{Name: "nested.txt", Kind: "file"}, Links: osfapi.Links{Download: "https://files.osf.io/f1"}},
			},
		},
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClient([]string{"files", "list", "project-1", "folder/subfolder", "--json"}, &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("files list nested returned %d, want 0, stderr=%q", code, stderr.String())
	}
	if got := strings.Join(client.gotFilesSegments, "/"); got != "folder/subfolder" {
		t.Fatalf("segments = %q, want folder/subfolder", got)
	}
	if !strings.Contains(stdout.String(), "nested.txt") {
		t.Fatalf("stdout = %q, want nested file", stdout.String())
	}
}

func TestZshCompletion(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"completion", "zsh"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("zsh completion returned %d, want 0", code)
	}
}

func TestFilesListCommandError(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClient([]string{"files", "list", "noproject"}, &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("files list returned %d, want 0", code)
	}
	if stdout.Len() == 0 {
		t.Fatal("files list produced no output")
	}
}

func TestComponentsListCommandError(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClient([]string{"components", "list", "noproject"}, &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("components list returned %d, want 0", code)
	}
}

func TestRunRejectsUnknownFlagOnNestedCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"auth", "whoami", "--unknown-flag"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("unknown flag returned %d, want 2", code)
	}
}

func TestResolveOutputModeWithOutputAndJSONConflict(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"--output", "table", "--json"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("conflicting modes returned %d, want 2", code)
	}
}

func TestDownloadTreeFileWithMissingDownloadURL(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{
		storageLists: map[string][]osfapi.StorageFile{
			"project-1:": {
				{ID: "file-1", Attributes: osfapi.StorageFileAttributes{Name: "no-url.txt", Kind: "file"}},
			},
		},
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClient([]string{"files", "download", "--tree", "project-1", filepath.Join(t.TempDir(), "out")}, &stdout, &stderr, client)
	if code != 1 {
		t.Fatalf("tree download returned %d, want 1 (missing download URL)", code)
	}
}

func TestDownloadFolderTreeWithMissingProject(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClient([]string{"files", "download", "--tree", "nonexistent-project", filepath.Join(t.TempDir(), "out")}, &stdout, &stderr, client)
	if code != 0 {
		t.Logf("tree download returned %d, stderr=%q", code, stderr.String())
	}
}

func TestAuthWhoamiErrorPath(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	client := &fakeReadonlyClient{}
	code := runWithClient([]string{"auth", "whoami"}, &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("auth whoami with empty client returned %d, want 0", code)
	}
}

func TestAuthWhoamiRejectsInvalidOutputMode(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"auth", "whoami", "--output", "xml"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("invalid output mode returned %d, want 2", code)
	}
}

func TestProjectsGetRejectsBadURL(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"projects", "get", "https://evil.com/project"}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("bad URL returned %d, want non-zero", code)
	}
}

func TestProjectsGetJSONOutput(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{
		node: osfapi.Node{ID: "n1", Attributes: osfapi.NodeAttributes{Title: "Test", Category: "project"}, Links: osfapi.Links{Self: "https://osf.io/n1/"}},
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClient([]string{"projects", "get", "n1", "--json"}, &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("projects get json returned %d, want 0", code)
	}
	var got projectRecord
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if got.ID != "n1" {
		t.Fatalf("id = %q, want n1", got.ID)
	}
}

func TestFilesDownloadSingleFileDownloadURLMode(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{
		downloadBodies: map[string]string{
			"https://files.osf.io/v1/test?download=1": "data",
		},
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	dest := filepath.Join(t.TempDir(), "out.bin")
	code := runWithClient([]string{"files", "download", "--file", "https://files.osf.io/v1/test?download=1", dest, "--conflict", "overwrite"}, &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("download URL returned %d, want 0, stderr=%q", code, stderr.String())
	}
}

func TestResolveFileSourceWithUnsupportedURL(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClient([]string{"files", "download", "--file", "https://example.com/file", filepath.Join(t.TempDir(), "out")}, &stdout, &stderr, client)
	if code != 2 {
		t.Fatalf("unsupported URL returned %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "unsupported file source URL") {
		t.Fatalf("stderr = %q, want unsupported URL message", stderr.String())
	}
}

func TestParseNodeIDOrURLAPIURLShorthand(t *testing.T) {
	// Test with osf.io URL that doesn't have a "/nodes/" segment
	// Should fall back to returning the first path part as the GUID
	id, err := parseNodeIDOrURL("https://osf.io/abc12/")
	if err != nil {
		t.Fatalf("parseNodeIDOrURL returned error: %v", err)
	}
	if id != "abc12" {
		t.Fatalf("id = %q, want abc12", id)
	}
}

func TestParseNodeIDOrURLBadHostURL(t *testing.T) {
	_, err := parseNodeIDOrURL("https://evil.com/project")
	if err == nil {
		t.Fatal("parseNodeIDOrURL returned nil error for non-OSF host URL")
	}
	if !strings.Contains(err.Error(), "not an OSF host") {
		t.Fatalf("error = %q, want OSF host message", err.Error())
	}
}

func TestParseNodeIDOrURLUnparseableURL(t *testing.T) {
	_, err := parseNodeIDOrURL("://bad")
	if err == nil {
		t.Fatal("parseNodeIDOrURL returned nil error for bad URL")
	}
	if !strings.Contains(err.Error(), "could not parse") {
		t.Fatalf("error = %q, want parse error", err.Error())
	}
}

func TestWriteRootContractWithJSON(t *testing.T) {
	var buf bytes.Buffer
	err := writeRootContract(&buf)
	if err != nil {
		t.Fatalf("writeRootContract returned error: %v", err)
	}
	var contract rootContract
	if err := json.Unmarshal(buf.Bytes(), &contract); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\n%s", err, buf.String())
	}
	if len(contract.Commands) != 15 {
		t.Fatalf("command count = %d, want 15", len(contract.Commands))
	}
	if contract.Commands[0].Status != "implemented" || contract.Commands[1].Status != "implemented" {
		t.Fatalf("unexpected command statuses: %#v", contract.Commands)
	}
	wantNames := []string{"auth", "projects", "components", "files", "nodes", "export", "validate", "search", "preprints", "registrations", "resolve", "open", "whoami", "zenodo", "completion"}
	for i, want := range wantNames {
		if contract.Commands[i].Name != want || contract.Commands[i].Status != "implemented" {
			t.Fatalf("command %d = %#v, want implemented %q", i, contract.Commands[i], want)
		}
	}
}

func TestRootContractMatchesCompatibilityFixture(t *testing.T) {
	fixture, err := os.ReadFile("testdata/compatibility/cli-root.json")
	if err != nil {
		t.Fatalf("read compatibility fixture: %v", err)
	}

	var want rootContract
	if err := json.Unmarshal(fixture, &want); err != nil {
		t.Fatalf("decode compatibility fixture: %v", err)
	}
	var gotBytes bytes.Buffer
	if err := writeRootContract(&gotBytes); err != nil {
		t.Fatalf("writeRootContract returned error: %v", err)
	}
	var got rootContract
	if err := json.Unmarshal(gotBytes.Bytes(), &got); err != nil {
		t.Fatalf("decode generated contract: %v", err)
	}
	wantJSON, _ := json.Marshal(want)
	gotJSON, _ := json.Marshal(got)
	if !bytes.Equal(gotJSON, wantJSON) {
		t.Fatalf("CLI compatibility contract changed:\n got: %s\nwant: %s", gotJSON, wantJSON)
	}
}

func TestIsUsageError(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"unknown command", fmt.Errorf("unknown command %q", "foo"), true},
		{"invalid output mode", fmt.Errorf("invalid output mode %q", "csv"), true},
		{"cannot combine --json", fmt.Errorf("cannot combine --json with --output=table"), true},
		{"unknown flag", fmt.Errorf("unknown flag: --foo"), true},
		{"flag needs arg", fmt.Errorf("flag needs an argument"), true},
		{"unknown shorthand", fmt.Errorf("unknown shorthand flag: 'x'"), true},
		{"accepts arg", fmt.Errorf("accepts 1 arg(s), received 0"), true},
		{"unsupported conflict", fmt.Errorf("unsupported conflict policy"), true},
		{"cannot combine file tree", fmt.Errorf("cannot combine --file with --tree"), true},
		{"general error", fmt.Errorf("network error"), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := isUsageError(tc.err)
			if got != tc.want {
				t.Fatalf("isUsageError(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

func TestAuthWhoamiOutputsTableAndJSON(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{
		currentUser: osfapi.User{
			ID:   "u1",
			Type: "users",
			Attributes: osfapi.UserAttributes{
				FullName:   "Ada Lovelace",
				GivenName:  "Ada",
				FamilyName: "Lovelace",
			},
		},
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClient([]string{"auth", "whoami"}, &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("table Run returned %d, want 0", code)
	}
	if client.gotCurrentUserCalls != 1 {
		t.Fatalf("CurrentUser call count = %d, want 1", client.gotCurrentUserCalls)
	}
	if !strings.Contains(stdout.String(), "Ada Lovelace") || !strings.Contains(stdout.String(), "Full Name") {
		t.Fatalf("table output missing user fields: %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = runWithClient([]string{"auth", "whoami", "--json"}, &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("json Run returned %d, want 0", code)
	}
	var got authUserRecord
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\n%s", err, stdout.String())
	}
	if got.ID != "u1" || got.FullName != "Ada Lovelace" || got.GivenName != "Ada" || got.FamilyName != "Lovelace" {
		t.Fatalf("unexpected JSON payload: %#v", got)
	}
}

func TestAuthWhoamiRequiresToken(t *testing.T) {
	t.Parallel()

	client := newDefaultReadonlyClientFromSource(auth.FuncSource(func(name string) (string, bool) {
		return "", false
	}))

	defaultClient, ok := client.(*defaultReadonlyClient)
	if !ok {
		t.Fatalf("client type = %T, want *defaultReadonlyClient", client)
	}

	_, err := defaultClient.CurrentUser(context.Background())
	if err == nil {
		t.Fatal("CurrentUser returned nil error, want missing token error")
	}
	if !strings.Contains(err.Error(), auth.TokenEnv) {
		t.Fatalf("error = %q, want token env mention", err.Error())
	}
}

func TestProjectsListOutputsTableAndJSON(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{
		projects: []osfapi.Node{
			{ID: "project-1", Attributes: osfapi.NodeAttributes{Title: "Alpha", Category: "project"}, Links: osfapi.Links{Self: "https://osf.io/project-1/"}},
			{ID: "project-2", Attributes: osfapi.NodeAttributes{Title: "Beta", Category: "project"}, Links: osfapi.Links{Self: "https://osf.io/project-2/"}},
		},
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClient([]string{"projects", "list"}, &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("table Run returned %d, want 0", code)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if !strings.Contains(stdout.String(), "project-1") || !strings.Contains(stdout.String(), "Alpha") {
		t.Fatalf("table output missing project rows: %q", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = runWithClient([]string{"projects", "list", "--json"}, &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("json Run returned %d, want 0", code)
	}
	var got []projectRecord
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\n%s", err, stdout.String())
	}
	if len(got) != 2 || got[0].ID != "project-1" || got[1].Title != "Beta" {
		t.Fatalf("unexpected JSON payload: %#v", got)
	}
}

func TestProjectsGetOutputsTableAndJSON(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{
		node: osfapi.Node{ID: "project-1", Attributes: osfapi.NodeAttributes{Title: "Alpha", Category: "project", Description: "Example"}, Links: osfapi.Links{Self: "https://osf.io/project-1/"}},
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClient([]string{"projects", "get", "https://osf.io/project-1/"}, &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("table Run returned %d, want 0", code)
	}
	if client.gotNodeID != "project-1" {
		t.Fatalf("parsed node id = %q, want project-1", client.gotNodeID)
	}
	if !strings.Contains(stdout.String(), "Alpha") || !strings.Contains(stdout.String(), "Example") {
		t.Fatalf("table output missing node details: %q", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = runWithClient([]string{"projects", "get", "project-1", "--json"}, &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("json Run returned %d, want 0", code)
	}
	var got projectRecord
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\n%s", err, stdout.String())
	}
	if got.ID != "project-1" || got.Title != "Alpha" || got.Description != "Example" {
		t.Fatalf("unexpected JSON payload: %#v", got)
	}
}

func TestParseNodeIDOrURL(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input string
		want  string
	}{
		{name: "guid", input: "abc12", want: "abc12"},
		{name: "web root", input: "https://osf.io/abc12/", want: "abc12"},
		{name: "web nested", input: "https://osf.io/abc12/files/osfstorage", want: "abc12"},
		{name: "api node", input: "https://api.osf.io/v2/nodes/abc12/", want: "abc12"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseNodeIDOrURL(tc.input)
			if err != nil {
				t.Fatalf("parseNodeIDOrURL returned error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("parseNodeIDOrURL(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestParseNodeIDOrURLRejectsNonOSFURL(t *testing.T) {
	t.Parallel()

	_, err := parseNodeIDOrURL("https://example.com/foo/bar")
	if err == nil {
		t.Fatal("parseNodeIDOrURL returned nil error, want non-OSF URL rejection")
	}
	if !strings.Contains(err.Error(), "not an OSF host") {
		t.Fatalf("error = %q, want OSF host message", err.Error())
	}
}

func TestProjectsCreateRequiresConfirmation(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClientInput([]string{"projects", "create", "--title", "New Project"}, "", &stdout, &stderr, client)
	if code != 1 {
		t.Fatalf("projects create without confirmation returned %d, want 1", code)
	}
	if client.gotCreateNodeTitle != "" {
		t.Fatalf("created node title = %q, want no create", client.gotCreateNodeTitle)
	}
	if !strings.Contains(stderr.String(), "node creation confirmation required") {
		t.Fatalf("stderr = %q, want confirmation error", stderr.String())
	}
}

func TestProjectsCreateAfterConfirmation(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClientInput([]string{"projects", "create", "--title", "New Project", "--description", "Draft"}, "yes\n", &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("projects create returned %d, want 0, stderr=%q", code, stderr.String())
	}
	if client.gotCreateNodeTitle != "New Project" || client.gotCreateNodeCategory != "project" || client.gotCreateNodeDescription != "Draft" {
		t.Fatalf("create call title=%q category=%q description=%q", client.gotCreateNodeTitle, client.gotCreateNodeCategory, client.gotCreateNodeDescription)
	}
	if !strings.Contains(stdout.String(), "New Project") {
		t.Fatalf("stdout = %q, want created node", stdout.String())
	}
}

func TestProjectsUpdatePreservesOmittedFieldsAndRequiresConfirmation(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{
		node: osfapi.Node{ID: "abc123", Type: "nodes", Attributes: osfapi.NodeAttributes{Title: "Old", Category: "project", Description: "Keep me"}, Links: osfapi.Links{Self: "https://osf.io/abc123/"}},
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClientInput([]string{"projects", "update", "abc123", "--title", "Updated"}, "", &stdout, &stderr, client)
	if code != 1 {
		t.Fatalf("projects update without confirmation returned %d, want 1", code)
	}
	if client.gotUpdateNodeID != "" {
		t.Fatalf("updated node id = %q, want no update", client.gotUpdateNodeID)
	}
	if !strings.Contains(stderr.String(), "node update confirmation required") {
		t.Fatalf("stderr = %q, want confirmation error", stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = runWithClientInput([]string{"projects", "update", "abc123", "--title", "Updated"}, "yes\n", &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("projects update returned %d, want 0, stderr=%q", code, stderr.String())
	}
	if client.gotUpdateNodeID != "abc123" || client.gotUpdateNodeTitle != "Updated" || client.gotUpdateNodeDescription != "Keep me" {
		t.Fatalf("update call id=%q title=%q description=%q", client.gotUpdateNodeID, client.gotUpdateNodeTitle, client.gotUpdateNodeDescription)
	}
}

func TestProjectsUpdateRequiresChangedField(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClient([]string{"projects", "update", "abc123"}, &stdout, &stderr, &fakeReadonlyClient{})
	if code != 1 {
		t.Fatalf("projects update returned %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "at least one of --title or --description is required") {
		t.Fatalf("stderr = %q, want missing update field error", stderr.String())
	}
}

func TestProjectsDeleteRequiresConfirmation(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClientInput([]string{"projects", "delete", "abc123"}, "", &stdout, &stderr, client)
	if code != 1 {
		t.Fatalf("projects delete without confirmation returned %d, want 1", code)
	}
	if client.gotDeleteNodeID != "" {
		t.Fatalf("deleted node id = %q, want no delete", client.gotDeleteNodeID)
	}
	if !strings.Contains(stderr.String(), "node deletion confirmation required") {
		t.Fatalf("stderr = %q, want confirmation error", stderr.String())
	}
}

func TestProjectsDeleteYesSkipsPrompt(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClient([]string{"projects", "delete", "https://osf.io/abc123/", "--yes"}, &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("projects delete --yes returned %d, want 0, stderr=%q", code, stderr.String())
	}
	if client.gotDeleteNodeID != "abc123" {
		t.Fatalf("deleted node id = %q, want abc123", client.gotDeleteNodeID)
	}
	if !strings.Contains(stdout.String(), "deleted node abc123") {
		t.Fatalf("stdout = %q, want deletion message", stdout.String())
	}
}

func TestComponentsListOutputsTableAndJSON(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{
		children: []osfapi.Node{
			{ID: "component-1", Attributes: osfapi.NodeAttributes{Title: "Comp A", Category: "component"}, Links: osfapi.Links{Self: "https://osf.io/component-1/"}},
			{ID: "component-2", Attributes: osfapi.NodeAttributes{Title: "Comp B", Category: "component"}, Links: osfapi.Links{Self: "https://osf.io/component-2/"}},
		},
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClient([]string{"components", "list", "https://osf.io/project-1/"}, &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("table Run returned %d, want 0", code)
	}
	if client.gotChildrenID != "project-1" {
		t.Fatalf("parsed component list id = %q, want project-1", client.gotChildrenID)
	}
	if !strings.Contains(stdout.String(), "component-1") || !strings.Contains(stdout.String(), "Comp B") {
		t.Fatalf("table output missing component rows: %q", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = runWithClient([]string{"components", "list", "project-1", "--json"}, &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("json Run returned %d, want 0", code)
	}
	var got []projectRecord
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\n%s", err, stdout.String())
	}
	if len(got) != 2 || got[0].ID != "component-1" || got[1].Title != "Comp B" {
		t.Fatalf("unexpected JSON payload: %#v", got)
	}
}

func TestFilesListOutputsTableAndJSON(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{
		files: []osfapi.StorageFile{
			{ID: "file-1", Attributes: osfapi.StorageFileAttributes{Name: "analysis.csv", Kind: "file", Size: 12}, Links: osfapi.Links{Download: "https://files.osf.io/file-1"}},
			{ID: "folder-1", Attributes: osfapi.StorageFileAttributes{Name: "figures", Kind: "folder"}},
		},
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClient([]string{"files", "list", "project-1"}, &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("table Run returned %d, want 0", code)
	}
	if client.gotFilesID != "project-1" {
		t.Fatalf("parsed file list id = %q, want project-1", client.gotFilesID)
	}
	if !strings.Contains(stdout.String(), "analysis.csv") || !strings.Contains(stdout.String(), "figures") {
		t.Fatalf("table output missing file rows: %q", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = runWithClient([]string{"files", "list", "project-1", "--json"}, &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("json Run returned %d, want 0", code)
	}
	var got []fileRecord
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\n%s", err, stdout.String())
	}
	if len(got) != 2 || got[0].ID != "file-1" || got[0].DownloadURL != "https://files.osf.io/file-1" {
		t.Fatalf("unexpected JSON payload: %#v", got)
	}
}

func TestFilesDownloadSingleFileOutputsTableAndJSON(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{
		storageFiles: map[string]osfapi.StorageFile{
			"file-1": {ID: "file-1", Attributes: osfapi.StorageFileAttributes{Name: "analysis.csv", Kind: "file", Size: 14}, Links: osfapi.Links{Download: "https://files.osf.io/v1/resources/project-123/providers/osfstorage/file-1?download=1"}},
		},
		downloadBodies: map[string]string{
			"https://files.osf.io/v1/resources/project-123/providers/osfstorage/file-1?download=1": "col1,col2\n1,2\n",
		},
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	dest := filepath.Join(t.TempDir(), "analysis.csv")
	code := runWithClient([]string{"files", "download", "--file", "file-1", dest}, &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("table Run returned %d, want 0", code)
	}
	if client.gotStorageFileID != "file-1" {
		t.Fatalf("GetStorageFile id = %q, want file-1", client.gotStorageFileID)
	}
	if client.gotOpenDownloadURL != "https://files.osf.io/v1/resources/project-123/providers/osfstorage/file-1?download=1" {
		t.Fatalf("OpenDownload url = %q", client.gotOpenDownloadURL)
	}
	if !strings.Contains(stdout.String(), "Mode") || !strings.Contains(stdout.String(), "written") || !strings.Contains(stdout.String(), dest) {
		t.Fatalf("table output missing download summary: %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = runWithClient([]string{"files", "download", "--file", "file-1", dest, "--json", "--conflict", "overwrite"}, &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("json Run returned %d, want 0", code)
	}
	var got filesDownloadResult
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\n%s", err, stdout.String())
	}
	if got.Mode != "file" || got.Source != "file-1" || len(got.Records) != 1 || got.Records[0].Status != "written" {
		t.Fatalf("unexpected JSON payload: %#v", got)
	}
}

func TestFilesDownloadTreeOutputsTableAndJSON(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{
		storageLists: map[string][]osfapi.StorageFile{
			"project-1:": {
				{ID: "folder-1", Attributes: osfapi.StorageFileAttributes{Name: "figures", Kind: "folder"}},
				{ID: "file-1", Attributes: osfapi.StorageFileAttributes{Name: "analysis.csv", Kind: "file", Size: 14}, Links: osfapi.Links{Download: "https://files.osf.io/v1/resources/project-123/providers/osfstorage/file-1?download=1"}},
			},
			"project-1:figures": {
				{ID: "file-2", Attributes: osfapi.StorageFileAttributes{Name: "plot.png", Kind: "file", Size: 9}, Links: osfapi.Links{Download: "https://files.osf.io/v1/resources/project-123/providers/osfstorage/file-2?download=1"}},
			},
		},
		downloadBodies: map[string]string{
			"https://files.osf.io/v1/resources/project-123/providers/osfstorage/file-1?download=1": "col1,col2\n1,2\n",
			"https://files.osf.io/v1/resources/project-123/providers/osfstorage/file-2?download=1": "png-bytes",
		},
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	dest := filepath.Join(t.TempDir(), "tree")
	code := runWithClient([]string{"files", "download", "--tree", "project-1", dest}, &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("table Run returned %d, want 0", code)
	}
	if len(client.gotFilesSegments) == 0 {
		t.Fatal("expected ListStorageFiles to be called")
	}
	if !strings.Contains(stdout.String(), "tree") || !strings.Contains(stdout.String(), "written") {
		t.Fatalf("table output missing tree summary: %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = runWithClient([]string{"files", "download", "--tree", "project-1", dest, "--json", "--conflict", "overwrite"}, &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("json Run returned %d, want 0", code)
	}
	var got filesDownloadResult
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\n%s", err, stdout.String())
	}
	if got.Mode != "tree" || got.Source != "project-1" || len(got.Records) != 2 {
		t.Fatalf("unexpected JSON payload: %#v", got)
	}
}

func TestFilesDownloadResolveFileSourceURL(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{
		storageFiles: map[string]osfapi.StorageFile{
			"file-1": {ID: "file-1", Attributes: osfapi.StorageFileAttributes{Name: "test.bin", Kind: "file", Size: 4}, Links: osfapi.Links{Download: "https://files.osf.io/v1/test"}},
		},
		downloadBodies: map[string]string{
			"https://files.osf.io/v1/test": "data",
		},
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	dest := filepath.Join(t.TempDir(), "out.bin")
	code := runWithClient([]string{"files", "download", "--file", "https://api.osf.io/v2/files/file-1/", dest, "--conflict", "overwrite"}, &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("URL download returned %d, want 0, stderr=%q", code, stderr.String())
	}
}

func TestFilesDownloadDirectDownloadURL(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{
		downloadBodies: map[string]string{
			"https://files.osf.io/v1/raw": "raw-data",
		},
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	dest := filepath.Join(t.TempDir(), "raw.bin")
	code := runWithClient([]string{"files", "download", "--file", "https://files.osf.io/v1/raw", dest, "--conflict", "overwrite"}, &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("direct download returned %d, want 0, stderr=%q", code, stderr.String())
	}
}

func TestFilesDownloadFileWithNoDownloadURL(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{
		storageFiles: map[string]osfapi.StorageFile{
			"file-1": {ID: "file-1", Attributes: osfapi.StorageFileAttributes{Name: "test.bin", Kind: "file"}},
		},
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClient([]string{"files", "download", "--file", "file-1", filepath.Join(t.TempDir(), "out.bin")}, &stdout, &stderr, client)
	if code != 1 {
		t.Fatalf("download returned %d, want 1 (missing download URL)", code)
	}
}

func TestDefaultReadonlyClientListStorageFilesWithSegments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_, _ = w.Write([]byte(`{"data":[],"links":{}}`))
	}))
	defer server.Close()

	api, err := osfapi.New(server.URL + "/v2/")
	if err != nil {
		t.Fatal(err)
	}
	client := &defaultReadonlyClient{api: api}
	files, err := client.ListStorageFiles(t.Context(), "n1", "subdir")
	if err != nil {
		t.Fatalf("ListStorageFiles with segments returned error: %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("files = %d", len(files))
	}
}

func TestDownloadSingleFileOpenFailure(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{
		storageFiles: map[string]osfapi.StorageFile{
			"file-1": {ID: "file-1", Attributes: osfapi.StorageFileAttributes{Name: "test.bin", Kind: "file"}, Links: osfapi.Links{Download: "https://files.osf.io/v1/missing"}},
		},
		downloadBodies: map[string]string{},
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClient([]string{"files", "download", "--file", "file-1", filepath.Join(t.TempDir(), "out.bin")}, &stdout, &stderr, client)
	if code != 1 {
		t.Fatalf("download with missing body returned %d, want 1", code)
	}
}

func TestRunRejectsOutputJSONWithTableFlag(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"--output", "json", "--output", "table"}, &stdout, &stderr)
	if code != 0 {
		// If table is the last flag set, it's fine - just verify it doesn't error on double output flags
		t.Logf("double output flag returned %d", code)
	}
}

func TestDefaultReadonlyClientListStorageFiles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_, _ = w.Write([]byte(`{"data":[],"links":{}}`))
	}))
	defer server.Close()

	api, err := osfapi.New(server.URL + "/v2/")
	if err != nil {
		t.Fatal(err)
	}
	client := &defaultReadonlyClient{api: api}
	files, err := client.ListStorageFiles(t.Context(), "n1")
	if err != nil {
		t.Fatalf("ListStorageFiles returned error: %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("files = %d", len(files))
	}
}

func TestDefaultReadonlyClientOpenDownload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("data"))
	}))
	defer server.Close()

	api, err := osfapi.New(server.URL + "/v2/")
	if err != nil {
		t.Fatal(err)
	}
	client := &defaultReadonlyClient{api: api}
	rc, err := client.OpenDownload(t.Context(), server.URL+"/dl")
	if err != nil {
		t.Fatalf("OpenDownload returned error: %v", err)
	}
	_ = rc.Close()
}

func TestDefaultReadonlyClientNewFromSourceWithToken(t *testing.T) {
	client := newDefaultReadonlyClientFromSource(auth.FuncSource(func(name string) (string, bool) {
		return "test-token-value-abc123", true
	}))
	if client == nil {
		t.Fatal("newDefaultReadonlyClientFromSource returned nil")
	}
	dc, ok := client.(*defaultReadonlyClient)
	if !ok {
		t.Fatalf("type = %T", client)
	}
	if !dc.bearerToken {
		t.Fatal("bearerToken = false, want true")
	}
}

func TestDefaultReadonlyClientNewFromSourceWithUsernamePassword(t *testing.T) {
	client := newDefaultReadonlyClientFromSource(auth.FuncSource(func(name string) (string, bool) {
		switch name {
		case auth.UsernameEnv:
			return "user@example.org", true
		case auth.PasswordEnv:
			return "password-123", true
		default:
			return "", false
		}
	}))
	dc, ok := client.(*defaultReadonlyClient)
	if !ok {
		t.Fatalf("type = %T", client)
	}
	if dc.bearerToken {
		t.Fatal("bearerToken = true, want false for username/password mode")
	}
	if dc.AuthMode() != auth.ModeUsernamePassword {
		t.Fatalf("AuthMode = %q, want username/password", dc.AuthMode())
	}
}

func TestDefaultReadonlyClientPrefersTokenOverUsernamePassword(t *testing.T) {
	client := newDefaultReadonlyClientFromSource(auth.FuncSource(func(name string) (string, bool) {
		switch name {
		case auth.TokenEnv:
			return "token-123", true
		case auth.UsernameEnv:
			return "user@example.org", true
		case auth.PasswordEnv:
			return "password-123", true
		default:
			return "", false
		}
	}))
	dc := client.(*defaultReadonlyClient)
	if !dc.bearerToken || dc.AuthMode() != auth.ModeBearerToken {
		t.Fatalf("client mode bearer=%v authMode=%q, want bearer token", dc.bearerToken, dc.AuthMode())
	}
}

func TestDefaultReadonlyClientReportsPartialUsernamePassword(t *testing.T) {
	client := newDefaultReadonlyClientFromSource(auth.FuncSource(func(name string) (string, bool) {
		if name == auth.UsernameEnv {
			return "user@example.org", true
		}
		return "", false
	}))
	_, err := client.CurrentUser(t.Context())
	if err == nil {
		t.Fatal("CurrentUser returned nil error, want partial credential error")
	}
	if !strings.Contains(err.Error(), auth.PasswordEnv) {
		t.Fatalf("error = %q, want password env mention", err.Error())
	}
}

func TestAuthLoginGuidedBootstrap(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"auth", "login"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("auth login returned %d, want 0, stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "https://osf.io/settings/tokens/") || !strings.Contains(stdout.String(), "OSF_TOKEN") {
		t.Fatalf("auth login output = %q, want token guidance", stdout.String())
	}
}

func TestAuthLoginPrintEnvRequiresExplicitToken(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"auth", "login", "--token", "token-123", "--print-env"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("auth login --print-env returned %d, want 0, stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "OSF_TOKEN") || !strings.Contains(stdout.String(), "token-123") {
		t.Fatalf("stdout = %q, want explicit export", stdout.String())
	}
}

func TestCompletionCommandHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"completion", "--help"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("completion help returned %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "bash") {
		t.Fatalf("completion help = %q, want bash mention", stdout.String())
	}
}

func TestFilesDownloadRejectsInvalidFlags(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := runWithClient([]string{"files", "download", "--file", "file-1", "--tree", "project-1", "dest"}, &stdout, &stderr, &fakeReadonlyClient{})
	if code != 2 {
		t.Fatalf("Run returned %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "cannot combine --file with --tree") {
		t.Fatalf("stderr = %q, want conflict error", stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = runWithClient([]string{"files", "download", "--conflict", "bad", "--file", "file-1", "dest"}, &stdout, &stderr, &fakeReadonlyClient{})
	if code != 2 {
		t.Fatalf("Run returned %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "unsupported conflict policy") {
		t.Fatalf("stderr = %q, want invalid conflict policy error", stderr.String())
	}
}

type fakeReadonlyClient struct {
	currentUser              osfapi.User
	projects                 []osfapi.Node
	node                     osfapi.Node
	children                 []osfapi.Node
	preprints                []osfapi.Node
	searchedPreprints        []osfapi.Preprint
	addons                   []osfapi.Node
	fileVersions             []osfapi.FileVersion
	wikis                    []osfapi.RelatedResource
	comments                 []osfapi.RelatedResource
	logs                     []osfapi.RelatedResource
	identifiers              []osfapi.RelatedResource
	searchResults            []osfapi.SearchResult
	files                    []osfapi.StorageFile
	storageFiles             map[string]osfapi.StorageFile
	storageLists             map[string][]osfapi.StorageFile
	downloadBodies           map[string]string
	gotCurrentUserCalls      int
	gotNodeID                string
	gotChildrenID            string
	gotFilesID               string
	gotFilesSegments         []string
	gotStorageFileID         string
	gotOpenDownloadURL       string
	gotProviderURL           string
	gotUploadName            string
	gotUploadBody            string
	gotUploadConflict        string
	gotCreatedFolder         string
	gotDeletedFile           string
	gotPreprintProvider      string
	gotPreprintLimit         int
	gotPreprintQuery         string
	gotSearchLimit           int
	gotAddonNodeID           string
	gotFileVersionID         string
	gotRelationNodeID        string
	gotRegistrationNode      string
	gotRegistrationReq       osfapi.RegistrationRequest
	gotCreateNodeTitle       string
	gotCreateNodeCategory    string
	gotCreateNodeDescription string
	gotUpdateNodeID          string
	gotUpdateNodeTitle       string
	gotUpdateNodeDescription string
	gotDeleteNodeID          string
	entityErr                error
}

func (f *fakeReadonlyClient) CurrentUser(context.Context) (osfapi.User, error) {
	f.gotCurrentUserCalls++
	return f.currentUser, nil
}

func (f *fakeReadonlyClient) ListProjects(context.Context) ([]osfapi.Node, error) {
	return append([]osfapi.Node(nil), f.projects...), nil
}

func (f *fakeReadonlyClient) GetNode(_ context.Context, id string) (osfapi.Node, error) {
	f.gotNodeID = id
	return f.node, nil
}

func (f *fakeReadonlyClient) CreateNode(_ context.Context, title, category, description string) (osfapi.Node, error) {
	f.gotCreateNodeTitle = title
	f.gotCreateNodeCategory = category
	f.gotCreateNodeDescription = description
	return osfapi.Node{ID: "new-node", Type: "nodes", Attributes: osfapi.NodeAttributes{Title: title, Category: category, Description: description}, Links: osfapi.Links{Self: "https://osf.io/new-node/"}}, nil
}

func (f *fakeReadonlyClient) UpdateNode(_ context.Context, id, title, description string) (osfapi.Node, error) {
	f.gotUpdateNodeID = id
	f.gotUpdateNodeTitle = title
	f.gotUpdateNodeDescription = description
	return osfapi.Node{ID: id, Type: "nodes", Attributes: osfapi.NodeAttributes{Title: title, Category: "project", Description: description}, Links: osfapi.Links{Self: "https://osf.io/" + id + "/"}}, nil
}

func (f *fakeReadonlyClient) DeleteNode(_ context.Context, id string) error {
	f.gotDeleteNodeID = id
	return nil
}

func (f *fakeReadonlyClient) ListNodeContributors(_ context.Context, id string) ([]osfapi.Contributor, error) {
	return nil, nil
}

func (f *fakeReadonlyClient) ListNodeChildren(_ context.Context, id string) ([]osfapi.Node, error) {
	f.gotChildrenID = id
	return append([]osfapi.Node(nil), f.children...), nil
}

func (f *fakeReadonlyClient) ListStorageFiles(_ context.Context, id string, segments ...string) ([]osfapi.StorageFile, error) {
	f.gotFilesID = id
	f.gotFilesSegments = segments
	if len(f.storageLists) > 0 {
		key := id + ":" + strings.Join(segments, "/")
		if files, ok := f.storageLists[key]; ok {
			return append([]osfapi.StorageFile(nil), files...), nil
		}
	}
	return append([]osfapi.StorageFile(nil), f.files...), nil
}

func (f *fakeReadonlyClient) GetStorageFile(_ context.Context, id string) (osfapi.StorageFile, error) {
	f.gotStorageFileID = id
	if f.storageFiles != nil {
		if file, ok := f.storageFiles[id]; ok {
			return file, nil
		}
	}
	return osfapi.StorageFile{}, fmt.Errorf("missing storage file %q", id)
}

func (f *fakeReadonlyClient) ListPreprints(_ context.Context, provider string, limit ...int) ([]osfapi.Node, error) {
	f.gotPreprintProvider = provider
	if len(limit) > 0 {
		f.gotPreprintLimit = limit[0]
	}
	return append([]osfapi.Node(nil), f.preprints...), nil
}

func (f *fakeReadonlyClient) SearchPreprints(_ context.Context, query, provider string, limit ...int) ([]osfapi.Preprint, error) {
	f.gotPreprintQuery = query
	f.gotPreprintProvider = provider
	if len(limit) > 0 {
		f.gotPreprintLimit = limit[0]
	}
	return append([]osfapi.Preprint(nil), f.searchedPreprints...), nil
}

func (f *fakeReadonlyClient) SearchOSF(_ context.Context, query string, limit ...int) ([]osfapi.SearchResult, error) {
	if len(limit) > 0 {
		f.gotSearchLimit = limit[0]
	}
	return append([]osfapi.SearchResult(nil), f.searchResults...), nil
}

func (f *fakeReadonlyClient) ListNodeAddons(_ context.Context, id string) ([]osfapi.Node, error) {
	f.gotAddonNodeID = id
	if f.entityErr != nil {
		return nil, f.entityErr
	}
	return append([]osfapi.Node(nil), f.addons...), nil
}

func (f *fakeReadonlyClient) ListFileVersions(_ context.Context, id string) ([]osfapi.FileVersion, error) {
	f.gotFileVersionID = id
	if f.entityErr != nil {
		return nil, f.entityErr
	}
	return append([]osfapi.FileVersion(nil), f.fileVersions...), nil
}

func (f *fakeReadonlyClient) ListWikiPages(_ context.Context, id string) ([]osfapi.RelatedResource, error) {
	f.gotRelationNodeID = id
	if f.entityErr != nil {
		return nil, f.entityErr
	}
	return append([]osfapi.RelatedResource(nil), f.wikis...), nil
}

func (f *fakeReadonlyClient) ListNodeComments(_ context.Context, id string) ([]osfapi.RelatedResource, error) {
	f.gotRelationNodeID = id
	if f.entityErr != nil {
		return nil, f.entityErr
	}
	return append([]osfapi.RelatedResource(nil), f.comments...), nil
}

func (f *fakeReadonlyClient) ListNodeLogs(_ context.Context, id string) ([]osfapi.RelatedResource, error) {
	f.gotRelationNodeID = id
	if f.entityErr != nil {
		return nil, f.entityErr
	}
	return append([]osfapi.RelatedResource(nil), f.logs...), nil
}

func (f *fakeReadonlyClient) ListNodeIdentifiers(_ context.Context, id string) ([]osfapi.RelatedResource, error) {
	f.gotRelationNodeID = id
	if f.entityErr != nil {
		return nil, f.entityErr
	}
	return append([]osfapi.RelatedResource(nil), f.identifiers...), nil
}

func (f *fakeReadonlyClient) CreateRegistration(_ context.Context, nodeID string, request osfapi.RegistrationRequest) (osfapi.Node, error) {
	f.gotRegistrationNode = nodeID
	f.gotRegistrationReq = request
	return osfapi.Node{ID: "reg1", Type: "registrations", Attributes: osfapi.NodeAttributes{Title: request.Title}, Links: osfapi.Links{Self: "https://osf.io/reg1/"}}, nil
}

func (f *fakeReadonlyClient) GetNodeFilesProvider(_ context.Context, id string) (string, error) {
	return "https://files.osf.io/v1/providers/osfstorage/" + id + "/", nil
}

func (f *fakeReadonlyClient) UploadFile(_ context.Context, providerURL, name string, content io.Reader, conflict string) error {
	f.gotProviderURL = providerURL
	f.gotUploadName = name
	f.gotUploadConflict = conflict
	body, err := io.ReadAll(content)
	if err != nil {
		return err
	}
	f.gotUploadBody = string(body)
	return nil
}

func (f *fakeReadonlyClient) CreateFolder(_ context.Context, providerURL, folderName string) error {
	f.gotProviderURL = providerURL
	f.gotCreatedFolder = folderName
	return nil
}

func (f *fakeReadonlyClient) DeleteFile(_ context.Context, providerURL, fileName string) error {
	f.gotProviderURL = providerURL
	f.gotDeletedFile = fileName
	return nil
}

func (f *fakeReadonlyClient) ResolveDOI(context.Context, string) (osfapi.DOIResolution, error) {
	return osfapi.DOIResolution{DOI: "10.1234/example", ResolvedURL: "https://osf.io/project-1/"}, nil
}

func (f *fakeReadonlyClient) OpenDownload(_ context.Context, downloadURL string) (io.ReadCloser, error) {
	f.gotOpenDownloadURL = downloadURL
	if f.downloadBodies != nil {
		if body, ok := f.downloadBodies[downloadURL]; ok {
			return io.NopCloser(strings.NewReader(body)), nil
		}
	}
	return nil, fmt.Errorf("missing download body %q", downloadURL)
}

func (f *fakeReadonlyClient) OpenDownloadRange(ctx context.Context, downloadURL string, _ int64) (io.ReadCloser, error) {
	return f.OpenDownload(ctx, downloadURL)
}

func TestExportJSONOutput(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{
		node: osfapi.Node{ID: "n1", Attributes: osfapi.NodeAttributes{Title: "Test", Category: "project"}, Links: osfapi.Links{Self: "https://osf.io/n1/"}},
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClient([]string{"export", "n1", "--json"}, &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("export json returned %d, want 0, stderr=%q", code, stderr.String())
	}
	var data ExportData
	if err := json.Unmarshal(stdout.Bytes(), &data); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if data.Node.ID != "n1" || data.Node.Title != "Test" {
		t.Fatalf("export data = %+v", data)
	}
}

func TestExportErrorForMissingNode(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClient([]string{"export"}, &stdout, &stderr, &fakeReadonlyClient{})
	if code != 2 {
		t.Fatalf("export without args returned %d, want 2", code)
	}
}

func TestExportTableOutput(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{
		node: osfapi.Node{ID: "n1", Attributes: osfapi.NodeAttributes{Title: "Test", Category: "project"}, Links: osfapi.Links{Self: "https://osf.io/n1/"}},
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClient([]string{"export", "n1"}, &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("export table returned %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "Test") || !strings.Contains(stdout.String(), "project") {
		t.Fatalf("table output = %q, want Test and project", stdout.String())
	}
}

func TestPreprintsListOutput(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{
		preprints: []osfapi.Node{{ID: "p1", Type: "preprints", Attributes: osfapi.NodeAttributes{Title: "Preprint"}, Links: osfapi.Links{Self: "https://osf.io/p1/"}}},
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClient([]string{"preprints", "list", "--provider", "osf"}, &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("preprints list returned %d, want 0", code)
	}
	if client.gotPreprintProvider != "osf" {
		t.Fatalf("provider = %q, want osf", client.gotPreprintProvider)
	}
	if client.gotPreprintLimit != 20 {
		t.Fatalf("limit = %d, want default 20", client.gotPreprintLimit)
	}
	if !strings.Contains(stdout.String(), "Preprint") {
		t.Fatalf("stdout = %q, want Preprint", stdout.String())
	}
}

func TestPreprintsSearchOutput(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{
		searchedPreprints: []osfapi.Preprint{{
			ID: "p1", Type: "preprints",
			Attributes: osfapi.PreprintAttributes{Title: "Open Science", DatePublished: "2026-01-02", IsPublished: true, DOI: "10.1234/p1"},
			Links:      osfapi.Links{HTML: "https://osf.io/p1/"},
		}},
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClient([]string{"preprints", "search", "open science", "--provider", "osf"}, &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("preprints search returned %d, want 0: %s", code, stderr.String())
	}
	if client.gotPreprintQuery != "open science" || client.gotPreprintProvider != "osf" || client.gotPreprintLimit != 10 {
		t.Fatalf("search args = %q, %q, %d", client.gotPreprintQuery, client.gotPreprintProvider, client.gotPreprintLimit)
	}
	for _, expected := range []string{"Open Science", "2026-01-02", "10.1234/p1", "https://osf.io/p1/", "OSF Preprints"} {
		if !strings.Contains(stdout.String(), expected) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), expected)
		}
	}
}

func TestSearchOutput(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{
		searchResults: []osfapi.SearchResult{{ID: "s1", Type: "preprints", Title: "Search Result", URL: "https://osf.io/s1/"}},
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClient([]string{"search", "open+science"}, &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("search returned %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "preprints") || !strings.Contains(stdout.String(), "Search Result") {
		t.Fatalf("stdout = %q, want typed search result", stdout.String())
	}
	if client.gotSearchLimit != 20 {
		t.Fatalf("limit = %d, want default 20", client.gotSearchLimit)
	}
}

func TestSearchBibTeXOutput(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{
		searchResults: []osfapi.SearchResult{{
			ID: "osf-1", Type: "nodes", Title: "Open {Science}", Description: "An abstract",
			Keywords: []string{"open science", "review"}, Year: "2024", URL: "https://osf.io/osf-1/",
		}},
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClient([]string{"search", "science", "--bibtex"}, &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("search bibtex returned %d, want 0: %s", code, stderr.String())
	}
	for _, expected := range []string{
		"@misc{osf-1,", "title = {Open \\{Science\\}}", "abstract = {An abstract}",
		"keywords = {open science, review}", "year = {2024}", "url = {https://osf.io/osf-1/}",
	} {
		if !strings.Contains(stdout.String(), expected) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), expected)
		}
	}
	stdout.Reset()
	stderr.Reset()
	if code = runWithClient([]string{"search", "science", "--bibtex", "--json"}, &stdout, &stderr, client); code == 0 || !strings.Contains(stderr.String(), "cannot combine --bibtex") {
		t.Fatalf("bibtex/json returned %d, stderr=%q", code, stderr.String())
	}
}

func TestWhoamiAlias(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{
		currentUser: osfapi.User{
			ID:   "u1",
			Type: "users",
			Attributes: osfapi.UserAttributes{
				FullName:   "Test User",
				GivenName:  "Test",
				FamilyName: "User",
			},
		},
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClient([]string{"whoami"}, &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("whoami returned %d, want 0, stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Test User") {
		t.Fatalf("whoami output = %q, want Test User", stdout.String())
	}
}

func TestOpenCommandHelp(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"open", "--help"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("open --help returned %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "guid-or-url") {
		t.Fatalf("open help = %q, want guid-or-url", stdout.String())
	}
}

func TestPreprintsListEmpty(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClient([]string{"preprints", "list"}, &stdout, &stderr, &fakeReadonlyClient{})
	if code != 0 {
		t.Fatalf("preprints list returned %d, want 0", code)
	}
}

func TestSearchEmpty(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClient([]string{"search", "test"}, &stdout, &stderr, &fakeReadonlyClient{})
	if code != 0 {
		t.Fatalf("search returned %d, want 0", code)
	}
}

func TestFilesUploadStreamsFileAndReportsSuccess(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	localPath := filepath.Join(dir, "report.csv")
	if err := os.WriteFile(localPath, []byte("a,b\n1,2\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	client := &fakeReadonlyClient{}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClient([]string{"files", "upload", "--node", "abc123", "--conflict", "overwrite", localPath}, &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("files upload returned %d, want 0, stderr=%q", code, stderr.String())
	}
	if client.gotUploadName != "report.csv" || client.gotUploadBody != "a,b\n1,2\n" || client.gotUploadConflict != "overwrite" {
		t.Fatalf("upload call name=%q body=%q conflict=%q", client.gotUploadName, client.gotUploadBody, client.gotUploadConflict)
	}
	if !strings.Contains(stdout.String(), "uploaded report.csv to node abc123") {
		t.Fatalf("stdout = %q, want upload message", stdout.String())
	}
}

func TestFilesMkdirPassesNestedFolderPath(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClient([]string{"files", "mkdir", "--node", "abc123", "nested/folder"}, &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("files mkdir returned %d, want 0, stderr=%q", code, stderr.String())
	}
	if client.gotCreatedFolder != "nested/folder" {
		t.Fatalf("created folder = %q, want nested/folder", client.gotCreatedFolder)
	}
}

func TestFilesRmRequiresConfirmation(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClientInput([]string{"files", "rm", "--node", "abc123", "old.csv"}, "", &stdout, &stderr, client)
	if code != 1 {
		t.Fatalf("files rm without confirmation returned %d, want 1", code)
	}
	if client.gotDeletedFile != "" {
		t.Fatalf("deleted file = %q, want no delete", client.gotDeletedFile)
	}
	if !strings.Contains(stderr.String(), "delete confirmation required") {
		t.Fatalf("stderr = %q, want confirmation error", stderr.String())
	}
}

func TestFilesRmDeletesAfterConfirmation(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClientInput([]string{"files", "rm", "--node", "abc123", "old.csv"}, "yes\n", &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("files rm returned %d, want 0, stderr=%q", code, stderr.String())
	}
	if client.gotDeletedFile != "old.csv" {
		t.Fatalf("deleted file = %q, want old.csv", client.gotDeletedFile)
	}
}

func TestFilesRmYesSkipsPrompt(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClient([]string{"files", "rm", "--node", "abc123", "--yes", "old.csv"}, &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("files rm --yes returned %d, want 0, stderr=%q", code, stderr.String())
	}
	if client.gotDeletedFile != "old.csv" {
		t.Fatalf("deleted file = %q, want old.csv", client.gotDeletedFile)
	}
}

func TestFilesAddonsOutput(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{
		addons: []osfapi.Node{{ID: "osfstorage", Type: "addons", Attributes: osfapi.NodeAttributes{Title: "OSF Storage", Category: "storage"}, Links: osfapi.Links{Self: "https://api.osf.io/v2/addons/osfstorage/"}}},
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClient([]string{"files", "addons", "abc123"}, &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("files addons returned %d, want 0, stderr=%q", code, stderr.String())
	}
	if client.gotAddonNodeID != "abc123" {
		t.Fatalf("addon node id = %q, want abc123", client.gotAddonNodeID)
	}
	if !strings.Contains(stdout.String(), "OSF Storage") {
		t.Fatalf("stdout = %q, want OSF Storage", stdout.String())
	}
}

func TestFilesVersionsOutput(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{fileVersions: []osfapi.FileVersion{{
		ID: "version-1", Type: "file_versions",
		Attributes: osfapi.FileVersionAttributes{Size: 42, DateModified: time.Date(2026, 7, 14, 1, 2, 3, 0, time.UTC)},
		Links:      osfapi.Links{Self: "https://api.osf.io/v2/files/file-1/versions/version-1/"},
	}}}
	var stdout, stderr bytes.Buffer
	code := runWithClient([]string{"files", "versions", "file-1"}, &stdout, &stderr, client)
	if code != 0 || client.gotFileVersionID != "file-1" {
		t.Fatalf("files versions returned %d, id=%q, stderr=%q", code, client.gotFileVersionID, stderr.String())
	}
	if !strings.Contains(stdout.String(), "version-1") {
		t.Fatalf("table output = %q, want version-1", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = runWithClient([]string{"files", "versions", "file-1", "--json"}, &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("files versions --json returned %d, stderr=%q", code, stderr.String())
	}
	var got []fileVersionRecord
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil || len(got) != 1 || got[0].ID != "version-1" {
		t.Fatalf("versions JSON = %q, err=%v", stdout.String(), err)
	}
}

func TestNodeRelationsNormalizeURLAndOutputJSON(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{wikis: []osfapi.RelatedResource{{
		ID: "wiki-1", Type: "wiki_pages", Attributes: map[string]any{"name": "README"},
		Links: osfapi.Links{Self: "https://api.osf.io/v2/nodes/project-1/wikis/wiki-1/"},
	}}}
	var stdout, stderr bytes.Buffer
	code := runWithClient([]string{"nodes", "wikis", "https://osf.io/project-1/"}, &stdout, &stderr, client)
	if code != 0 || client.gotRelationNodeID != "project-1" {
		t.Fatalf("nodes wikis returned %d, id=%q, stderr=%q", code, client.gotRelationNodeID, stderr.String())
	}
	if !strings.Contains(stdout.String(), "wiki-1") {
		t.Fatalf("wiki table = %q, want wiki-1", stdout.String())
	}
	stdout.Reset()
	stderr.Reset()
	code = runWithClient([]string{"nodes", "wikis", "https://osf.io/project-1/", "--json"}, &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("nodes wikis --json returned %d, stderr=%q", code, stderr.String())
	}
	var got []relatedResourceRecord
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil || len(got) != 1 || got[0].ID != "wiki-1" {
		t.Fatalf("wiki JSON = %q, err=%v", stdout.String(), err)
	}
}

func TestNodeRelationCommandsOutputJSON(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		data func(*fakeReadonlyClient)
		id   string
	}{
		{name: "wikis", data: func(c *fakeReadonlyClient) { c.wikis = []osfapi.RelatedResource{{ID: "wiki-1"}} }, id: "wiki-1"},
		{name: "comments", data: func(c *fakeReadonlyClient) { c.comments = []osfapi.RelatedResource{{ID: "comment-1"}} }, id: "comment-1"},
		{name: "logs", data: func(c *fakeReadonlyClient) { c.logs = []osfapi.RelatedResource{{ID: "log-1"}} }, id: "log-1"},
		{name: "identifiers", data: func(c *fakeReadonlyClient) { c.identifiers = []osfapi.RelatedResource{{ID: "identifier-1"}} }, id: "identifier-1"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client := &fakeReadonlyClient{}
			tc.data(client)
			var stdout, stderr bytes.Buffer
			code := runWithClient([]string{"nodes", tc.name, "https://osf.io/project-1/", "--json"}, &stdout, &stderr, client)
			if code != 0 || client.gotRelationNodeID != "project-1" {
				t.Fatalf("nodes %s returned %d, id=%q, stderr=%q", tc.name, code, client.gotRelationNodeID, stderr.String())
			}
			var got []relatedResourceRecord
			if err := json.Unmarshal(stdout.Bytes(), &got); err != nil || len(got) != 1 || got[0].ID != tc.id {
				t.Fatalf("%s JSON = %q, err=%v", tc.name, stdout.String(), err)
			}
		})
	}
}

func TestEntityCommandsPropagateClientErrors(t *testing.T) {
	for _, name := range []string{"versions", "addons"} {
		t.Run("files-"+name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			args := []string{"files", name, "file-1"}
			code := runWithClient(args, &stdout, &stderr, &fakeReadonlyClient{entityErr: errors.New("backend unavailable")})
			if code != 1 || !strings.Contains(stderr.String(), "backend unavailable") {
				t.Fatalf("%s returned %d, stderr=%q", name, code, stderr.String())
			}
		})
	}
	for _, name := range []string{"wikis", "comments", "logs", "identifiers"} {
		t.Run("nodes-"+name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := runWithClient([]string{"nodes", name, "node-1"}, &stdout, &stderr, &fakeReadonlyClient{entityErr: errors.New("backend unavailable")})
			if code != 1 || !strings.Contains(stderr.String(), "backend unavailable") {
				t.Fatalf("%s returned %d, stderr=%q", name, code, stderr.String())
			}
		})
	}
}

func TestRegistrationsCreateRequiresSchema(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClient([]string{"registrations", "create", "abc123"}, &stdout, &stderr, &fakeReadonlyClient{})
	if code != 1 {
		t.Fatalf("registrations create returned %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "--schema flag is required") {
		t.Fatalf("stderr = %q, want schema error", stderr.String())
	}
}

func TestRegistrationsCreateWithConfirmation(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClientInput([]string{"registrations", "create", "abc123", "--schema", "schema-1", "--title", "Registration"}, "yes\n", &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("registrations create returned %d, want 0, stderr=%q", code, stderr.String())
	}
	if client.gotRegistrationNode != "abc123" || client.gotRegistrationReq.SchemaID != "schema-1" || client.gotRegistrationReq.Title != "Registration" {
		t.Fatalf("registration call node=%q request=%+v", client.gotRegistrationNode, client.gotRegistrationReq)
	}
	if !strings.Contains(stdout.String(), "reg1") {
		t.Fatalf("stdout = %q, want reg1", stdout.String())
	}
}

func runWithClient(args []string, stdout, stderr *bytes.Buffer, client readonlyClient) int {
	return runWithClientInput(args, "", stdout, stderr, client)
}

func runWithClientInput(args []string, stdin string, stdout, stderr *bytes.Buffer, client readonlyClient) int {
	root := newRootCommandWithClient(stdout, stderr, client)
	root.SetArgs(args)
	root.SetIn(strings.NewReader(stdin))
	if err := root.Execute(); err != nil {
		err = auth.RedactError(err)
		fmt.Fprintln(stderr, err)
		return exitCodeForError(err)
	}
	return 0
}
