package mcpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/edithatogo/osf-cli-go/internal/osfapi"
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
		"osf_contributors_list",
		"osf_search",
		"osf_preprints_list",
		"osf_doi_resolve",
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

func TestServerToolInputSchemasMatchPackagedContract(t *testing.T) {
	session := connectTestServer(t, &fakeOSFClient{})

	want := map[string][]string{
		"osf_whoami":            {},
		"osf_projects_list":     {},
		"osf_project_get":       {"id"},
		"osf_components_list":   {"id"},
		"osf_files_list":        {"id", "path"},
		"osf_contributors_list": {"id"},
		"osf_search":            {"query", "limit"},
		"osf_preprints_list":    {"provider", "limit"},
		"osf_doi_resolve":       {"identifier"},
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
	server := New(osf, Options{Version: "test"})
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
	gotNodeID       string
	gotFileSegments []string
	gotQuery        string
	gotProvider     string
	gotLimit        int
	gotDOI          string
	failErr         error
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
		Attributes: osfapi.StorageFileAttributes{Name: "data.csv", Kind: "file", Size: 12},
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
