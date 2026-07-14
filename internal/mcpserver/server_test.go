package mcpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/edithatogo/osf-cli-go/internal/observability"
	"github.com/edithatogo/osf-cli-go/internal/osfapi"
	"github.com/edithatogo/osf-cli-go/internal/zenodooai"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestServerExposesReadOnlyTools(t *testing.T) {
	session := connectTestServer(t, &fakeOSFClient{})

	var names []string
	for tool, err := range session.Tools(t.Context(), nil) {
		if err != nil {
			t.Fatalf("Tools returned error: %v", err)
		}
		names = append(names, tool.Name)
	}

	want := []string{
		"osf_whoami",
		"osf_projects_list",
		"osf_project_get",
		"osf_components_list",
		"osf_files_list",
		"osf_file_versions_list",
		"osf_addons_list",
		"osf_wikis_list",
		"osf_comments_list",
		"osf_logs_list",
		"osf_identifiers_list",
		"osf_contributors_list",
		"osf_search",
		"osf_preprints_list",
		"osf_preprints_search",
		"osf_doi_resolve",
		"zenodo_oai_records_list",
		"zenodo_oai_sets_list",
		"zenodo_oai_formats_list",
	}
	got := map[string]bool{}
	for _, name := range names {
		got[name] = true
	}
	for _, name := range want {
		if !got[name] {
			t.Fatalf("missing tool %q; all tools: %v", name, names)
		}
	}
}

func TestMCPToolErrorsEmitRedactedEvents(t *testing.T) {
	var events strings.Builder
	server := &Server{events: observability.NewJSONEmitter(&events, observability.LevelInfo)}
	_, _, err := server.Search(context.Background(), nil, SearchInput{Query: "  "})
	if err == nil {
		t.Fatal("Search returned nil error for empty query")
	}
	if !strings.Contains(events.String(), `"name":"mcp.tool.error"`) || !strings.Contains(events.String(), `"class":"validation"`) {
		t.Fatalf("events=%q", events.String())
	}
}

func TestServerToolInputSchemasMatchPackagedContract(t *testing.T) {
	session := connectTestServer(t, &fakeOSFClient{})

	want := map[string][]string{
		"osf_whoami":              {},
		"osf_projects_list":       {},
		"osf_project_get":         {"id"},
		"osf_components_list":     {"id"},
		"osf_files_list":          {"id", "path"},
		"osf_file_versions_list":  {"id"},
		"osf_addons_list":         {"id"},
		"osf_wikis_list":          {"id"},
		"osf_comments_list":       {"id"},
		"osf_logs_list":           {"id"},
		"osf_identifiers_list":    {"id"},
		"osf_contributors_list":   {"id"},
		"osf_search":              {"query", "limit"},
		"osf_preprints_list":      {"provider", "limit"},
		"osf_preprints_search":    {"query", "provider", "limit"},
		"osf_doi_resolve":         {"identifier"},
		"zenodo_oai_records_list": {"metadataPrefix", "set", "from", "until", "resumptionToken"},
		"zenodo_oai_sets_list":    {},
		"zenodo_oai_formats_list": {"identifier"},
	}
	for tool, err := range session.Tools(t.Context(), nil) {
		if err != nil {
			t.Fatalf("Tools returned error: %v", err)
		}
		properties := schemaProperties(t, tool.InputSchema)
		wantProperties, ok := want[tool.Name]
		if !ok {
			t.Fatalf("unexpected tool %q", tool.Name)
		}
		if diff := missingOrExtra(properties, wantProperties); diff != "" {
			t.Fatalf("tool %s schema properties mismatch: %s", tool.Name, diff)
		}
		delete(want, tool.Name)
	}
	for name := range want {
		t.Fatalf("missing tool %q", name)
	}
}

func TestServerToolSchemasMatchCompatibilityFixture(t *testing.T) {
	fixture, err := os.ReadFile("testdata/compatibility/mcp-tools.json")
	if err != nil {
		t.Fatalf("read compatibility fixture: %v", err)
	}
	var want struct {
		SchemaVersion int `json:"schemaVersion"`
		Tools         []struct {
			Name       string   `json:"name"`
			Properties []string `json:"properties"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(fixture, &want); err != nil {
		t.Fatalf("decode compatibility fixture: %v", err)
	}
	if want.SchemaVersion != 1 {
		t.Fatalf("schema version = %d, want 1", want.SchemaVersion)
	}
	wantTools := map[string][]string{}
	for _, tool := range want.Tools {
		wantTools[tool.Name] = append([]string{}, tool.Properties...)
		sort.Strings(wantTools[tool.Name])
	}

	session := connectTestServer(t, &fakeOSFClient{})
	gotTools := map[string][]string{}
	for tool, err := range session.Tools(t.Context(), nil) {
		if err != nil {
			t.Fatalf("Tools returned error: %v", err)
		}
		properties := schemaProperties(t, tool.InputSchema)
		got := make([]string, 0, len(properties))
		for property := range properties {
			got = append(got, property)
		}
		sort.Strings(got)
		gotTools[tool.Name] = got
	}
	gotJSON, _ := json.Marshal(gotTools)
	wantJSON, _ := json.Marshal(wantTools)
	if string(gotJSON) != string(wantJSON) {
		t.Fatalf("MCP compatibility contract changed:\n got: %s\nwant: %s", gotJSON, wantJSON)
	}
}

func TestProjectGetAcceptsOSFURL(t *testing.T) {
	client := &fakeOSFClient{}
	session := connectTestServer(t, client)

	result, err := session.CallTool(t.Context(), &mcp.CallToolParams{
		Name:      "osf_project_get",
		Arguments: map[string]any{"id": "https://osf.io/project-1/files/osfstorage"},
	})
	if err != nil {
		t.Fatalf("CallTool returned error: %v", err)
	}
	if result.IsError {
		t.Fatalf("tool returned error: %#v", result.Content)
	}
	if client.gotNodeID != "project-1" {
		t.Fatalf("GetNode id = %q, want project-1", client.gotNodeID)
	}
	content, ok := result.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("structured content is %T, want map[string]any", result.StructuredContent)
	}
	node, ok := content["node"].(map[string]any)
	if !ok {
		t.Fatalf("structured node is %T, want map[string]any", content["node"])
	}
	if got := node["title"]; got != "Alpha" {
		t.Fatalf("structured node title = %v, want Alpha", got)
	}
}

func TestFilesListSplitsStoragePath(t *testing.T) {
	client := &fakeOSFClient{}
	session := connectTestServer(t, client)

	result, err := session.CallTool(t.Context(), &mcp.CallToolParams{
		Name:      "osf_files_list",
		Arguments: map[string]any{"id": "project-1", "path": "data/raw"},
	})
	if err != nil {
		t.Fatalf("CallTool returned error: %v", err)
	}
	if result.IsError {
		t.Fatalf("tool returned error: %#v", result.Content)
	}
	if got, want := client.gotFileSegments, []string{"data", "raw"}; len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("segments = %#v, want %#v", got, want)
	}
	content, ok := result.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("structured content is %T, want map[string]any", result.StructuredContent)
	}
	files, ok := content["files"].([]any)
	if !ok || len(files) != 1 {
		t.Fatalf("structured files = %#v, want one file", content["files"])
	}
	file, ok := files[0].(map[string]any)
	if !ok || file["md5"] != "abc123" {
		t.Fatalf("structured file = %#v, want md5", files[0])
	}
}

func TestFilesListRejectsTraversal(t *testing.T) {
	session := connectTestServer(t, &fakeOSFClient{})

	result, err := session.CallTool(t.Context(), &mcp.CallToolParams{
		Name:      "osf_files_list",
		Arguments: map[string]any{"id": "project-1", "path": "../outside"},
	})
	if err != nil {
		t.Fatalf("CallTool returned error: %v", err)
	}
	if !result.IsError {
		t.Fatal("tool succeeded, want MCP tool error")
	}
}

func TestEntityCoverageToolsReturnStructuredResults(t *testing.T) {
	client := &fakeOSFClient{}
	session := connectTestServer(t, client)
	cases := []struct {
		name string
		key  string
		id   string
	}{
		{name: "osf_file_versions_list", key: "versions", id: "version-1"},
		{name: "osf_addons_list", key: "nodes", id: "addon-1"},
		{name: "osf_wikis_list", key: "resources", id: "wiki-1"},
		{name: "osf_comments_list", key: "resources", id: "comment-1"},
		{name: "osf_logs_list", key: "resources", id: "log-1"},
		{name: "osf_identifiers_list", key: "resources", id: "identifier-1"},
	}
	for _, tc := range cases {
		result, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: tc.name, Arguments: map[string]any{"id": "project-1"}})
		if err != nil || result.IsError {
			t.Fatalf("%s failed: result=%#v err=%v", tc.name, result, err)
		}
		content, ok := result.StructuredContent.(map[string]any)
		if !ok {
			t.Fatalf("%s structured content = %T", tc.name, result.StructuredContent)
		}
		items, ok := content[tc.key].([]any)
		if !ok || len(items) != 1 {
			t.Fatalf("%s content[%q] = %#v", tc.name, tc.key, content[tc.key])
		}
		item, ok := items[0].(map[string]any)
		if !ok || item["id"] != tc.id {
			t.Fatalf("%s item = %#v, want id %s", tc.name, items[0], tc.id)
		}
	}
}

func TestEntityCoverageToolsPropagateClientErrors(t *testing.T) {
	cases := []string{"osf_file_versions_list", "osf_addons_list", "osf_wikis_list", "osf_comments_list", "osf_logs_list", "osf_identifiers_list"}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			session := connectTestServer(t, &fakeOSFClient{failErr: errors.New("backend unavailable")})
			result, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: name, Arguments: map[string]any{"id": "node-1"}})
			if err != nil {
				t.Fatalf("CallTool returned error: %v", err)
			}
			if !result.IsError || !strings.Contains(contentText(result.Content), "backend unavailable") {
				t.Fatalf("result = %#v, want backend error", result)
			}
		})
	}
}

func TestEntityCoverageToolsRejectMissingIDs(t *testing.T) {
	for _, name := range []string{"osf_file_versions_list", "osf_addons_list", "osf_wikis_list", "osf_comments_list", "osf_logs_list", "osf_identifiers_list"} {
		t.Run(name, func(t *testing.T) {
			session := connectTestServer(t, &fakeOSFClient{})
			result, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: name, Arguments: map[string]any{"id": " "}})
			if err != nil || !result.IsError {
				t.Fatalf("%s result=%#v err=%v, want missing-id tool error", name, result, err)
			}
		})
	}
}

func TestSearchRequiresQueryAndPassesLimit(t *testing.T) {
	client := &fakeOSFClient{}
	session := connectTestServer(t, client)
	result, err := session.CallTool(t.Context(), &mcp.CallToolParams{
		Name: "osf_search", Arguments: map[string]any{"query": " reproducibility ", "limit": 10},
	})
	if err != nil || result.IsError {
		t.Fatalf("search failed: result=%#v err=%v", result, err)
	}
	if client.gotQuery != "reproducibility" || client.gotLimit != 10 {
		t.Fatalf("search args = %q, %d; want reproducibility, 10", client.gotQuery, client.gotLimit)
	}
	result, err = session.CallTool(t.Context(), &mcp.CallToolParams{
		Name: "osf_search", Arguments: map[string]any{"query": " "},
	})
	if err != nil || !result.IsError {
		t.Fatalf("blank search query result=%#v err=%v, want tool error", result, err)
	}
}

func TestPreprintsPassProviderAndLimit(t *testing.T) {
	client := &fakeOSFClient{}
	session := connectTestServer(t, client)
	result, err := session.CallTool(t.Context(), &mcp.CallToolParams{
		Name: "osf_preprints_list", Arguments: map[string]any{"provider": "osf", "limit": 5},
	})
	if err != nil || result.IsError {
		t.Fatalf("preprints failed: result=%#v err=%v", result, err)
	}
	if client.gotProvider != "osf" || client.gotLimit != 5 {
		t.Fatalf("preprint args = %q, %d; want osf, 5", client.gotProvider, client.gotLimit)
	}
}

func TestPreprintSearchRequiresQueryAndPassesFilters(t *testing.T) {
	client := &fakeOSFClient{}
	session := connectTestServer(t, client)
	result, err := session.CallTool(t.Context(), &mcp.CallToolParams{
		Name: "osf_preprints_search", Arguments: map[string]any{"query": "open science", "provider": "osf", "limit": 7},
	})
	if err != nil || result.IsError {
		t.Fatalf("preprint search failed: result=%#v err=%v", result, err)
	}
	if client.gotPreprintQuery != "open science" || client.gotProvider != "osf" || client.gotLimit != 7 {
		t.Fatalf("preprint search args = %q, %q, %d", client.gotPreprintQuery, client.gotProvider, client.gotLimit)
	}
	result, err = session.CallTool(t.Context(), &mcp.CallToolParams{
		Name: "osf_preprints_search", Arguments: map[string]any{"query": " "},
	})
	if err != nil || !result.IsError {
		t.Fatalf("blank preprint search result=%#v err=%v, want tool error", result, err)
	}
}

func TestDOIResolveRequiresIdentifierAndReturnsResolution(t *testing.T) {
	client := &fakeOSFClient{}
	session := connectTestServer(t, client)
	result, err := session.CallTool(t.Context(), &mcp.CallToolParams{
		Name: "osf_doi_resolve", Arguments: map[string]any{"identifier": "10.1234/example"},
	})
	if err != nil || result.IsError {
		t.Fatalf("DOI resolve failed: result=%#v err=%v", result, err)
	}
	if client.gotDOI != "10.1234/example" {
		t.Fatalf("DOI = %q, want 10.1234/example", client.gotDOI)
	}
	result, err = session.CallTool(t.Context(), &mcp.CallToolParams{
		Name: "osf_doi_resolve", Arguments: map[string]any{"identifier": " "},
	})
	if err != nil || !result.IsError {
		t.Fatalf("blank DOI result=%#v err=%v, want tool error", result, err)
	}
}

func TestToolFailureRedactsSecretMaterial(t *testing.T) {
	session := connectTestServer(t, &fakeOSFClient{
		failErr: errors.New("request failed Authorization: Bearer osf_live_token_abc123def456ghi789xyz OSF_PASSWORD=password-123"),
	})

	result, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: "osf_whoami"})
	if err != nil {
		t.Fatalf("CallTool returned error: %v", err)
	}
	if !result.IsError {
		t.Fatal("tool succeeded, want MCP tool error")
	}
	rendered := contentText(result.Content)
	for _, forbidden := range []string{"osf_live_token_abc123def456ghi789xyz", "password-123"} {
		if strings.Contains(rendered, forbidden) {
			t.Fatalf("tool error leaked %q in %#v", forbidden, result.Content)
		}
	}
	if !strings.Contains(rendered, "[REDACTED]") {
		t.Fatalf("tool error = %#v, want redaction marker", result.Content)
	}
}

func connectTestServer(t *testing.T, osf OSFClient) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()
	server := New(osf, Options{Version: "test", ZenodoOAI: fakeZenodoOAI{}})
	t1, t2 := mcp.NewInMemoryTransports()
	if _, err := server.Connect(ctx, t1, nil); err != nil {
		t.Fatalf("server Connect: %v", err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "test"}, nil)
	session, err := client.Connect(ctx, t2, nil)
	if err != nil {
		t.Fatalf("client Connect: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })
	return session
}

func schemaProperties(t *testing.T, schema any) map[string]bool {
	t.Helper()
	data, err := json.Marshal(schema)
	if err != nil {
		t.Fatalf("marshal schema: %v", err)
	}
	var decoded struct {
		Properties map[string]any `json:"properties"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal schema %s: %v", string(data), err)
	}
	properties := map[string]bool{}
	for name := range decoded.Properties {
		properties[name] = true
	}
	return properties
}

func missingOrExtra(got map[string]bool, want []string) string {
	wantSet := map[string]bool{}
	for _, name := range want {
		wantSet[name] = true
		if !got[name] {
			return "missing " + name
		}
	}
	for name := range got {
		if !wantSet[name] {
			return "unexpected " + name
		}
	}
	return ""
}

func contentText(content []mcp.Content) string {
	var b strings.Builder
	for _, item := range content {
		switch value := item.(type) {
		case *mcp.TextContent:
			b.WriteString(value.Text)
		default:
			_, _ = fmt.Fprint(&b, item)
		}
	}
	return b.String()
}

type fakeOSFClient struct {
	gotNodeID        string
	gotFileSegments  []string
	gotQuery         string
	gotPreprintQuery string
	gotProvider      string
	gotLimit         int
	gotDOI           string
	gotRelatedID     string
	failErr          error
}

type fakeZenodoOAI struct{ failErr error }

func (fake fakeZenodoOAI) ListRecords(_ context.Context, request zenodooai.Request) (zenodooai.Page, error) {
	if fake.failErr != nil {
		return zenodooai.Page{}, fake.failErr
	}
	return zenodooai.Page{Records: []zenodooai.Record{{Header: zenodooai.Header{Identifier: "oai:zenodo.org:1001", Datestamp: "2026-07-15"}, Provenance: zenodooai.Provenance{MetadataPrefix: request.MetadataPrefix}}}}, nil
}

func (fake fakeZenodoOAI) ListSets(context.Context) ([]zenodooai.Set, error) {
	if fake.failErr != nil {
		return nil, fake.failErr
	}
	return []zenodooai.Set{{Spec: "user-demo", Name: "Demo"}}, nil
}

func (fake fakeZenodoOAI) ListMetadataFormats(context.Context, string) ([]zenodooai.MetadataFormat, error) {
	if fake.failErr != nil {
		return nil, fake.failErr
	}
	return []zenodooai.MetadataFormat{{Prefix: "oai_dc"}}, nil
}

func (f *fakeOSFClient) ListFileVersions(_ context.Context, id string) ([]osfapi.FileVersion, error) {
	f.gotRelatedID = id
	if f.failErr != nil {
		return nil, f.failErr
	}
	return []osfapi.FileVersion{{ID: "version-1", Type: "file_versions", Attributes: osfapi.FileVersionAttributes{Size: 42}, Links: osfapi.Links{Self: "https://api.osf.io/v2/files/file-1/versions/1/"}}}, nil
}

func (f *fakeOSFClient) ListNodeAddons(_ context.Context, id string) ([]osfapi.Node, error) {
	f.gotRelatedID = id
	if f.failErr != nil {
		return nil, f.failErr
	}
	return []osfapi.Node{{ID: "addon-1", Type: "files", Attributes: osfapi.NodeAttributes{Title: "OSF Storage"}}}, nil
}

func (f *fakeOSFClient) ListWikiPages(_ context.Context, id string) ([]osfapi.RelatedResource, error) {
	f.gotRelatedID = id
	if f.failErr != nil {
		return nil, f.failErr
	}
	return []osfapi.RelatedResource{{ID: "wiki-1", Type: "wikis", Attributes: map[string]any{"name": "README"}}}, nil
}

func (f *fakeOSFClient) ListNodeComments(_ context.Context, id string) ([]osfapi.RelatedResource, error) {
	f.gotRelatedID = id
	if f.failErr != nil {
		return nil, f.failErr
	}
	return []osfapi.RelatedResource{{ID: "comment-1", Type: "comments", Attributes: map[string]any{"content": "hello"}}}, nil
}

func (f *fakeOSFClient) ListNodeLogs(_ context.Context, id string) ([]osfapi.RelatedResource, error) {
	f.gotRelatedID = id
	if f.failErr != nil {
		return nil, f.failErr
	}
	return []osfapi.RelatedResource{{ID: "log-1", Type: "logs", Attributes: map[string]any{"action": "view"}}}, nil
}

func (f *fakeOSFClient) ListNodeIdentifiers(_ context.Context, id string) ([]osfapi.RelatedResource, error) {
	f.gotRelatedID = id
	if f.failErr != nil {
		return nil, f.failErr
	}
	return []osfapi.RelatedResource{{ID: "identifier-1", Type: "identifiers", Attributes: map[string]any{"value": "10.1234/example"}}}, nil
}

func (f *fakeOSFClient) CurrentUser(context.Context) (osfapi.User, error) {
	if f.failErr != nil {
		return osfapi.User{}, f.failErr
	}
	return osfapi.User{
		ID:         "user-1",
		Type:       "users",
		Attributes: osfapi.UserAttributes{FullName: "Test User"},
		Links:      osfapi.Links{Self: "https://osf.io/user-1/"},
	}, nil
}

func (f *fakeOSFClient) ListCurrentUserProjects(context.Context) ([]osfapi.Node, error) {
	if f.failErr != nil {
		return nil, f.failErr
	}
	return []osfapi.Node{node("project-1", "Alpha")}, nil
}

func (f *fakeOSFClient) GetNode(_ context.Context, id string) (osfapi.Node, error) {
	if f.failErr != nil {
		return osfapi.Node{}, f.failErr
	}
	f.gotNodeID = id
	return node(id, "Alpha"), nil
}

func (f *fakeOSFClient) ListNodeChildren(_ context.Context, id string) ([]osfapi.Node, error) {
	if f.failErr != nil {
		return nil, f.failErr
	}
	f.gotNodeID = id
	return []osfapi.Node{node("component-1", "Component")}, nil
}

func (f *fakeOSFClient) ListNodeContributors(_ context.Context, id string) ([]osfapi.Contributor, error) {
	if f.failErr != nil {
		return nil, f.failErr
	}
	f.gotNodeID = id
	return []osfapi.Contributor{{
		ID:         "contrib-1",
		Type:       "contributors",
		Attributes: osfapi.ContributorAttributes{FullName: "Contributor", Bibliographic: true, Permission: "admin"},
		Links:      osfapi.Links{Self: "https://osf.io/users/contrib-1/"},
	}}, nil
}

func (f *fakeOSFClient) ListStorageFiles(_ context.Context, id string, segments ...string) ([]osfapi.StorageFile, error) {
	if f.failErr != nil {
		return nil, f.failErr
	}
	f.gotNodeID = id
	f.gotFileSegments = append([]string(nil), segments...)
	return []osfapi.StorageFile{{
		ID:         "file-1",
		Type:       "files",
		Attributes: osfapi.StorageFileAttributes{Name: "data.csv", Kind: "file", Size: 12, Extra: osfapi.StorageFileExtra{Hashes: osfapi.StorageFileHashes{MD5: "abc123"}}},
		Links:      osfapi.Links{Self: "https://files.osf.io/file-1", Download: "https://files.osf.io/file-1?download=1"},
	}}, nil
}

func (f *fakeOSFClient) SearchOSF(_ context.Context, query string, limit ...int) ([]osfapi.SearchResult, error) {
	if f.failErr != nil {
		return nil, f.failErr
	}
	f.gotQuery = query
	f.gotLimit = limit[0]
	return []osfapi.SearchResult{{ID: "project-1", Type: "nodes", Title: "Alpha", URL: "https://osf.io/project-1/"}}, nil
}

func (f *fakeOSFClient) ListPreprints(_ context.Context, provider string, limit ...int) ([]osfapi.Node, error) {
	if f.failErr != nil {
		return nil, f.failErr
	}
	f.gotProvider = provider
	f.gotLimit = limit[0]
	return []osfapi.Node{node("preprint-1", "Preprint")}, nil
}

func (f *fakeOSFClient) SearchPreprints(_ context.Context, query, provider string, limit ...int) ([]osfapi.Preprint, error) {
	if f.failErr != nil {
		return nil, f.failErr
	}
	f.gotPreprintQuery = query
	f.gotProvider = provider
	f.gotLimit = limit[0]
	return []osfapi.Preprint{{
		ID: "preprint-1", Type: "preprints",
		Attributes: osfapi.PreprintAttributes{Title: "Open Science", DOI: "10.1234/preprint-1", IsPublished: true},
		Links:      osfapi.Links{HTML: "https://osf.io/preprint-1/"},
	}}, nil
}

func (f *fakeOSFClient) ResolveDOI(_ context.Context, identifier string) (osfapi.DOIResolution, error) {
	if f.failErr != nil {
		return osfapi.DOIResolution{}, f.failErr
	}
	f.gotDOI = identifier
	return osfapi.DOIResolution{DOI: identifier, ResolvedURL: "https://osf.io/project-1/"}, nil
}

func node(id, title string) osfapi.Node {
	return osfapi.Node{
		ID:         id,
		Type:       "nodes",
		Attributes: osfapi.NodeAttributes{Title: title, Category: "project"},
		Links:      osfapi.Links{Self: "https://osf.io/" + id + "/"},
	}
}
