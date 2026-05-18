package mcpserver

import (
	"context"
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

type fakeOSFClient struct {
	gotNodeID       string
	gotFileSegments []string
}

func (f *fakeOSFClient) CurrentUser(context.Context) (osfapi.User, error) {
	return osfapi.User{
		ID:         "user-1",
		Type:       "users",
		Attributes: osfapi.UserAttributes{FullName: "Test User"},
		Links:      osfapi.Links{Self: "https://osf.io/user-1/"},
	}, nil
}

func (f *fakeOSFClient) ListCurrentUserProjects(context.Context) ([]osfapi.Node, error) {
	return []osfapi.Node{node("project-1", "Alpha")}, nil
}

func (f *fakeOSFClient) GetNode(_ context.Context, id string) (osfapi.Node, error) {
	f.gotNodeID = id
	return node(id, "Alpha"), nil
}

func (f *fakeOSFClient) ListNodeChildren(_ context.Context, id string) ([]osfapi.Node, error) {
	f.gotNodeID = id
	return []osfapi.Node{node("component-1", "Component")}, nil
}

func (f *fakeOSFClient) ListNodeContributors(_ context.Context, id string) ([]osfapi.Contributor, error) {
	f.gotNodeID = id
	return []osfapi.Contributor{{
		ID:         "contrib-1",
		Type:       "contributors",
		Attributes: osfapi.ContributorAttributes{FullName: "Contributor", Bibliographic: true, Permission: "admin"},
		Links:      osfapi.Links{Self: "https://osf.io/users/contrib-1/"},
	}}, nil
}

func (f *fakeOSFClient) ListStorageFiles(_ context.Context, id string, segments ...string) ([]osfapi.StorageFile, error) {
	f.gotNodeID = id
	f.gotFileSegments = append([]string(nil), segments...)
	return []osfapi.StorageFile{{
		ID:         "file-1",
		Type:       "files",
		Attributes: osfapi.StorageFileAttributes{Name: "data.csv", Kind: "file", Size: 12},
		Links:      osfapi.Links{Self: "https://files.osf.io/file-1", Download: "https://files.osf.io/file-1?download=1"},
	}}, nil
}

func node(id, title string) osfapi.Node {
	return osfapi.Node{
		ID:         id,
		Type:       "nodes",
		Attributes: osfapi.NodeAttributes{Title: title, Category: "project"},
		Links:      osfapi.Links{Self: "https://osf.io/" + id + "/"},
	}
}
