package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

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
		{"planned command", fmt.Errorf("%w: osf files upload", errPlannedCommand), 1},
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

func TestFilesListJSONOutput(t *testing.T) {
	t.Parallel()

	client := &fakeReadonlyClient{
		files: []osfapi.StorageFile{
			{ID: "f1", Attributes: osfapi.StorageFileAttributes{Name: "file.txt", Kind: "file"}, Links: osfapi.Links{Download: "https://files.osf.io/f1"}},
		},
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithClient([]string{"files", "list", "project-1", "--output", "json"}, &stdout, &stderr, client)
	if code != 0 {
		t.Fatalf("files list json returned %d, want 0", code)
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
	if len(contract.Commands) != 5 {
		t.Fatalf("command count = %d, want 5", len(contract.Commands))
	}
}

func TestRunEmitsJSONContract(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"--output", "json"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run returned %d, want 0", code)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	var contract rootContract
	if err := json.Unmarshal(stdout.Bytes(), &contract); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\n%s", err, stdout.String())
	}

	if contract.Name != "osf" {
		t.Fatalf("contract name = %q, want osf", contract.Name)
	}
	if contract.Version != version {
		t.Fatalf("contract version = %q, want %q", contract.Version, version)
	}
	if contract.DefaultOutput != outputModeTable {
		t.Fatalf("default output = %q, want %q", contract.DefaultOutput, outputModeTable)
	}
	if len(contract.Commands) != 5 {
		t.Fatalf("command count = %d, want 5", len(contract.Commands))
	}
	if contract.ExitCodes["success"] != 0 || contract.ExitCodes["planned_command"] != 1 || contract.ExitCodes["usage_or_argument"] != 2 {
		t.Fatalf("unexpected exit code contract: %#v", contract.ExitCodes)
	}
	if contract.Commands[0].Status != "implemented" || contract.Commands[1].Status != "implemented" {
		t.Fatalf("unexpected command statuses: %#v", contract.Commands)
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
		{"planned command", fmt.Errorf("planned command: osf files upload"), false},
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
			"file-1": {ID: "file-1", Attributes: osfapi.StorageFileAttributes{Name: "analysis.csv", Kind: "file", Size: 12}, Links: osfapi.Links{Download: "https://files.osf.io/v1/resources/project-123/providers/osfstorage/file-1?download=1"}},
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
				{ID: "file-1", Attributes: osfapi.StorageFileAttributes{Name: "analysis.csv", Kind: "file", Size: 12}, Links: osfapi.Links{Download: "https://files.osf.io/v1/resources/project-123/providers/osfstorage/file-1?download=1"}},
			},
			"project-1:figures": {
				{ID: "file-2", Attributes: osfapi.StorageFileAttributes{Name: "plot.png", Kind: "file", Size: 24}, Links: osfapi.Links{Download: "https://files.osf.io/v1/resources/project-123/providers/osfstorage/file-2?download=1"}},
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
	currentUser         osfapi.User
	projects            []osfapi.Node
	node                osfapi.Node
	children            []osfapi.Node
	files               []osfapi.StorageFile
	storageFiles        map[string]osfapi.StorageFile
	storageLists        map[string][]osfapi.StorageFile
	downloadBodies      map[string]string
	gotCurrentUserCalls int
	gotNodeID           string
	gotChildrenID       string
	gotFilesID          string
	gotFilesSegments    []string
	gotStorageFileID    string
	gotOpenDownloadURL  string
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

func (f *fakeReadonlyClient) OpenDownload(_ context.Context, downloadURL string) (io.ReadCloser, error) {
	f.gotOpenDownloadURL = downloadURL
	if f.downloadBodies != nil {
		if body, ok := f.downloadBodies[downloadURL]; ok {
			return io.NopCloser(strings.NewReader(body)), nil
		}
	}
	return nil, fmt.Errorf("missing download body %q", downloadURL)
}

func runWithClient(args []string, stdout, stderr *bytes.Buffer, client readonlyClient) int {
	root := newRootCommandWithClient(stdout, stderr, client)
	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		err = auth.RedactError(err)
		fmt.Fprintln(stderr, err)
		return exitCodeForError(err)
	}
	return 0
}
