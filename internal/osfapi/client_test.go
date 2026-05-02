package osfapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestCurrentUser(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer token-123" {
			t.Fatalf("authorization header = %q", got)
		}
		if got := r.URL.Path; got != "/v2/users/me/" {
			t.Fatalf("path = %q", got)
		}
		writeFixture(t, w, "user_me.json")
	}))
	defer srv.Close()

	client, err := New(srv.URL, WithHTTPClient(srv.Client()), WithBearerToken("token-123"))
	if err != nil {
		t.Fatal(err)
	}

	user, err := client.CurrentUser(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if user.ID != "u1" || user.Attributes.FullName != "Ada Lovelace" {
		t.Fatalf("unexpected user: %+v", user)
	}
}

func TestGetNodeAndChildrenPagination(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/nodes/project-123/":
			writeFixture(t, w, "node_project.json")
		case "/v2/nodes/project-123/children/":
			if r.URL.RawQuery == "" {
				writeFixture(t, w, "node_children_page1.json")
				return
			}
			if r.URL.RawQuery == "page=2" {
				writeFixture(t, w, "node_children_page2.json")
				return
			}
			t.Fatalf("unexpected query: %s", r.URL.RawQuery)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	client, err := New(srv.URL, WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatal(err)
	}

	node, err := client.GetNode(context.Background(), "project-123")
	if err != nil {
		t.Fatal(err)
	}
	if node.Attributes.Title != "Project Alpha" {
		t.Fatalf("node title = %q", node.Attributes.Title)
	}

	children, err := client.ListNodeChildren(context.Background(), "project-123")
	if err != nil {
		t.Fatal(err)
	}
	if len(children) != 3 {
		t.Fatalf("children length = %d", len(children))
	}
	if children[0].Attributes.Title != "Child One" || children[2].Attributes.Title != "Child Three" {
		t.Fatalf("unexpected children: %+v", children)
	}
}

func TestListContributorsAndStorageFiles(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/nodes/project-123/contributors/":
			if r.URL.RawQuery == "" {
				writeFixture(t, w, "contributors_page1.json")
				return
			}
			if r.URL.RawQuery == "page=2" {
				writeFixture(t, w, "contributors_page2.json")
				return
			}
			t.Fatalf("unexpected query: %s", r.URL.RawQuery)
		case "/v2/nodes/project-123/files/osfstorage/":
			if r.URL.RawQuery == "" {
				writeFixture(t, w, "storage_files_page1.json")
				return
			}
			if r.URL.RawQuery == "page=2" {
				writeFixture(t, w, "storage_files_page2.json")
				return
			}
			t.Fatalf("unexpected query: %s", r.URL.RawQuery)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	client, err := New(srv.URL, WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatal(err)
	}

	contributors, err := client.ListNodeContributors(context.Background(), "project-123")
	if err != nil {
		t.Fatal(err)
	}
	if len(contributors) != 3 || contributors[0].Attributes.FullName != "Taylor Researcher" {
		t.Fatalf("unexpected contributors: %+v", contributors)
	}

	files, err := client.ListStorageFiles(context.Background(), "project-123")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 3 {
		t.Fatalf("files length = %d", len(files))
	}
	if got := files[0].DownloadURL(); got != "https://files.osf.io/v1/resources/project-123/providers/osfstorage/file-1?download=1" {
		t.Fatalf("download url = %q", got)
	}
}

func TestAPIError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		writeFixture(t, w, "error_not_found.json")
	}))
	defer srv.Close()

	client, err := New(srv.URL, WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.GetNode(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("error type = %T", err)
	}
	if apiErr.StatusCode != http.StatusNotFound || apiErr.Path != "/v2/nodes/missing/" {
		t.Fatalf("unexpected api error: %+v", apiErr)
	}
	if apiErr.Title != "Not Found" || apiErr.Detail != "Node not found." {
		t.Fatalf("unexpected api error detail: %+v", apiErr)
	}
}

func writeFixture(t *testing.T, w http.ResponseWriter, name string) {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/vnd.api+json")
	_, _ = w.Write(body)
}
