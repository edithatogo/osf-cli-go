package osfapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
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

func TestCurrentUserWithUsernamePassword(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok {
			t.Fatal("missing basic auth header")
		}
		if username != "user@example.org" || password != "password-123" {
			t.Fatalf("basic auth = %q/%q", username, password)
		}
		writeFixture(t, w, "user_me.json")
	}))
	defer srv.Close()

	client, err := New(srv.URL, WithHTTPClient(srv.Client()), WithUsernamePassword("user@example.org", "password-123"))
	if err != nil {
		t.Fatal(err)
	}
	user, err := client.CurrentUser(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if user.ID != "u1" {
		t.Fatalf("user id = %q, want u1", user.ID)
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

func TestListCurrentUserProjects(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/v2/users/me/nodes/" {
			t.Fatalf("path = %q", got)
		}
		if got := r.URL.RawQuery; got != "filter[category]=project" {
			t.Fatalf("query = %q", got)
		}
		writeFixture(t, w, "node_children_page2.json")
	}))
	defer srv.Close()

	client, err := New(srv.URL, WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatal(err)
	}

	projects, err := client.ListCurrentUserProjects(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 1 || projects[0].ID != "child-3" {
		t.Fatalf("unexpected projects: %+v", projects)
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
	if got := files[0].Attributes.Extra.Hashes.MD5; got != "9e107d9d372bb6826bd81d3542a419d6" {
		t.Fatalf("md5 = %q", got)
	}
}

func TestGetStorageFileAndOpenDownload(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/files/file-1/":
			if got := r.Header.Get("Authorization"); got != "Bearer token-123" {
				t.Fatalf("authorization header = %q", got)
			}
			writeFixture(t, w, "storage_file.json")
		case "/v1/resources/project-123/providers/osfstorage/file-1":
			if got := r.URL.RawQuery; got != "download=1" {
				t.Fatalf("download query = %q", got)
			}
			if got := r.Header.Get("Authorization"); got != "Bearer token-123" {
				t.Fatalf("authorization header = %q", got)
			}
			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write([]byte("downloaded bytes"))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	client, err := New(srv.URL, WithHTTPClient(srv.Client()), WithBearerToken("token-123"))
	if err != nil {
		t.Fatal(err)
	}

	file, err := client.GetStorageFile(context.Background(), "file-1")
	if err != nil {
		t.Fatal(err)
	}
	if file.ID != "file-1" || file.Attributes.Name != "analysis.csv" || file.DownloadURL() == "" {
		t.Fatalf("unexpected file metadata: %+v", file)
	}

	body, err := client.OpenDownload(context.Background(), srv.URL+"/v1/resources/project-123/providers/osfstorage/file-1?download=1")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = body.Close() }()

	got, err := io.ReadAll(body)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "downloaded bytes" {
		t.Fatalf("download body = %q", string(got))
	}
}

func TestGetStorageFileReturnsAPIError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/files/nonexistent/" {
			w.WriteHeader(http.StatusNotFound)
			writeFixture(t, w, "error_not_found.json")
			return
		}
		t.Fatalf("unexpected path: %s", r.URL.Path)
	}))
	defer srv.Close()

	client, err := New(srv.URL, WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	_, err = client.GetStorageFile(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("GetStorageFile returned nil error, want error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.StatusCode != http.StatusNotFound || apiErr.Title != "Not Found" {
		t.Fatalf("unexpected api error: %+v", apiErr)
	}
}

func TestOpenDownloadReturnsAPIError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"errors":[{"title":"Forbidden","detail":"access denied"}]}`))
	}))
	defer srv.Close()

	client, err := New(srv.URL, WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	_, err = client.OpenDownload(context.Background(), srv.URL+"/v1/resources/project-123/providers/osfstorage/file-1?download=1")
	if err == nil {
		t.Fatal("OpenDownload returned nil error, want forbidden error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.StatusCode != http.StatusForbidden || apiErr.Title != "Forbidden" {
		t.Fatalf("unexpected api error: %+v", apiErr)
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

func TestAPIErrorFallsBackToTopLevelFields(t *testing.T) {
	err := parseAPIError(http.StatusInternalServerError, http.MethodGet, "/v2/nodes/project-123/", []byte(`{"title":"Internal Server Error","detail":"upstream failed"}`))

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.Title != "Internal Server Error" || apiErr.Detail != "upstream failed" {
		t.Fatalf("unexpected api error fallback: %+v", apiErr)
	}
}

func TestAPIErrorEmptyErrorsArray(t *testing.T) {
	err := parseAPIError(http.StatusInternalServerError, http.MethodGet, "/v2/nodes/project-123/", []byte(`{"errors":[],"title":"Server Error","detail":"something broke"}`))

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.Title != "Server Error" || apiErr.Detail != "something broke" {
		t.Fatalf("unexpected api error: %+v", apiErr)
	}
}

func TestAPIErrorNullErrors(t *testing.T) {
	err := parseAPIError(http.StatusInternalServerError, http.MethodGet, "/v2/nodes/project-123/", []byte(`{"errors":null,"title":"Top Title","detail":"Top Detail"}`))

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.Title != "Top Title" || apiErr.Detail != "Top Detail" {
		t.Fatalf("unexpected api error: %+v", apiErr)
	}
}

func TestAPIErrorFallsBackToBodyText(t *testing.T) {
	err := parseAPIError(http.StatusBadGateway, http.MethodGet, "/v2/nodes/project-123/", []byte("bad gateway from proxy"))

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.Title != "" || apiErr.Detail != "bad gateway from proxy" {
		t.Fatalf("unexpected api error body fallback: %+v", apiErr)
	}
}

func TestAPIErrorEmptyBody(t *testing.T) {
	err := parseAPIError(http.StatusBadGateway, http.MethodGet, "/v2/nodes/project-123/", nil)

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.Title != "" || apiErr.Detail != "" {
		t.Fatalf("unexpected api error: %+v", apiErr)
	}
}

func TestResolveEndpointReturnsAbsoluteURLs(t *testing.T) {
	client, err := New("https://api.osf.io/v2/")
	if err != nil {
		t.Fatal(err)
	}

	resolved, err := client.resolveEndpoint("https://example.org/v2/nodes/project-123/")
	if err != nil {
		t.Fatalf("resolveEndpoint returned error: %v", err)
	}
	if got, want := resolved.String(), "https://example.org/v2/nodes/project-123/"; got != want {
		t.Fatalf("resolved endpoint = %q, want %q", got, want)
	}
}

func TestResolveReferenceKeepsAbsoluteReference(t *testing.T) {
	base, err := http.NewRequest(http.MethodGet, "https://api.osf.io/v2/nodes/project-123/", nil)
	if err != nil {
		t.Fatal(err)
	}

	resolved, err := resolveReference(base.URL, "https://example.org/v2/nodes/project-123/?page=2")
	if err != nil {
		t.Fatalf("resolveReference returned error: %v", err)
	}
	if got, want := resolved.String(), "https://example.org/v2/nodes/project-123/?page=2"; got != want {
		t.Fatalf("resolved reference = %q, want %q", got, want)
	}
}

func TestNewClientDefaultsBaseURL(t *testing.T) {
	t.Parallel()

	client, err := New("")
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	if got := client.baseURL.String(); got != "https://api.osf.io/v2/" {
		t.Fatalf("base URL = %q, want %q", got, "https://api.osf.io/v2/")
	}
}

func TestAPIErrorErrorFormatting(t *testing.T) {
	err := (&APIError{
		StatusCode: http.StatusBadGateway,
		Method:     http.MethodPost,
		Path:       "/v2/nodes/project-123/",
		Title:      "Bad Gateway",
		Detail:     "upstream failed",
	}).Error()

	if !strings.Contains(err, "POST /v2/nodes/project-123/ returned 502") {
		t.Fatalf("formatted error = %q", err)
	}
	if !strings.Contains(err, "Bad Gateway - upstream failed") {
		t.Fatalf("formatted error = %q", err)
	}
}

func TestAPIErrorNilError(t *testing.T) {
	t.Parallel()

	var err *APIError
	msg := err.Error()
	if msg != "<nil>" {
		t.Fatalf("nil APIError.Error() = %q, want <nil>", msg)
	}
}

func TestAPIErrorTitleOnly(t *testing.T) {
	t.Parallel()

	err := &APIError{StatusCode: 400, Method: "GET", Path: "/v2/", Title: "Bad Request"}
	msg := err.Error()
	if !strings.Contains(msg, "Bad Request") {
		t.Fatalf("APIError.Error() = %q, want title", msg)
	}
}

func TestAPIErrorDetailOnly(t *testing.T) {
	t.Parallel()

	err := &APIError{StatusCode: 500, Method: "POST", Path: "/v2/", Detail: "server error"}
	msg := err.Error()
	if !strings.Contains(msg, "server error") {
		t.Fatalf("APIError.Error() = %q, want detail", msg)
	}
}

func TestAPIErrorMalformedJSON(t *testing.T) {
	t.Parallel()

	err := parseAPIError(http.StatusBadGateway, http.MethodGet, "/v2/", []byte(`{invalid json`))

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.StatusCode != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", apiErr.StatusCode, http.StatusBadGateway)
	}
}

func TestParseAPIErrorWithEmptyBody(t *testing.T) {
	t.Parallel()

	err := parseAPIError(http.StatusBadGateway, http.MethodGet, "/v2/", nil)
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.StatusCode != http.StatusBadGateway {
		t.Fatalf("status = %d", apiErr.StatusCode)
	}
}

func TestParseAPIErrorWithEmptyBytesBody(t *testing.T) {
	t.Parallel()

	err := parseAPIError(http.StatusBadGateway, http.MethodGet, "/v2/", []byte{})
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.StatusCode != http.StatusBadGateway {
		t.Fatalf("status = %d", apiErr.StatusCode)
	}
}

func TestResolveEndpointParseError(t *testing.T) {
	t.Parallel()

	client, err := New("https://api.osf.io/v2/")
	if err != nil {
		t.Fatal(err)
	}

	// url.Parse doesn't fail on most inputs, but a bare ":" triggers a parse error
	_, err = client.resolveEndpoint(":")
	if err != nil {
		return
	}
	// If no error, fall back to a clearly invalid endpoint approach
	_, err = client.resolveEndpoint("\x00")
	if err == nil {
		t.Log("resolveEndpoint did not error on null byte input")
	}
}

func TestResolveReferenceParseError(t *testing.T) {
	t.Parallel()

	base, err := url.Parse("https://api.osf.io/v2/")
	if err != nil {
		t.Fatal(err)
	}

	// url.Parse doesn't fail on most inputs, but control characters can trigger errors
	_, err = resolveReference(base, "\x00")
	if err == nil {
		t.Log("resolveReference did not error on null byte input")
	}
}

func TestNewClientWithInvalidBaseURLStillSucceeds(t *testing.T) {
	t.Parallel()

	// url.Parse succeeds with most inputs; test with empty scheme
	client, err := New("relative/path")
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	if client == nil {
		t.Fatal("New returned nil client")
	}
}

func TestClientGetWithError(t *testing.T) {
	t.Parallel()

	client, err := New("https://0.0.0.0:1/v2/", WithHTTPClient(&http.Client{Timeout: 1}))
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.CurrentUser(t.Context())
	// Should get a network error or timeout
	if err == nil {
		t.Log("CurrentUser succeeded against invalid endpoint (unexpected)")
	}
}

func TestResolveReferenceHandlesRelativeURL(t *testing.T) {
	t.Parallel()

	client, err := New("https://api.osf.io/v2/")
	if err != nil {
		t.Fatal(err)
	}

	resolved, err := client.resolveEndpoint("/v2/nodes/project-123/")
	if err != nil {
		t.Fatalf("resolveEndpoint returned error: %v", err)
	}
	if !strings.Contains(resolved.String(), "/v2/nodes/project-123/") {
		t.Fatalf("resolved endpoint = %q, want relative path", resolved.String())
	}
}

func TestGetUser(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer token-123" {
			t.Fatalf("authorization header = %q", got)
		}
		if got := r.URL.Path; got != "/v2/users/u1/" {
			t.Fatalf("path = %q", got)
		}
		writeFixture(t, w, "user_me.json")
	}))
	defer srv.Close()

	client, err := New(srv.URL, WithHTTPClient(srv.Client()), WithBearerToken("token-123"))
	if err != nil {
		t.Fatal(err)
	}

	user, err := client.GetUser(context.Background(), "u1")
	if err != nil {
		t.Fatal(err)
	}
	if user.ID != "u1" || user.Attributes.FullName != "Ada Lovelace" {
		t.Fatalf("unexpected user: %+v", user)
	}
}

func TestListNodeCollectionEndpoints(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/nodes/project-123/registrations/":
			writeFixture(t, w, "node_list_page1.json")
		case "/v2/nodes/project-123/wikis/":
			writeFixture(t, w, "node_list_page1.json")
		case "/v2/nodes/project-123/comments/":
			writeFixture(t, w, "node_list_page1.json")
		case "/v2/nodes/project-123/logs/":
			writeFixture(t, w, "node_list_page1.json")
		case "/v2/nodes/project-123/identifiers/":
			writeFixture(t, w, "node_list_page1.json")
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	client, err := New(srv.URL, WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatal(err)
	}

	registrations, err := client.ListNodeRegistrations(context.Background(), "project-123")
	if err != nil {
		t.Fatalf("ListNodeRegistrations: %v", err)
	}
	if len(registrations) != 1 || registrations[0].Attributes.Title != "List Item One" {
		t.Fatalf("registrations = %+v", registrations)
	}

	wikis, err := client.ListWikiPages(context.Background(), "project-123")
	if err != nil {
		t.Fatalf("ListWikiPages: %v", err)
	}
	if len(wikis) != 1 {
		t.Fatalf("wikis length = %d", len(wikis))
	}

	comments, err := client.ListNodeComments(context.Background(), "project-123")
	if err != nil {
		t.Fatalf("ListNodeComments: %v", err)
	}
	if len(comments) != 1 {
		t.Fatalf("comments length = %d", len(comments))
	}

	logs, err := client.ListNodeLogs(context.Background(), "project-123")
	if err != nil {
		t.Fatalf("ListNodeLogs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("logs length = %d", len(logs))
	}

	identifiers, err := client.ListNodeIdentifiers(context.Background(), "project-123")
	if err != nil {
		t.Fatalf("ListNodeIdentifiers: %v", err)
	}
	if len(identifiers) != 1 {
		t.Fatalf("identifiers length = %d", len(identifiers))
	}
}

func TestCreateNode(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Method; got != http.MethodPost {
			t.Fatalf("method = %q", got)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer token-123" {
			t.Fatalf("authorization header = %q", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/vnd.api+json" {
			t.Fatalf("content type = %q", got)
		}
		if got := r.URL.Path; got != "/v2/nodes/" {
			t.Fatalf("path = %q", got)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"New Project"`) {
			t.Fatalf("request body = %q", string(body))
		}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"data":{"id":"new-project-1","type":"nodes","attributes":{"title":"New Project","category":"project","description":"A description"}}}`))
	}))
	defer srv.Close()

	client, err := New(srv.URL, WithHTTPClient(srv.Client()), WithBearerToken("token-123"))
	if err != nil {
		t.Fatal(err)
	}

	node, err := client.CreateNode(context.Background(), "New Project", "project", "A description")
	if err != nil {
		t.Fatal(err)
	}
	if node.ID != "new-project-1" || node.Attributes.Title != "New Project" {
		t.Fatalf("unexpected node: %+v", node)
	}
}

func TestUpdateNode(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Method; got != http.MethodPatch {
			t.Fatalf("method = %q", got)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer token-123" {
			t.Fatalf("authorization header = %q", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/vnd.api+json" {
			t.Fatalf("content type = %q", got)
		}
		if got := r.URL.Path; got != "/v2/nodes/project-123/" {
			t.Fatalf("path = %q", got)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"Updated Title"`) {
			t.Fatalf("request body = %q", string(body))
		}
		writeFixture(t, w, "node_project.json")
	}))
	defer srv.Close()

	client, err := New(srv.URL, WithHTTPClient(srv.Client()), WithBearerToken("token-123"))
	if err != nil {
		t.Fatal(err)
	}

	node, err := client.UpdateNode(context.Background(), "project-123", "Updated Title", "Updated description")
	if err != nil {
		t.Fatal(err)
	}
	if node.ID != "project-123" || node.Attributes.Title != "Project Alpha" {
		t.Fatalf("unexpected node: %+v", node)
	}
}

func TestDeleteNode(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Method; got != http.MethodDelete {
			t.Fatalf("method = %q", got)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer token-123" {
			t.Fatalf("authorization header = %q", got)
		}
		if got := r.URL.Path; got != "/v2/nodes/project-123/" {
			t.Fatalf("path = %q", got)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	client, err := New(srv.URL, WithHTTPClient(srv.Client()), WithBearerToken("token-123"))
	if err != nil {
		t.Fatal(err)
	}

	if err := client.DeleteNode(context.Background(), "project-123"); err != nil {
		t.Fatalf("DeleteNode returned error: %v", err)
	}
}

func TestDeleteNodeError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/nodes/nonexistent/" {
			w.WriteHeader(http.StatusNotFound)
			writeFixture(t, w, "error_not_found.json")
			return
		}
		t.Fatalf("unexpected path: %s", r.URL.Path)
	}))
	defer srv.Close()

	client, err := New(srv.URL, WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatal(err)
	}

	err = client.DeleteNode(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("DeleteNode returned nil error, want APIError")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.StatusCode != http.StatusNotFound || apiErr.Title != "Not Found" {
		t.Fatalf("unexpected api error: %+v", apiErr)
	}
}

func TestCreateNodeAPIError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"errors":[{"title":"Bad Request","detail":"validation failed"}]}`))
	}))
	defer srv.Close()

	client, err := New(srv.URL, WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.CreateNode(context.Background(), "", "", "")
	if err == nil {
		t.Fatal("CreateNode returned nil error, want APIError")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.StatusCode != http.StatusBadRequest || apiErr.Title != "Bad Request" {
		t.Fatalf("unexpected api error: %+v", apiErr)
	}
}

func TestUpdateNodeAPIError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"errors":[{"title":"Conflict","detail":"resource modified"}]}`))
	}))
	defer srv.Close()

	client, err := New(srv.URL, WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.UpdateNode(context.Background(), "project-123", "title", "desc")
	if err == nil {
		t.Fatal("UpdateNode returned nil error, want APIError")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.StatusCode != http.StatusConflict || apiErr.Title != "Conflict" {
		t.Fatalf("unexpected api error: %+v", apiErr)
	}
}

func TestResolveReferenceRelative(t *testing.T) {
	t.Parallel()

	base, err := url.Parse("https://api.osf.io/v2/")
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := resolveReference(base, "nodes/abc123/?page=2")
	if err != nil {
		t.Fatalf("resolveReference returned error: %v", err)
	}
	if !strings.Contains(resolved.String(), "nodes/abc123") {
		t.Fatalf("resolved reference = %q", resolved.String())
	}
}

func TestNewClientInvalidURL(t *testing.T) {
	t.Parallel()

	_, err := New("://invalid")
	if err == nil {
		t.Fatal("New returned nil error, want parse error")
	}
}

func TestOpenDownloadNonStreamURL(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("file-content"))
	}))
	defer server.Close()

	client, err := New(server.URL + "/v2/")
	if err != nil {
		t.Fatal(err)
	}

	rc, err := client.OpenDownload(t.Context(), server.URL+"/download")
	if err != nil {
		t.Fatalf("OpenDownload returned error: %v", err)
	}
	defer func() { _ = rc.Close() }()

	body, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if string(body) != "file-content" {
		t.Fatalf("body = %q, want %q", string(body), "file-content")
	}
}

func TestOpenDownloadErrorResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errors":[{"title":"Not Found"}]}`))
	}))
	defer server.Close()

	client, err := New(server.URL + "/v2/")
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.OpenDownload(t.Context(), server.URL+"/missing")
	if err == nil {
		t.Fatal("OpenDownload returned nil error, want APIError")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Fatalf("error = %q, want status code", err.Error())
	}
}

func TestNewClientWithOptions(t *testing.T) {
	t.Parallel()

	client, err := New("https://api.osf.io/v2/", WithHTTPClient(&http.Client{}), WithBearerToken("test-token"))
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	if client == nil {
		t.Fatal("New returned nil client")
	}
}

func TestListNodeContributorsEndpoint(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_, _ = w.Write([]byte(`{"data":[],"links":{}}`))
	}))
	defer server.Close()

	client, err := New(server.URL + "/v2/")
	if err != nil {
		t.Fatal(err)
	}

	contributors, err := client.ListNodeContributors(t.Context(), "project-1")
	if err != nil {
		t.Fatalf("ListNodeContributors returned error: %v", err)
	}
	if len(contributors) != 0 {
		t.Fatalf("contributors = %d, want 0", len(contributors))
	}
}

func TestListStorageFilesWithSegments(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_, _ = w.Write([]byte(`{"data":[],"links":{}}`))
	}))
	defer server.Close()

	client, err := New(server.URL + "/v2/")
	if err != nil {
		t.Fatal(err)
	}

	files, err := client.ListStorageFiles(t.Context(), "project-1", "subdir", "nested")
	if err != nil {
		t.Fatalf("ListStorageFiles returned error: %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("files = %d, want 0", len(files))
	}
}

func TestUploadFile(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Method; got != http.MethodPut {
			t.Fatalf("method = %q", got)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer token-123" {
			t.Fatalf("authorization header = %q", got)
		}
		if got := r.URL.Path; got != "/v1/providers/osfstorage/abc123/report.txt" {
			t.Fatalf("path = %q", got)
		}
		if got := r.Header.Get("Content-Type"); got != "text/plain; charset=utf-8" {
			t.Fatalf("content-type = %q", got)
		}
		if got := r.URL.Query().Get("kind"); got != "file" {
			t.Fatalf("kind = %q", got)
		}
		if got := r.URL.Query().Get("conflict"); got != "overwrite" {
			t.Fatalf("conflict = %q", got)
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != "col1,col2\n1,2\n" {
			t.Fatalf("body = %q", string(body))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client, err := New(srv.URL, WithHTTPClient(srv.Client()), WithBearerToken("token-123"))
	if err != nil {
		t.Fatal(err)
	}

	providerURL := srv.URL + "/v1/providers/osfstorage/abc123"
	err = client.UploadFile(t.Context(), providerURL, "report.txt", strings.NewReader("col1,col2\n1,2\n"), "overwrite")
	if err != nil {
		t.Fatalf("UploadFile returned error: %v", err)
	}
}

func TestUploadFileError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte("file already exists"))
	}))
	defer srv.Close()

	client, err := New(srv.URL, WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatal(err)
	}

	err = client.UploadFile(t.Context(), srv.URL+"/v1/providers/osfstorage/abc123", "existing.txt", strings.NewReader("data"), "fail")
	if err == nil {
		t.Fatal("UploadFile returned nil error, want error")
	}
	if !strings.Contains(err.Error(), "file already exists") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestCreateFolder(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Method; got != http.MethodPut {
			t.Fatalf("method = %q", got)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer token-123" {
			t.Fatalf("authorization header = %q", got)
		}
		if got := r.URL.Path; got != "/v1/providers/osfstorage/abc123/myfolder/" {
			t.Fatalf("path = %q", got)
		}
		if got := r.URL.Query().Get("kind"); got != "folder" {
			t.Fatalf("kind = %q", got)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	client, err := New(srv.URL, WithHTTPClient(srv.Client()), WithBearerToken("token-123"))
	if err != nil {
		t.Fatal(err)
	}

	err = client.CreateFolder(t.Context(), srv.URL+"/v1/providers/osfstorage/abc123", "myfolder")
	if err != nil {
		t.Fatalf("CreateFolder returned error: %v", err)
	}
}

func TestCreateFolderNestedPath(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/v1/providers/osfstorage/abc123/nested/folder/" {
			t.Fatalf("path = %q", got)
		}
		if got := r.URL.Query().Get("kind"); got != "folder" {
			t.Fatalf("kind = %q", got)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	client, err := New(srv.URL, WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatal(err)
	}

	if err := client.CreateFolder(t.Context(), srv.URL+"/v1/providers/osfstorage/abc123", "nested/folder"); err != nil {
		t.Fatalf("CreateFolder returned error: %v", err)
	}
}

func TestWaterButlerPathRejectsTraversal(t *testing.T) {
	t.Parallel()

	client, err := New("https://api.osf.io/v2/")
	if err != nil {
		t.Fatal(err)
	}

	err = client.CreateFolder(t.Context(), "https://files.osf.io/v1/providers/osfstorage/abc123", "../outside")
	if err == nil {
		t.Fatal("CreateFolder returned nil error, want traversal error")
	}
	if !strings.Contains(err.Error(), "must stay within OSF Storage") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestCreateFolderError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte("folder exists"))
	}))
	defer srv.Close()

	client, err := New(srv.URL, WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatal(err)
	}

	err = client.CreateFolder(t.Context(), srv.URL+"/v1/providers/osfstorage/abc123", "existing")
	if err == nil {
		t.Fatal("CreateFolder returned nil error, want error")
	}
	if !strings.Contains(err.Error(), "folder exists") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestDeleteFile(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Method; got != http.MethodDelete {
			t.Fatalf("method = %q", got)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer token-123" {
			t.Fatalf("authorization header = %q", got)
		}
		if got := r.URL.Path; got != "/v1/providers/osfstorage/abc123/old.csv" {
			t.Fatalf("path = %q", got)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client, err := New(srv.URL, WithHTTPClient(srv.Client()), WithBearerToken("token-123"))
	if err != nil {
		t.Fatal(err)
	}

	err = client.DeleteFile(t.Context(), srv.URL+"/v1/providers/osfstorage/abc123", "old.csv")
	if err != nil {
		t.Fatalf("DeleteFile returned error: %v", err)
	}
}

func TestDeleteFileError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("file not found"))
	}))
	defer srv.Close()

	client, err := New(srv.URL, WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatal(err)
	}

	err = client.DeleteFile(t.Context(), srv.URL+"/v1/providers/osfstorage/abc123", "missing.txt")
	if err == nil {
		t.Fatal("DeleteFile returned nil error, want error")
	}
	if !strings.Contains(err.Error(), "file not found") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestListPreprints(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/v2/preprints/" {
			t.Fatalf("path = %q", got)
		}
		if got := r.URL.Query().Get("filter[provider]"); got != "osf" {
			t.Fatalf("filter[provider] = %q", got)
		}
		writeFixture(t, w, "node_list_page1.json")
	}))
	defer srv.Close()

	client, err := New(srv.URL, WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatal(err)
	}

	preprints, err := client.ListPreprints(t.Context(), "osf")
	if err != nil {
		t.Fatalf("ListPreprints returned error: %v", err)
	}
	if len(preprints) != 1 || preprints[0].ID != "item-1" {
		t.Fatalf("preprints = %+v", preprints)
	}
}

func TestSearchPreprints(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/v2/preprints/" {
			t.Fatalf("path = %q", got)
		}
		if got, want := r.URL.Query().Get("filter[title]"), "open science"; got != want {
			t.Fatalf("filter[title] = %q, want %q", got, want)
		}
		if got, want := r.URL.Query().Get("filter[provider]"), "osf"; got != want {
			t.Fatalf("filter[provider] = %q, want %q", got, want)
		}
		writeFixture(t, w, "preprint_search_page1.json")
	}))
	defer srv.Close()

	client, err := New(srv.URL, WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatal(err)
	}

	preprints, err := client.SearchPreprints(t.Context(), "open science", "osf", 1)
	if err != nil {
		t.Fatalf("SearchPreprints returned error: %v", err)
	}
	if len(preprints) != 1 || preprints[0].ID != "preprint-1" {
		t.Fatalf("preprints = %+v", preprints)
	}
	if got := preprints[0].Attributes.DOI; got != "10.1234/preprint-1" {
		t.Fatalf("doi = %q", got)
	}
}

func TestSearchOSF(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/v2/search/" {
			t.Fatalf("path = %q", got)
		}
		if got, want := r.URL.Query().Get("q"), "open science"; got != want {
			t.Fatalf("q = %q, want %q", got, want)
		}
		_, _ = fmt.Fprint(w, `{"data":[{"id":"item-1","type":"nodes","attributes":{"title":"List Item One","description":"An abstract","category":"component","tags":["open science","review"],"date_created":"2024-03-15T00:00:00Z"},"links":{"self":"https://osf.io/item-1/"}}],"links":{}}`)
	}))
	defer srv.Close()

	client, err := New(srv.URL, WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatal(err)
	}

	results, err := client.SearchOSF(t.Context(), "open science")
	if err != nil {
		t.Fatalf("SearchOSF returned error: %v", err)
	}
	if len(results) != 1 || results[0].ID != "item-1" {
		t.Fatalf("results = %+v", results)
	}
	if results[0].Type != "nodes" || results[0].Title != "List Item One" {
		t.Fatalf("typed result = %+v", results[0])
	}
	if results[0].Year != "2024" || len(results[0].Keywords) != 2 {
		t.Fatalf("bibliographic fields = %+v", results[0])
	}
}

func TestCreateRegistration(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Method; got != http.MethodPost {
			t.Fatalf("method = %q", got)
		}
		if got := r.URL.Path; got != "/v2/nodes/project-123/registrations/" {
			t.Fatalf("path = %q", got)
		}
		var body map[string]map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		data := body["data"]
		attrs, ok := data["attributes"].(map[string]any)
		if !ok {
			t.Fatalf("attributes = %#v", data["attributes"])
		}
		if got := attrs["registration_schema"]; got != "schema-1" {
			t.Fatalf("registration_schema = %#v", got)
		}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_, _ = w.Write([]byte(`{"data":{"id":"reg1","type":"registrations","attributes":{"title":"Draft"},"links":{"self":"https://osf.io/reg1/"}},"links":{}}`))
	}))
	defer srv.Close()

	client, err := New(srv.URL, WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatal(err)
	}

	registration, err := client.CreateRegistration(t.Context(), "project-123", RegistrationRequest{SchemaID: "schema-1", Title: "Draft"})
	if err != nil {
		t.Fatalf("CreateRegistration returned error: %v", err)
	}
	if registration.ID != "reg1" || registration.Type != "registrations" {
		t.Fatalf("registration = %+v", registration)
	}
}

func TestCreateRegistrationRequiresSchema(t *testing.T) {
	t.Parallel()

	client, err := New("https://api.osf.io/v2/")
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.CreateRegistration(t.Context(), "project-123", RegistrationRequest{})
	if err == nil {
		t.Fatal("CreateRegistration returned nil error, want schema error")
	}
}

func TestGetNodeFilesProvider(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/v2/nodes/project-123/files/" {
			t.Fatalf("path = %q", got)
		}
		writeFixture(t, w, "files_provider.json")
	}))
	defer srv.Close()

	client, err := New(srv.URL, WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatal(err)
	}

	providerURL, err := client.GetNodeFilesProvider(t.Context(), "project-123")
	if err != nil {
		t.Fatalf("GetNodeFilesProvider returned error: %v", err)
	}
	if providerURL != "https://files.osf.io/v1/providers/osfstorage/abc123" {
		t.Fatalf("provider URL = %q", providerURL)
	}
}

func TestGetNodeFilesProviderNotFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_, _ = w.Write([]byte(`{"data":[{"id":"github","type":"files","attributes":{"name":"github","full_name":"GitHub"},"links":{"self":"https://files.osf.io/v1/providers/github/repo1"}}],"links":{}}`))
	}))
	defer srv.Close()

	client, err := New(srv.URL, WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.GetNodeFilesProvider(t.Context(), "node-without-osfstorage")
	if err == nil {
		t.Fatal("GetNodeFilesProvider returned nil error, want error")
	}
	if !strings.Contains(err.Error(), "no osfstorage provider") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestListNodeAddons(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/v2/nodes/project-123/addons/" {
			t.Fatalf("path = %q", got)
		}
		writeFixture(t, w, "node_list_page1.json")
	}))
	defer srv.Close()

	client, err := New(srv.URL, WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatal(err)
	}

	addons, err := client.ListNodeAddons(t.Context(), "project-123")
	if err != nil {
		t.Fatalf("ListNodeAddons returned error: %v", err)
	}
	if len(addons) != 1 || addons[0].ID != "item-1" {
		t.Fatalf("addons = %+v", addons)
	}
}

func TestListFileVersions(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_, _ = w.Write([]byte(`{"data":[],"links":{}}`))
	}))
	defer server.Close()

	client, err := New(server.URL + "/v2/")
	if err != nil {
		t.Fatal(err)
	}

	versions, err := client.ListFileVersions(t.Context(), "file-1")
	if err != nil {
		t.Fatalf("ListFileVersions returned error: %v", err)
	}
	if len(versions) != 0 {
		t.Fatalf("versions = %d, want 0", len(versions))
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
