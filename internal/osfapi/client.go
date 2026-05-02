package osfapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
)

const defaultBaseURL = "https://api.osf.io/v2/"

// Client talks to the OSF API v2 JSON:API surface.
type Client struct {
	baseURL     *url.URL
	httpClient  *http.Client
	bearerToken string
}

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient injects the HTTP client used for requests.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithBearerToken configures an Authorization bearer token.
func WithBearerToken(token string) Option {
	return func(c *Client) {
		c.bearerToken = token
	}
}

// New creates a Client with the provided base URL.
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
		httpClient: &http.Client{},
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.httpClient == nil {
		c.httpClient = &http.Client{}
	}
	return c, nil
}

// CurrentUser loads the authenticated user resource.
func (c *Client) CurrentUser(ctx context.Context) (User, error) {
	var doc document[User]
	if _, err := c.get(ctx, "/v2/users/me/", &doc); err != nil {
		return User{}, err
	}
	return doc.Data, nil
}

// ListCurrentUserProjects loads all project-category nodes for the current user.
func (c *Client) ListCurrentUserProjects(ctx context.Context) ([]Node, error) {
	return collectPages[Node](ctx, c, "/v2/users/me/nodes/?filter[category]=project")
}

// GetNode loads a node/project by id.
func (c *Client) GetNode(ctx context.Context, id string) (Node, error) {
	var doc document[Node]
	if _, err := c.get(ctx, "/v2/nodes/"+url.PathEscape(id)+"/", &doc); err != nil {
		return Node{}, err
	}
	return doc.Data, nil
}

// ListNodeChildren loads all child components for a node.
func (c *Client) ListNodeChildren(ctx context.Context, id string) ([]Node, error) {
	return collectPages[Node](ctx, c, "/v2/nodes/"+url.PathEscape(id)+"/children/")
}

// ListNodeContributors loads all contributors for a node.
func (c *Client) ListNodeContributors(ctx context.Context, id string) ([]Contributor, error) {
	return collectPages[Contributor](ctx, c, "/v2/nodes/"+url.PathEscape(id)+"/contributors/")
}

// ListStorageFiles loads all files from a node's OSF Storage root or subfolder.
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
	var all []T
	next := endpoint
	for next != "" {
		var page document[[]T]
		respURL, err := c.get(ctx, next, &page)
		if err != nil {
			return nil, err
		}
		all = append(all, page.Data...)
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
	if c.bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.bearerToken)
	}
	req.Header.Set("Accept", "application/vnd.api+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

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
