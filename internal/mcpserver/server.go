package mcpserver

import (
	"context"
	"errors"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/edithatogo/osf-cli-go/internal/auth"
	"github.com/edithatogo/osf-cli-go/internal/observability"
	"github.com/edithatogo/osf-cli-go/internal/osfapi"
	"github.com/edithatogo/osf-cli-go/internal/zenodooai"
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
	ListFileVersions(context.Context, string) ([]osfapi.FileVersion, error)
	ListNodeAddons(context.Context, string) ([]osfapi.Node, error)
	ListWikiPages(context.Context, string) ([]osfapi.RelatedResource, error)
	ListNodeComments(context.Context, string) ([]osfapi.RelatedResource, error)
	ListNodeLogs(context.Context, string) ([]osfapi.RelatedResource, error)
	ListNodeIdentifiers(context.Context, string) ([]osfapi.RelatedResource, error)
	SearchOSF(context.Context, string, ...int) ([]osfapi.SearchResult, error)
	ListPreprints(context.Context, string, ...int) ([]osfapi.Node, error)
	SearchPreprints(context.Context, string, string, ...int) ([]osfapi.Preprint, error)
	ResolveDOI(context.Context, string) (osfapi.DOIResolution, error)
}

// ZenodoOAIClient is the public metadata-harvesting surface exposed separately from REST.
type ZenodoOAIClient interface {
	ListRecords(context.Context, zenodooai.Request) (zenodooai.Page, error)
	ListSets(context.Context) ([]zenodooai.Set, error)
	ListMetadataFormats(context.Context, string) ([]zenodooai.MetadataFormat, error)
}

type Server struct {
	client OSFClient
	oai    ZenodoOAIClient
	events observability.Emitter
}

type Options struct {
	Version   string
	Events    observability.Emitter
	ZenodoOAI ZenodoOAIClient
}

type EmptyInput struct{}

type NodeInput struct {
	ID string `json:"id" jsonschema:"OSF project/component id or URL"`
}

type FileInput struct {
	ID string `json:"id" jsonschema:"OSF file id"`
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

type PreprintSearchInput struct {
	Query    string `json:"query" jsonschema:"title search query"`
	Provider string `json:"provider,omitempty" jsonschema:"optional preprint provider filter"`
	Limit    int    `json:"limit,omitempty" jsonschema:"maximum number of results; zero uses the server default"`
}

type DOIInput struct {
	Identifier string `json:"identifier" jsonschema:"DOI, doi.org URL, or OSF DOI URL"`
}

type OAIRecordsInput struct {
	MetadataPrefix  string `json:"metadataPrefix,omitempty" jsonschema:"OAI metadata prefix; defaults to oai_dc"`
	Set             string `json:"set,omitempty" jsonschema:"optional OAI set spec"`
	From            string `json:"from,omitempty" jsonschema:"inclusive RFC3339 or YYYY-MM-DD start"`
	Until           string `json:"until,omitempty" jsonschema:"inclusive RFC3339 or YYYY-MM-DD end"`
	ResumptionToken string `json:"resumptionToken,omitempty" jsonschema:"opaque token from a prior page; exclusive with set/from/until"`
}

type OAIFormatsInput struct {
	Identifier string `json:"identifier,omitempty" jsonschema:"optional OAI identifier"`
}

type FileVersionOutput struct {
	ID           string    `json:"id"`
	Type         string    `json:"type,omitempty"`
	Size         int64     `json:"size,omitempty"`
	DateCreated  time.Time `json:"dateCreated,omitempty"`
	DateModified time.Time `json:"dateModified,omitempty"`
	SelfURL      string    `json:"selfUrl,omitempty"`
}

type RelatedResourceOutput struct {
	ID         string         `json:"id"`
	Type       string         `json:"type,omitempty"`
	Attributes map[string]any `json:"attributes,omitempty"`
	SelfURL    string         `json:"selfUrl,omitempty"`
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

type PreprintOutput struct {
	ID            string `json:"id"`
	Type          string `json:"type,omitempty"`
	Title         string `json:"title,omitempty"`
	DatePublished string `json:"datePublished,omitempty"`
	Published     bool   `json:"published"`
	DOI           string `json:"doi,omitempty"`
	URL           string `json:"url,omitempty"`
	Source        string `json:"source"`
}

type PreprintsResult struct {
	Preprints []PreprintOutput `json:"preprints"`
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
	MD5         string `json:"md5,omitempty"`
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

type FileVersionsResult struct {
	Versions []FileVersionOutput `json:"versions"`
}

type RelatedResourcesResult struct {
	Resources []RelatedResourceOutput `json:"resources"`
}

type SearchResult struct {
	ID          string   `json:"id"`
	Type        string   `json:"type,omitempty"`
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	Category    string   `json:"category,omitempty"`
	URL         string   `json:"url,omitempty"`
	Keywords    []string `json:"keywords,omitempty"`
	Year        string   `json:"year,omitempty"`
}

type SearchResults struct {
	Results []SearchResult `json:"results"`
}

type DOIResult struct {
	DOI         string `json:"doi"`
	ResolvedURL string `json:"resolvedUrl"`
}

type OAIRecordOutput struct {
	Identifier        string               `json:"identifier"`
	Datestamp         string               `json:"datestamp"`
	SetSpecs          []string             `json:"setSpecs,omitempty"`
	Deleted           bool                 `json:"deleted,omitempty"`
	NativeMetadataXML string               `json:"nativeMetadataXml,omitempty"`
	AboutXML          string               `json:"aboutXml,omitempty"`
	Provenance        zenodooai.Provenance `json:"provenance"`
}

type OAIRecordsResult struct {
	Records []OAIRecordOutput          `json:"records"`
	Next    *zenodooai.ResumptionToken `json:"next,omitempty"`
}

type OAISetsResult struct {
	Sets []zenodooai.Set `json:"sets"`
}
type OAIFormatsResult struct {
	Formats []zenodooai.MetadataFormat `json:"formats"`
}

// New returns an MCP server with read-only OSF tools registered.
func New(client OSFClient, opts Options) *mcp.Server {
	version := strings.TrimSpace(opts.Version)
	if version == "" {
		version = "0.0.0-dev"
	}

	service := &Server{client: client, oai: opts.ZenodoOAI, events: opts.Events}
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
	mcp.AddTool(server, &mcp.Tool{Name: "osf_file_versions_list", Description: "List versions for an OSF Storage file."}, service.ListFileVersions)
	mcp.AddTool(server, &mcp.Tool{Name: "osf_addons_list", Description: "List configured storage add-ons for an OSF node."}, service.ListAddons)
	mcp.AddTool(server, &mcp.Tool{Name: "osf_wikis_list", Description: "List wiki pages linked to an OSF node."}, service.ListWikis)
	mcp.AddTool(server, &mcp.Tool{Name: "osf_comments_list", Description: "List comments linked to an OSF node."}, service.ListComments)
	mcp.AddTool(server, &mcp.Tool{Name: "osf_logs_list", Description: "List audit logs linked to an OSF node."}, service.ListLogs)
	mcp.AddTool(server, &mcp.Tool{Name: "osf_identifiers_list", Description: "List identifiers linked to an OSF node."}, service.ListIdentifiers)
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
	mcp.AddTool(server, &mcp.Tool{
		Name:        "osf_preprints_search",
		Description: "Search OSF preprints by title, optionally filtered by provider.",
	}, service.SearchPreprints)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "osf_doi_resolve",
		Description: "Resolve a DOI to an OSF web resource without downloading or writing data.",
	}, service.ResolveDOI)
	if service.oai != nil {
		mcp.AddTool(server, &mcp.Tool{Name: "zenodo_oai_records_list", Description: "List one public Zenodo OAI-PMH record page and return its opaque continuation."}, service.ListOAIRecords)
		mcp.AddTool(server, &mcp.Tool{Name: "zenodo_oai_sets_list", Description: "List public Zenodo OAI-PMH selective-harvesting sets."}, service.ListOAISets)
		mcp.AddTool(server, &mcp.Tool{Name: "zenodo_oai_formats_list", Description: "List public Zenodo OAI-PMH metadata formats."}, service.ListOAIFormats)
	}

	return server
}

func (s *Server) Whoami(ctx context.Context, _ *mcp.CallToolRequest, _ EmptyInput) (*mcp.CallToolResult, UserResult, error) {
	user, err := s.client.CurrentUser(ctx)
	if err != nil {
		return nil, UserResult{}, s.mcpError(ctx, err)
	}
	return nil, UserResult{User: toUserOutput(user)}, nil
}

func (s *Server) ListProjects(ctx context.Context, _ *mcp.CallToolRequest, _ EmptyInput) (*mcp.CallToolResult, NodesResult, error) {
	nodes, err := s.client.ListCurrentUserProjects(ctx)
	if err != nil {
		return nil, NodesResult{}, s.mcpError(ctx, err)
	}
	return nil, NodesResult{Nodes: toNodeOutputs(nodes)}, nil
}

func (s *Server) GetProject(ctx context.Context, _ *mcp.CallToolRequest, in NodeInput) (*mcp.CallToolResult, NodeResult, error) {
	id, err := normalizeNodeID(in.ID)
	if err != nil {
		return nil, NodeResult{}, s.mcpError(ctx, err)
	}
	node, err := s.client.GetNode(ctx, id)
	if err != nil {
		return nil, NodeResult{}, s.mcpError(ctx, err)
	}
	return nil, NodeResult{Node: toNodeOutput(node)}, nil
}

func (s *Server) ListComponents(ctx context.Context, _ *mcp.CallToolRequest, in NodeInput) (*mcp.CallToolResult, NodesResult, error) {
	id, err := normalizeNodeID(in.ID)
	if err != nil {
		return nil, NodesResult{}, s.mcpError(ctx, err)
	}
	nodes, err := s.client.ListNodeChildren(ctx, id)
	if err != nil {
		return nil, NodesResult{}, s.mcpError(ctx, err)
	}
	return nil, NodesResult{Nodes: toNodeOutputs(nodes)}, nil
}

func (s *Server) ListFiles(ctx context.Context, _ *mcp.CallToolRequest, in FilesInput) (*mcp.CallToolResult, FilesResult, error) {
	id, err := normalizeNodeID(in.ID)
	if err != nil {
		return nil, FilesResult{}, s.mcpError(ctx, err)
	}
	segments, err := storagePathSegments(in.Path)
	if err != nil {
		return nil, FilesResult{}, s.mcpError(ctx, err)
	}
	files, err := s.client.ListStorageFiles(ctx, id, segments...)
	if err != nil {
		return nil, FilesResult{}, s.mcpError(ctx, err)
	}
	return nil, FilesResult{Files: toFileOutputs(files)}, nil
}

func (s *Server) ListContributors(ctx context.Context, _ *mcp.CallToolRequest, in NodeInput) (*mcp.CallToolResult, ContributorsResult, error) {
	id, err := normalizeNodeID(in.ID)
	if err != nil {
		return nil, ContributorsResult{}, s.mcpError(ctx, err)
	}
	contributors, err := s.client.ListNodeContributors(ctx, id)
	if err != nil {
		return nil, ContributorsResult{}, s.mcpError(ctx, err)
	}
	return nil, ContributorsResult{Contributors: toContributorOutputs(contributors)}, nil
}

func (s *Server) ListFileVersions(ctx context.Context, _ *mcp.CallToolRequest, in FileInput) (*mcp.CallToolResult, FileVersionsResult, error) {
	fileID := strings.TrimSpace(in.ID)
	if fileID == "" {
		return nil, FileVersionsResult{}, s.mcpError(ctx, errors.New("id is required"))
	}
	versions, err := s.client.ListFileVersions(ctx, fileID)
	if err != nil {
		return nil, FileVersionsResult{}, s.mcpError(ctx, err)
	}
	out := make([]FileVersionOutput, 0, len(versions))
	for _, version := range versions {
		out = append(out, FileVersionOutput{ID: version.ID, Type: version.Type, Size: version.Attributes.Size, DateCreated: version.Attributes.DateCreated, DateModified: version.Attributes.DateModified, SelfURL: version.Links.Self})
	}
	return nil, FileVersionsResult{Versions: out}, nil
}

func (s *Server) ListAddons(ctx context.Context, _ *mcp.CallToolRequest, in NodeInput) (*mcp.CallToolResult, NodesResult, error) {
	id, err := normalizeNodeID(in.ID)
	if err != nil {
		return nil, NodesResult{}, s.mcpError(ctx, err)
	}
	addons, err := s.client.ListNodeAddons(ctx, id)
	if err != nil {
		return nil, NodesResult{}, s.mcpError(ctx, err)
	}
	return nil, NodesResult{Nodes: toNodeOutputs(addons)}, nil
}

func (s *Server) ListWikis(ctx context.Context, _ *mcp.CallToolRequest, in NodeInput) (*mcp.CallToolResult, RelatedResourcesResult, error) {
	return s.listRelated(ctx, in.ID, s.client.ListWikiPages)
}

func (s *Server) ListComments(ctx context.Context, _ *mcp.CallToolRequest, in NodeInput) (*mcp.CallToolResult, RelatedResourcesResult, error) {
	return s.listRelated(ctx, in.ID, s.client.ListNodeComments)
}

func (s *Server) ListLogs(ctx context.Context, _ *mcp.CallToolRequest, in NodeInput) (*mcp.CallToolResult, RelatedResourcesResult, error) {
	return s.listRelated(ctx, in.ID, s.client.ListNodeLogs)
}

func (s *Server) ListIdentifiers(ctx context.Context, _ *mcp.CallToolRequest, in NodeInput) (*mcp.CallToolResult, RelatedResourcesResult, error) {
	return s.listRelated(ctx, in.ID, s.client.ListNodeIdentifiers)
}

func (s *Server) listRelated(ctx context.Context, rawID string, list func(context.Context, string) ([]osfapi.RelatedResource, error)) (*mcp.CallToolResult, RelatedResourcesResult, error) {
	id, err := normalizeNodeID(rawID)
	if err != nil {
		return nil, RelatedResourcesResult{}, s.mcpError(ctx, err)
	}
	resources, err := list(ctx, id)
	if err != nil {
		return nil, RelatedResourcesResult{}, s.mcpError(ctx, err)
	}
	out := make([]RelatedResourceOutput, 0, len(resources))
	for _, resource := range resources {
		out = append(out, RelatedResourceOutput{ID: resource.ID, Type: resource.Type, Attributes: resource.Attributes, SelfURL: resource.Links.Self})
	}
	return nil, RelatedResourcesResult{Resources: out}, nil
}

func (s *Server) Search(ctx context.Context, _ *mcp.CallToolRequest, in SearchInput) (*mcp.CallToolResult, SearchResults, error) {
	query := strings.TrimSpace(in.Query)
	if query == "" {
		return nil, SearchResults{}, s.mcpError(ctx, errors.New("query is required"))
	}
	limit, err := boundedLimit(in.Limit)
	if err != nil {
		return nil, SearchResults{}, s.mcpError(ctx, err)
	}
	results, err := s.client.SearchOSF(ctx, query, limit)
	if err != nil {
		return nil, SearchResults{}, s.mcpError(ctx, err)
	}
	out := make([]SearchResult, 0, len(results))
	for _, result := range results {
		out = append(out, SearchResult{ID: result.ID, Type: result.Type, Title: result.Title, Description: result.Description, Category: result.Category, URL: result.URL, Keywords: result.Keywords, Year: result.Year})
	}
	return nil, SearchResults{Results: out}, nil
}

func (s *Server) ListPreprints(ctx context.Context, _ *mcp.CallToolRequest, in PreprintsInput) (*mcp.CallToolResult, NodesResult, error) {
	limit, err := boundedLimit(in.Limit)
	if err != nil {
		return nil, NodesResult{}, s.mcpError(ctx, err)
	}
	preprints, err := s.client.ListPreprints(ctx, strings.TrimSpace(in.Provider), limit)
	if err != nil {
		return nil, NodesResult{}, s.mcpError(ctx, err)
	}
	return nil, NodesResult{Nodes: toNodeOutputs(preprints)}, nil
}

func (s *Server) SearchPreprints(ctx context.Context, _ *mcp.CallToolRequest, in PreprintSearchInput) (*mcp.CallToolResult, PreprintsResult, error) {
	query := strings.TrimSpace(in.Query)
	if query == "" {
		return nil, PreprintsResult{}, s.mcpError(ctx, errors.New("query is required"))
	}
	limit, err := boundedSearchLimit(in.Limit)
	if err != nil {
		return nil, PreprintsResult{}, s.mcpError(ctx, err)
	}
	preprints, err := s.client.SearchPreprints(ctx, query, strings.TrimSpace(in.Provider), limit)
	if err != nil {
		return nil, PreprintsResult{}, s.mcpError(ctx, err)
	}
	out := make([]PreprintOutput, 0, len(preprints))
	for _, preprint := range preprints {
		out = append(out, PreprintOutput{
			ID:            preprint.ID,
			Type:          preprint.Type,
			Title:         preprint.Attributes.Title,
			DatePublished: preprint.Attributes.DatePublished,
			Published:     preprint.Attributes.IsPublished,
			DOI:           preprint.Attributes.DOI,
			URL:           preprint.Links.HTML,
			Source:        "OSF Preprints",
		})
	}
	return nil, PreprintsResult{Preprints: out}, nil
}

func (s *Server) ResolveDOI(ctx context.Context, _ *mcp.CallToolRequest, in DOIInput) (*mcp.CallToolResult, DOIResult, error) {
	identifier := strings.TrimSpace(in.Identifier)
	if identifier == "" {
		return nil, DOIResult{}, s.mcpError(ctx, errors.New("identifier is required"))
	}
	resolution, err := s.client.ResolveDOI(ctx, identifier)
	if err != nil {
		return nil, DOIResult{}, s.mcpError(ctx, err)
	}
	return nil, DOIResult{DOI: resolution.DOI, ResolvedURL: resolution.ResolvedURL}, nil
}

func (s *Server) ListOAIRecords(ctx context.Context, _ *mcp.CallToolRequest, in OAIRecordsInput) (*mcp.CallToolResult, OAIRecordsResult, error) {
	request, err := oaiRequest(in)
	if err != nil {
		return nil, OAIRecordsResult{}, s.mcpError(ctx, err)
	}
	page, err := s.oai.ListRecords(ctx, request)
	if err != nil {
		return nil, OAIRecordsResult{}, s.mcpError(ctx, err)
	}
	result := OAIRecordsResult{Records: make([]OAIRecordOutput, 0, len(page.Records))}
	if !page.Next.Empty() {
		result.Next = &page.Next
	}
	for _, record := range page.Records {
		var native string
		if record.NativeMetadata != nil {
			native = string(record.NativeMetadata.Bytes())
		}
		result.Records = append(result.Records, OAIRecordOutput{Identifier: record.Header.Identifier, Datestamp: record.Header.Datestamp, SetSpecs: append([]string(nil), record.Header.SetSpecs...), Deleted: record.Header.Deleted, NativeMetadataXML: native, AboutXML: string(record.AboutXML), Provenance: record.Provenance})
	}
	return nil, result, nil
}

func (s *Server) ListOAISets(ctx context.Context, _ *mcp.CallToolRequest, _ EmptyInput) (*mcp.CallToolResult, OAISetsResult, error) {
	sets, err := s.oai.ListSets(ctx)
	if err != nil {
		return nil, OAISetsResult{}, s.mcpError(ctx, err)
	}
	return nil, OAISetsResult{Sets: sets}, nil
}

func (s *Server) ListOAIFormats(ctx context.Context, _ *mcp.CallToolRequest, in OAIFormatsInput) (*mcp.CallToolResult, OAIFormatsResult, error) {
	formats, err := s.oai.ListMetadataFormats(ctx, strings.TrimSpace(in.Identifier))
	if err != nil {
		return nil, OAIFormatsResult{}, s.mcpError(ctx, err)
	}
	return nil, OAIFormatsResult{Formats: formats}, nil
}

func oaiRequest(in OAIRecordsInput) (zenodooai.Request, error) {
	prefix := strings.TrimSpace(in.MetadataPrefix)
	if prefix == "" {
		prefix = "oai_dc"
	}
	if token := strings.TrimSpace(in.ResumptionToken); token != "" {
		if strings.TrimSpace(in.Set) != "" || strings.TrimSpace(in.From) != "" || strings.TrimSpace(in.Until) != "" {
			return zenodooai.Request{}, errors.New("resumptionToken cannot be combined with set, from, or until")
		}
		return zenodooai.Request{Token: zenodooai.ResumptionToken{Value: token, MetadataPrefix: prefix}}, nil
	}
	from, err := parseOAIDate(in.From)
	if err != nil {
		return zenodooai.Request{}, errors.New("from must be RFC3339 or YYYY-MM-DD")
	}
	until, err := parseOAIDate(in.Until)
	if err != nil {
		return zenodooai.Request{}, errors.New("until must be RFC3339 or YYYY-MM-DD")
	}
	return zenodooai.Request{MetadataPrefix: prefix, Set: strings.TrimSpace(in.Set), From: from, Until: until}, nil
}

func parseOAIDate(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}
	for _, layout := range []string{time.RFC3339, time.DateOnly} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, errors.New("invalid OAI date")
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

func boundedSearchLimit(limit int) (int, error) {
	if limit == 0 {
		return 10, nil
	}
	if limit < 1 || limit > 100 {
		return 0, errors.New("limit must be between 1 and 100")
	}
	return limit, nil
}

func (s *Server) mcpError(ctx context.Context, err error) error {
	redacted := auth.RedactError(err)
	observability.Emit(ctx, s.events, observability.Event{
		Level:   observability.LevelError,
		Name:    "mcp.tool.error",
		Outcome: observability.OutcomeError,
		Error:   observability.RedactedError(redacted),
	})
	return redacted
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
			MD5:         file.Attributes.Extra.Hashes.MD5,
			SelfURL:     file.Links.Self,
			DownloadURL: file.DownloadURL(),
		})
	}
	return out
}
