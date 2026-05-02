package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"osf-cli-go/internal/auth"
	"osf-cli-go/internal/osfapi"
)

func TestRunPrintsHelpWithoutArgs(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run(nil, &stdout, &stderr)

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
	if len(contract.Commands) != 4 {
		t.Fatalf("command count = %d, want 4", len(contract.Commands))
	}
	if contract.ExitCodes["success"] != 0 || contract.ExitCodes["planned_command"] != 1 || contract.ExitCodes["usage_or_argument"] != 2 {
		t.Fatalf("unexpected exit code contract: %#v", contract.ExitCodes)
	}
	if contract.Commands[0].Status != "implemented" || contract.Commands[1].Status != "implemented" {
		t.Fatalf("unexpected command statuses: %#v", contract.Commands)
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

type fakeReadonlyClient struct {
	currentUser         osfapi.User
	projects            []osfapi.Node
	node                osfapi.Node
	children            []osfapi.Node
	files               []osfapi.StorageFile
	gotCurrentUserCalls int
	gotNodeID           string
	gotChildrenID       string
	gotFilesID          string
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

func (f *fakeReadonlyClient) ListStorageFiles(_ context.Context, id string) ([]osfapi.StorageFile, error) {
	f.gotFilesID = id
	return append([]osfapi.StorageFile(nil), f.files...), nil
}

func runWithClient(args []string, stdout, stderr *bytes.Buffer, client readonlyClient) int {
	root := newRootCommandWithClient(stdout, stderr, client)
	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		return exitCodeForError(err)
	}
	return 0
}
