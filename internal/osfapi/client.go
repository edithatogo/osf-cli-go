package osfapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/edithatogo/osf-cli-go/internal/auth"
)

const defaultBaseURL = "https://api.osf.io/v2/"
const defaultHTTPTimeout = 30 * time.Second

// Client is an HTTP client for the OSF API v2 JSON:API surface.
// It handles authentication, pagination, and JSON deserialization.
type Client struct {
	baseURL     *url.URL
	httpClient  *http.Client
	bearerToken string
	credentials auth.Credentials
}

// Option configures a Client using the functional options pattern.
type Option func(*Client)

// WithHTTPClient sets the HTTP client used for API requests.
// If nil, a client with the default OSF timeout is used.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithBearerToken sets the Authorization header for API requests.
func WithBearerToken(token string) Option {
	return func(c *Client) {
		c.credentials = auth.Credentials{Mode: auth.ModeBearerToken, Token: token}
		c.bearerToken = token
	}
}

// WithCredentials sets the OSF credential mode for API requests.
func WithCredentials(credentials auth.Credentials) Option {
	return func(c *Client) {
		c.credentials = credentials
		if credentials.Mode == auth.ModeBearerToken {
			c.bearerToken = credentials.Token
		}
	}
}

// WithUsernamePassword sets HTTP Basic request signing credentials.
func WithUsernamePassword(username, password string) Option {
	return func(c *Client) {
		c.credentials = auth.Credentials{Mode: auth.ModeUsernamePassword, Username: username, Password: password}
	}
}

// New creates a Client that communicates with the given OSF API base URL.
// If baseURL is empty, the default production URL (https://api.osf.io/v2/) is used.
// Returns an error if the URL cannot be parsed.
func New(baseURL string, opts ...Option) (*Client, error) {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base url: %w", err)
	}

	c := &Client{
		baseURL:    parsed,
		httpClient: &http.Client{Timeout: defaultHTTPTimeout},
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.httpClient == nil {
		c.httpClient = &http.Client{Timeout: defaultHTTPTimeout}
	}
	return c, nil
}

// CurrentUser returns the authenticated user's profile information.
// Requires a valid bearer token; returns MissingTokenError otherwise.
func (c *Client) CurrentUser(ctx context.Context) (User, error) {
	var doc document[User]
	if _, err := c.get(ctx, "/v2/users/me/", &doc); err != nil {
		return User{}, err
	}
	return doc.Data, nil
}

// ListCurrentUserProjects returns all project-category nodes owned by the current user.
// Automatically follows pagination links. Requires a bearer token.
func (c *Client) ListCurrentUserProjects(ctx context.Context) ([]Node, error) {
	return collectPages[Node](ctx, c, "/v2/users/me/nodes/?filter[category]=project")
}

// GetNode returns a single node (project or component) by its OSF GUID.
func (c *Client) GetNode(ctx context.Context, id string) (Node, error) {
	var doc document[Node]
	if _, err := c.get(ctx, "/v2/nodes/"+url.PathEscape(id)+"/", &doc); err != nil {
		return Node{}, err
	}
	return doc.Data, nil
}

// GetStorageFile returns a single storage file or folder metadata record by its OSF GUID.
func (c *Client) GetStorageFile(ctx context.Context, id string) (StorageFile, error) {
	var doc document[StorageFile]
	if _, err := c.get(ctx, "/v2/files/"+url.PathEscape(id)+"/", &doc); err != nil {
		return StorageFile{}, err
	}
	return doc.Data, nil
}

// ResolveDOI resolves an OSF DOI without using the OSF API credentials.
func (c *Client) ResolveDOI(ctx context.Context, identifier string) (DOIResolution, error) {
	return ResolveDOI(ctx, identifier)
}

// ListFileVersions loads all versions for a file.
func (c *Client) ListFileVersions(ctx context.Context, fileID string) ([]FileVersion, error) {
	return collectPages[FileVersion](ctx, c, "/v2/files/"+url.PathEscape(fileID)+"/versions/")
}

// ListNodeChildren returns all immediate child components of the specified node.
// Automatically follows pagination links.
func (c *Client) ListNodeChildren(ctx context.Context, id string) ([]Node, error) {
	return collectPages[Node](ctx, c, "/v2/nodes/"+url.PathEscape(id)+"/children/")
}

// ListNodeContributors returns all contributors for the specified node.
// Automatically follows pagination links.
func (c *Client) ListNodeContributors(ctx context.Context, id string) ([]Contributor, error) {
	return collectPages[Contributor](ctx, c, "/v2/nodes/"+url.PathEscape(id)+"/contributors/")
}

// ListStorageFiles returns all files and folders from a node's OSF Storage provider.
// Additional path segments can be provided to navigate into subfolders.
// Automatically follows pagination links.
func (c *Client) ListStorageFiles(ctx context.Context, nodeID string, segments ...string) ([]StorageFile, error) {
	escaped := []string{"/v2/nodes", url.PathEscape(nodeID), "files", "osfstorage"}
	for _, segment := range segments {
		if segment == "" {
			continue
		}
		escaped = append(escaped, url.PathEscape(segment))
	}
	endpoint := path.Join(escaped...) + "/"
	return collectPages[StorageFile](ctx, c, endpoint)
}

func collectPages[T any](ctx context.Context, c *Client, endpoint string) ([]T, error) {
	return collectPagesLimit[T](ctx, c, endpoint, 0)
}

func collectPagesLimit[T any](ctx context.Context, c *Client, endpoint string, limit int) ([]T, error) {
	var all []T
	next := endpoint
	for next != "" {
		var page document[[]T]
		respURL, err := c.get(ctx, next, &page)
		if err != nil {
			return nil, err
		}
		for _, item := range page.Data {
			if limit > 0 && len(all) >= limit {
				return all, nil
			}
			all = append(all, item)
		}
		if limit > 0 && len(all) >= limit {
			break
		}
		if page.Links.Next == "" {
			break
		}
		resolved, err := resolveReference(respURL, page.Links.Next)
		if err != nil {
			return nil, err
		}
		next = resolved.String()
	}
	return all, nil
}

func (c *Client) get(ctx context.Context, endpoint string, dst any) (*url.URL, error) {
	reqURL, err := c.resolveEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return nil, err
	}
	c.sign(req)
	req.Header.Set("Accept", "application/vnd.api+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, parseAPIError(resp.StatusCode, req.Method, req.URL.RequestURI(), body)
	}

	if err := json.NewDecoder(bytes.NewReader(body)).Decode(dst); err != nil {
		return nil, fmt.Errorf("decode osf response from %s: %w", req.URL.RequestURI(), err)
	}
	return req.URL, nil
}

func (c *Client) post(ctx context.Context, endpoint string, body any, dst any) (*url.URL, error) {
	reqURL, err := c.resolveEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	var reqBody io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("encode request body: %w", err)
		}
		reqBody = bytes.NewReader(encoded)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL.String(), reqBody)
	if err != nil {
		return nil, err
	}
	c.sign(req)
	req.Header.Set("Accept", "application/vnd.api+json")
	req.Header.Set("Content-Type", "application/vnd.api+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, parseAPIError(resp.StatusCode, req.Method, req.URL.RequestURI(), respBody)
	}

	if dst != nil {
		if err := json.NewDecoder(bytes.NewReader(respBody)).Decode(dst); err != nil {
			return nil, fmt.Errorf("decode osf response from %s: %w", req.URL.RequestURI(), err)
		}
	}
	return req.URL, nil
}

func (c *Client) patch(ctx context.Context, endpoint string, body any, dst any) (*url.URL, error) {
	reqURL, err := c.resolveEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	var reqBody io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("encode request body: %w", err)
		}
		reqBody = bytes.NewReader(encoded)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, reqURL.String(), reqBody)
	if err != nil {
		return nil, err
	}
	c.sign(req)
	req.Header.Set("Accept", "application/vnd.api+json")
	req.Header.Set("Content-Type", "application/vnd.api+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, parseAPIError(resp.StatusCode, req.Method, req.URL.RequestURI(), respBody)
	}

	if dst != nil {
		if err := json.NewDecoder(bytes.NewReader(respBody)).Decode(dst); err != nil {
			return nil, fmt.Errorf("decode osf response from %s: %w", req.URL.RequestURI(), err)
		}
	}
	return req.URL, nil
}

func (c *Client) delete(ctx context.Context, endpoint string) error {
	reqURL, err := c.resolveEndpoint(endpoint)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, reqURL.String(), nil)
	if err != nil {
		return err
	}
	c.sign(req)
	req.Header.Set("Accept", "application/vnd.api+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return parseAPIError(resp.StatusCode, req.Method, req.URL.RequestURI(), respBody)
	}
	return nil
}

// OpenDownload opens the given download URL and returns the response body.
// The caller is responsible for closing the returned io.ReadCloser.
func (c *Client) OpenDownload(ctx context.Context, downloadURL string) (io.ReadCloser, error) {
	reqURL, err := c.resolveEndpoint(downloadURL)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return nil, err
	}
	c.sign(req)
	req.Header.Set("Accept", "*/*")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("read download error body: %w", readErr)
		}
		return nil, parseAPIError(resp.StatusCode, req.Method, req.URL.RequestURI(), body)
	}

	return resp.Body, nil
}

// UploadFile uploads content to a storage provider via WaterButler.
// The providerURL is typically obtained from GET /v2/nodes/{id}/files/osfstorage/
// and looks like "https://files.osf.io/v1/providers/osfstorage/..."
func (c *Client) UploadFile(ctx context.Context, providerURL, fileName string, content io.Reader, conflict string) error {
	fullURL, err := waterButlerPath(providerURL, fileName, false)
	if err != nil {
		return err
	}
	if conflict != "" {
		parsed, _ := url.Parse(fullURL)
		q := parsed.Query()
		q.Set("kind", "file")
		if conflict == "overwrite" {
			q.Set("conflict", "overwrite")
		} else {
			q.Set("conflict", "fail")
		}
		parsed.RawQuery = q.Encode()
		fullURL = parsed.String()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, fullURL, content)
	if err != nil {
		return err
	}
	if contentType := mime.TypeByExtension(filepath.Ext(fileName)); contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	c.sign(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed: %s", strings.TrimSpace(string(body)))
	}
	return nil
}

// CreateFolder creates a folder via WaterButler.
func (c *Client) CreateFolder(ctx context.Context, providerURL, folderName string) error {
	fullURL, err := waterButlerPath(providerURL, folderName, true)
	if err != nil {
		return err
	}
	parsed, _ := url.Parse(fullURL)
	q := parsed.Query()
	q.Set("kind", "folder")
	parsed.RawQuery = q.Encode()
	fullURL = parsed.String()

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, fullURL, http.NoBody)
	if err != nil {
		return err
	}
	c.sign(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("create folder failed: %s", strings.TrimSpace(string(body)))
	}
	return nil
}

// DeleteFile deletes a file via WaterButler.
func (c *Client) DeleteFile(ctx context.Context, providerURL, fileName string) error {
	fullURL, err := waterButlerPath(providerURL, fileName, false)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, fullURL, nil)
	if err != nil {
		return err
	}
	c.sign(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete file failed: %s", strings.TrimSpace(string(body)))
	}
	return nil
}

func waterButlerPath(providerURL, remotePath string, trailingSlash bool) (string, error) {
	trimmed := strings.TrimSpace(remotePath)
	if trimmed == "" {
		return "", fmt.Errorf("remote path is required")
	}

	clean := path.Clean(strings.ReplaceAll(trimmed, "\\", "/"))
	if clean == "." || clean == "/" || strings.HasPrefix(clean, "../") || strings.Contains(clean, "/../") {
		return "", fmt.Errorf("remote path %q must stay within OSF Storage", remotePath)
	}

	parts := strings.Split(strings.Trim(clean, "/"), "/")
	escaped := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			return "", fmt.Errorf("remote path %q must stay within OSF Storage", remotePath)
		}
		escaped = append(escaped, url.PathEscape(part))
	}

	fullURL := strings.TrimRight(providerURL, "/") + "/" + strings.Join(escaped, "/")
	if trailingSlash {
		fullURL += "/"
	}
	return fullURL, nil
}

func (c *Client) sign(req *http.Request) {
	switch c.credentials.Mode {
	case auth.ModeBearerToken:
		token := strings.TrimSpace(c.credentials.Token)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	case auth.ModeUsernamePassword:
		username := strings.TrimSpace(c.credentials.Username)
		password := strings.TrimSpace(c.credentials.Password)
		if username != "" && password != "" {
			req.SetBasicAuth(username, password)
		}
	default:
		if c.bearerToken != "" {
			req.Header.Set("Authorization", "Bearer "+c.bearerToken)
		}
	}
}

// ListPreprints loads preprints visible to the current request context.
func (c *Client) ListPreprints(ctx context.Context, provider string, limit ...int) ([]Node, error) {
	endpoint := "/v2/preprints/"
	if provider != "" {
		endpoint = appendQuery(endpoint, "filter[provider]", provider)
	}
	return collectPagesLimit[Node](ctx, c, endpoint, optionalLimit(limit))
}

// SearchPreprints searches OSF preprints by title and optionally provider.
func (c *Client) SearchPreprints(ctx context.Context, query, provider string, limit ...int) ([]Preprint, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("preprint search query is required")
	}
	endpoint := "/v2/preprints/"
	endpoint = appendQuery(endpoint, "filter[title]", query)
	if provider != "" {
		endpoint = appendQuery(endpoint, "filter[provider]", provider)
	}
	return collectPagesLimit[Preprint](ctx, c, endpoint, optionalLimit(limit))
}

// SearchOSF performs a search across OSF content.
func (c *Client) SearchOSF(ctx context.Context, query string, limit ...int) ([]SearchResult, error) {
	endpoint := "/v2/search/?q=" + url.QueryEscape(query)
	nodes, err := collectPagesLimit[Node](ctx, c, endpoint, optionalLimit(limit))
	if err != nil {
		return nil, err
	}
	results := make([]SearchResult, 0, len(nodes))
	for _, n := range nodes {
		results = append(results, SearchResult{
			ID:          n.ID,
			Type:        n.Type,
			Title:       n.Attributes.Title,
			Description: n.Attributes.Description,
			Category:    n.Attributes.Category,
			URL:         n.Links.Self,
			Keywords:    append([]string(nil), n.Attributes.Tags...),
			Year:        yearFromDate(n.Attributes.DateCreated),
		})
	}
	return results, nil
}

func yearFromDate(value string) string {
	if len(value) >= 4 {
		return value[:4]
	}
	return value
}

func optionalLimit(values []int) int {
	if len(values) == 0 || values[0] < 0 {
		return 0
	}
	return values[0]
}

func appendQuery(endpoint, key, value string) string {
	separator := "?"
	if strings.Contains(endpoint, "?") {
		separator = "&"
	}
	return endpoint + separator + url.QueryEscape(key) + "=" + url.QueryEscape(value)
}

// ListNodeAddons lists all storage add-ons configured for a node.
func (c *Client) ListNodeAddons(ctx context.Context, id string) ([]Node, error) {
	return collectPages[Node](ctx, c, "/v2/nodes/"+url.PathEscape(id)+"/addons/")
}

// CreateRegistration creates a draft registration for a node.
func (c *Client) CreateRegistration(ctx context.Context, nodeID string, request RegistrationRequest) (Node, error) {
	if strings.TrimSpace(request.SchemaID) == "" {
		return Node{}, fmt.Errorf("registration schema id is required")
	}

	attributes := map[string]string{
		"registration_schema": request.SchemaID,
	}
	if request.Title != "" {
		attributes["title"] = request.Title
	}
	if request.Description != "" {
		attributes["description"] = request.Description
	}

	var doc document[Node]
	_, err := c.post(ctx, "/v2/nodes/"+url.PathEscape(nodeID)+"/registrations/", map[string]any{
		"data": map[string]any{
			"type":       "registrations",
			"attributes": attributes,
		},
	}, &doc)
	if err != nil {
		return Node{}, err
	}
	return doc.Data, nil
}

// GetNodeFilesProvider gets the files provider URL for a node's OSF Storage.
// Returns the provider URL for use with WaterButler operations.
func (c *Client) GetNodeFilesProvider(ctx context.Context, nodeID string) (string, error) {
	var doc document[[]struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			Name     string `json:"name"`
			FullName string `json:"full_name"`
		} `json:"attributes"`
		Links Links `json:"links"`
	}]
	if _, err := c.get(ctx, "/v2/nodes/"+url.PathEscape(nodeID)+"/files/", &doc); err != nil {
		return "", err
	}
	for _, provider := range doc.Data {
		if provider.ID == "osfstorage" {
			return provider.Links.Self, nil
		}
	}
	return "", fmt.Errorf("no osfstorage provider found for node %q", nodeID)
}

// GetUser returns a user profile by OSF user ID.
func (c *Client) GetUser(ctx context.Context, id string) (User, error) {
	var doc document[User]
	if _, err := c.get(ctx, "/v2/users/"+url.PathEscape(id)+"/", &doc); err != nil {
		return User{}, err
	}
	return doc.Data, nil
}

// ListNodeRegistrations loads registrations linked to a node.
func (c *Client) ListNodeRegistrations(ctx context.Context, id string) ([]Node, error) {
	return collectPages[Node](ctx, c, "/v2/nodes/"+url.PathEscape(id)+"/registrations/")
}

// ListWikiPages loads wiki pages linked to a node.
func (c *Client) ListWikiPages(ctx context.Context, id string) ([]RelatedResource, error) {
	return collectPages[RelatedResource](ctx, c, "/v2/nodes/"+url.PathEscape(id)+"/wikis/")
}

// ListNodeComments loads comments linked to a node.
func (c *Client) ListNodeComments(ctx context.Context, id string) ([]RelatedResource, error) {
	return collectPages[RelatedResource](ctx, c, "/v2/nodes/"+url.PathEscape(id)+"/comments/")
}

// ListNodeLogs loads audit log entries linked to a node.
func (c *Client) ListNodeLogs(ctx context.Context, id string) ([]RelatedResource, error) {
	return collectPages[RelatedResource](ctx, c, "/v2/nodes/"+url.PathEscape(id)+"/logs/")
}

// ListNodeIdentifiers loads identifiers linked to a node.
func (c *Client) ListNodeIdentifiers(ctx context.Context, id string) ([]RelatedResource, error) {
	return collectPages[RelatedResource](ctx, c, "/v2/nodes/"+url.PathEscape(id)+"/identifiers/")
}

// CreateNode creates an OSF node with the supplied title, category, and description.
func (c *Client) CreateNode(ctx context.Context, title, category, description string) (Node, error) {
	var doc document[Node]
	_, err := c.post(ctx, "/v2/nodes/", map[string]any{
		"data": map[string]any{
			"type": "nodes",
			"attributes": map[string]string{
				"title":       title,
				"category":    category,
				"description": description,
			},
		},
	}, &doc)
	if err != nil {
		return Node{}, err
	}
	return doc.Data, nil
}

// UpdateNode updates title and description metadata for an OSF node.
func (c *Client) UpdateNode(ctx context.Context, id, title, description string) (Node, error) {
	var doc document[Node]
	body := map[string]any{
		"data": map[string]any{
			"id":   id,
			"type": "nodes",
			"attributes": map[string]string{
				"title":       title,
				"description": description,
			},
		},
	}
	_, err := c.patch(ctx, "/v2/nodes/"+url.PathEscape(id)+"/", body, &doc)
	if err != nil {
		return Node{}, err
	}
	return doc.Data, nil
}

// DeleteNode deletes an OSF node by ID.
func (c *Client) DeleteNode(ctx context.Context, id string) error {
	return c.delete(ctx, "/v2/nodes/"+url.PathEscape(id)+"/")
}

func (c *Client) resolveEndpoint(endpoint string) (*url.URL, error) {
	ref, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	if ref.IsAbs() {
		return ref, nil
	}
	return resolveReference(c.baseURL, endpoint)
}

func resolveReference(base *url.URL, ref string) (*url.URL, error) {
	parsed, err := url.Parse(ref)
	if err != nil {
		return nil, err
	}
	if parsed.IsAbs() {
		return parsed, nil
	}
	return base.ResolveReference(parsed), nil
}

func parseAPIError(statusCode int, method string, path string, body []byte) error {
	var payload struct {
		Errors []struct {
			Status string `json:"status"`
			Title  string `json:"title"`
			Detail string `json:"detail"`
		} `json:"errors"`
		Title  string `json:"title"`
		Detail string `json:"detail"`
	}
	_ = json.Unmarshal(body, &payload)

	apiErr := APIError{
		StatusCode: statusCode,
		Method:     method,
		Path:       path,
	}
	if len(payload.Errors) > 0 {
		apiErr.Title = payload.Errors[0].Title
		apiErr.Detail = payload.Errors[0].Detail
	}
	if apiErr.Title == "" {
		apiErr.Title = payload.Title
	}
	if apiErr.Detail == "" {
		apiErr.Detail = payload.Detail
	}
	if apiErr.Title == "" && len(bytes.TrimSpace(body)) > 0 {
		apiErr.Detail = strings.TrimSpace(string(body))
	}
	return &apiErr
}
