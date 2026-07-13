package mcpserver

import (
	"context"
	"errors"
	"net/url"
	"path"
	"strings"

	"github.com/edithatogo/osf-cli-go/internal/auth"
	"github.com/edithatogo/osf-cli-go/internal/osfapi"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// OSFClient is the read-only OSF API surface exposed through MCP tools.
type OSFClient interface {
	CurrentUser(context.Context) (osfapi.User, error)
	ListCurrentUserProjects(context.Context) ([]osfapi.Node, error)
	GetNode(context.Context, string) (osfapi.Node, error)
	ListNodeChildren(context.Context, string) ([]osfapi.Node, error)
	ListNodeContributors(context.Context, string) ([]osfapi.Contributor, error)
	ListStorageFiles(context.Context, string, ...string) ([]osfapi.StorageFile, error)
	SearchOSF(context.Context, string, ...int) ([]osfapi.SearchResult, error)
	ListPreprints(context.Context, string, ...int) ([]osfapi.Node, error)
}

type Server struct {
	client OSFClient
}

type Options struct {
	Version string
}

type EmptyInput struct{}

type NodeInput struct {
	ID string `json:"id" jsonschema:"OSF project/component id or URL"`
}

type FilesInput struct {
	ID   string `json:"id" jsonschema:"OSF project/component id or URL"`
	Path string `json:"path,omitempty" jsonschema:"optional path below OSF Storage"`
}

type SearchInput struct {
	Query string `json:"query" jsonschema:"OSF search query"`
	Limit int    `json:"limit,omitempty" jsonschema:"maximum number of results; zero uses the server default"`
}

type PreprintsInput struct {
	Provider string `json:"provider,omitempty" jsonschema:"optional preprint provider filter"`
	Limit    int    `json:"limit,omitempty" jsonschema:"maximum number of results; zero uses the server default"`
}

type UserOutput struct {
	ID       string `json:"id"`
	Type     string `json:"type,omitempty"`
	FullName string `json:"fullName,omitempty"`
	SelfURL  string `json:"selfUrl,omitempty"`
}

type NodeOutput struct {
	ID          string `json:"id"`
	Type        string `json:"type,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
	SelfURL     string `json:"selfUrl,omitempty"`
}

type ContributorOutput struct {
	ID             string `json:"id"`
	Type           string `json:"type,omitempty"`
	FullName       string `json:"fullName,omitempty"`
	Bibliographic  bool   `json:"bibliographic"`
	Permission     string `json:"permission,omitempty"`
	ProfileSelfURL string `json:"profileSelfUrl,omitempty"`
}

type FileOutput struct {
	ID          string `json:"id"`
	Type        string `json:"type,omitempty"`
	Name        string `json:"name,omitempty"`
	Kind        string `json:"kind,omitempty"`
	Size        int64  `json:"size,omitempty"`
	SelfURL     string `json:"selfUrl,omitempty"`
	DownloadURL string `json:"downloadUrl,omitempty"`
}

type UserResult struct {
	User UserOutput `json:"user"`
}

type NodesResult struct {
	Nodes []NodeOutput `json:"nodes"`
}

type NodeResult struct {
	Node NodeOutput `json:"node"`
}

type ContributorsResult struct {
	Contributors []ContributorOutput `json:"contributors"`
}

type FilesResult struct {
	Files []FileOutput `json:"files"`
}

type SearchResult struct {
	ID          string `json:"id"`
	Type        string `json:"type,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
	URL         string `json:"url,omitempty"`
}

type SearchResults struct {
	Results []SearchResult `json:"results"`
}

// New returns an MCP server with read-only OSF tools registered.
func New(client OSFClient, opts Options) *mcp.Server {
	version := strings.TrimSpace(opts.Version)
	if version == "" {
		version = "0.0.0-dev"
	}

	service := &Server{client: client}
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "osf-cli-go",
		Version: version,
		Title:   "OSF CLI Go MCP Server",
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "osf_whoami",
		Description: "Return the authenticated OSF user profile.",
	}, service.Whoami)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "osf_projects_list",
		Description: "List projects owned by the authenticated OSF user.",
	}, service.ListProjects)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "osf_project_get",
		Description: "Get an OSF project or component by id or URL.",
	}, service.GetProject)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "osf_components_list",
		Description: "List immediate child components for an OSF project or component.",
	}, service.ListComponents)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "osf_files_list",
		Description: "List OSF Storage files and folders for a project/component, optionally below a path.",
	}, service.ListFiles)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "osf_contributors_list",
		Description: "List contributors for an OSF project or component.",
	}, service.ListContributors)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "osf_search",
		Description: "Search public and authenticated OSF content.",
	}, service.Search)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "osf_preprints_list",
		Description: "List OSF preprints, optionally filtered by provider.",
	}, service.ListPreprints)

	return server
}

func (s *Server) Whoami(ctx context.Context, _ *mcp.CallToolRequest, _ EmptyInput) (*mcp.CallToolResult, UserResult, error) {
	user, err := s.client.CurrentUser(ctx)
	if err != nil {
		return nil, UserResult{}, mcpError(err)
	}
	return nil, UserResult{User: toUserOutput(user)}, nil
}

func (s *Server) ListProjects(ctx context.Context, _ *mcp.CallToolRequest, _ EmptyInput) (*mcp.CallToolResult, NodesResult, error) {
	nodes, err := s.client.ListCurrentUserProjects(ctx)
	if err != nil {
		return nil, NodesResult{}, mcpError(err)
	}
	return nil, NodesResult{Nodes: toNodeOutputs(nodes)}, nil
}

func (s *Server) GetProject(ctx context.Context, _ *mcp.CallToolRequest, in NodeInput) (*mcp.CallToolResult, NodeResult, error) {
	id, err := normalizeNodeID(in.ID)
	if err != nil {
		return nil, NodeResult{}, mcpError(err)
	}
	node, err := s.client.GetNode(ctx, id)
	if err != nil {
		return nil, NodeResult{}, mcpError(err)
	}
	return nil, NodeResult{Node: toNodeOutput(node)}, nil
}

func (s *Server) ListComponents(ctx context.Context, _ *mcp.CallToolRequest, in NodeInput) (*mcp.CallToolResult, NodesResult, error) {
	id, err := normalizeNodeID(in.ID)
	if err != nil {
		return nil, NodesResult{}, mcpError(err)
	}
	nodes, err := s.client.ListNodeChildren(ctx, id)
	if err != nil {
		return nil, NodesResult{}, mcpError(err)
	}
	return nil, NodesResult{Nodes: toNodeOutputs(nodes)}, nil
}

func (s *Server) ListFiles(ctx context.Context, _ *mcp.CallToolRequest, in FilesInput) (*mcp.CallToolResult, FilesResult, error) {
	id, err := normalizeNodeID(in.ID)
	if err != nil {
		return nil, FilesResult{}, mcpError(err)
	}
	segments, err := storagePathSegments(in.Path)
	if err != nil {
		return nil, FilesResult{}, mcpError(err)
	}
	files, err := s.client.ListStorageFiles(ctx, id, segments...)
	if err != nil {
		return nil, FilesResult{}, mcpError(err)
	}
	return nil, FilesResult{Files: toFileOutputs(files)}, nil
}

func (s *Server) ListContributors(ctx context.Context, _ *mcp.CallToolRequest, in NodeInput) (*mcp.CallToolResult, ContributorsResult, error) {
	id, err := normalizeNodeID(in.ID)
	if err != nil {
		return nil, ContributorsResult{}, mcpError(err)
	}
	contributors, err := s.client.ListNodeContributors(ctx, id)
	if err != nil {
		return nil, ContributorsResult{}, mcpError(err)
	}
	return nil, ContributorsResult{Contributors: toContributorOutputs(contributors)}, nil
}

func (s *Server) Search(ctx context.Context, _ *mcp.CallToolRequest, in SearchInput) (*mcp.CallToolResult, SearchResults, error) {
	query := strings.TrimSpace(in.Query)
	if query == "" {
		return nil, SearchResults{}, mcpError(errors.New("query is required"))
	}
	limit, err := boundedLimit(in.Limit)
	if err != nil {
		return nil, SearchResults{}, mcpError(err)
	}
	results, err := s.client.SearchOSF(ctx, query, limit)
	if err != nil {
		return nil, SearchResults{}, mcpError(err)
	}
	out := make([]SearchResult, 0, len(results))
	for _, result := range results {
		out = append(out, SearchResult{ID: result.ID, Type: result.Type, Title: result.Title, Description: result.Description, Category: result.Category, URL: result.URL})
	}
	return nil, SearchResults{Results: out}, nil
}

func (s *Server) ListPreprints(ctx context.Context, _ *mcp.CallToolRequest, in PreprintsInput) (*mcp.CallToolResult, NodesResult, error) {
	limit, err := boundedLimit(in.Limit)
	if err != nil {
		return nil, NodesResult{}, mcpError(err)
	}
	preprints, err := s.client.ListPreprints(ctx, strings.TrimSpace(in.Provider), limit)
	if err != nil {
		return nil, NodesResult{}, mcpError(err)
	}
	return nil, NodesResult{Nodes: toNodeOutputs(preprints)}, nil
}

func boundedLimit(limit int) (int, error) {
	if limit < 0 {
		return 0, errors.New("limit must be zero or greater")
	}
	if limit > 100 {
		return 0, errors.New("limit must be 100 or less")
	}
	return limit, nil
}

func mcpError(err error) error {
	return auth.RedactError(err)
}

func normalizeNodeID(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", errors.New("id is required")
	}

	parsed, err := url.Parse(value)
	if err == nil && parsed.Scheme != "" && parsed.Host != "" {
		parts := splitPath(parsed.Path)
		for i, part := range parts {
			if part == "nodes" && i+1 < len(parts) {
				return parts[i+1], nil
			}
		}
		if parsed.Host == "osf.io" && len(parts) > 0 {
			return parts[0], nil
		}
	}

	return strings.Trim(value, "/"), nil
}

func storagePathSegments(raw string) ([]string, error) {
	value := strings.Trim(strings.TrimSpace(raw), "/\\")
	if value == "" {
		return nil, nil
	}
	normalized := path.Clean(strings.ReplaceAll(value, "\\", "/"))
	if normalized == "." {
		return nil, nil
	}
	if normalized == ".." || strings.HasPrefix(normalized, "../") || strings.Contains(normalized, "/../") {
		return nil, errors.New("path must stay within OSF Storage")
	}
	return splitPath(normalized), nil
}

func splitPath(value string) []string {
	fields := strings.FieldsFunc(value, func(r rune) bool { return r == '/' || r == '\\' })
	parts := make([]string, 0, len(fields))
	for _, field := range fields {
		if field != "" {
			parts = append(parts, field)
		}
	}
	return parts
}

func toUserOutput(user osfapi.User) UserOutput {
	return UserOutput{
		ID:       user.ID,
		Type:     user.Type,
		FullName: user.Attributes.FullName,
		SelfURL:  user.Links.Self,
	}
}

func toNodeOutputs(nodes []osfapi.Node) []NodeOutput {
	out := make([]NodeOutput, 0, len(nodes))
	for _, node := range nodes {
		out = append(out, toNodeOutput(node))
	}
	return out
}

func toNodeOutput(node osfapi.Node) NodeOutput {
	return NodeOutput{
		ID:          node.ID,
		Type:        node.Type,
		Title:       node.Attributes.Title,
		Description: node.Attributes.Description,
		Category:    node.Attributes.Category,
		SelfURL:     node.Links.Self,
	}
}

func toContributorOutputs(contributors []osfapi.Contributor) []ContributorOutput {
	out := make([]ContributorOutput, 0, len(contributors))
	for _, contributor := range contributors {
		out = append(out, ContributorOutput{
			ID:             contributor.ID,
			Type:           contributor.Type,
			FullName:       contributor.Attributes.FullName,
			Bibliographic:  contributor.Attributes.Bibliographic,
			Permission:     contributor.Attributes.Permission,
			ProfileSelfURL: contributor.Links.Self,
		})
	}
	return out
}

func toFileOutputs(files []osfapi.StorageFile) []FileOutput {
	out := make([]FileOutput, 0, len(files))
	for _, file := range files {
		out = append(out, FileOutput{
			ID:          file.ID,
			Type:        file.Type,
			Name:        file.Attributes.Name,
			Kind:        file.Attributes.Kind,
			Size:        file.Attributes.Size,
			SelfURL:     file.Links.Self,
			DownloadURL: file.DownloadURL(),
		})
	}
	return out
}
