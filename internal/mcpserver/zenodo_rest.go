package mcpserver

import (
	"context"
	"errors"
	"strings"

	"github.com/edithatogo/osf-cli-go/internal/repository"
	"github.com/edithatogo/osf-cli-go/internal/zenodoapi"
	"github.com/edithatogo/osf-cli-go/internal/zenodoid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ZenodoRESTClient is the reviewed public published-record read surface.
type ZenodoRESTClient interface {
	SearchRecords(context.Context, string, int) ([]zenodoapi.Record, error)
	GetRecord(context.Context, string) (zenodoapi.Record, error)
	ListRecordFiles(context.Context, string) ([]zenodoapi.File, error)
}

type ProviderInput struct {
	Provider string `json:"provider" jsonschema:"repository provider: osf or zenodo"`
}

type ZenodoSearchInput struct {
	Query string `json:"query,omitempty" jsonschema:"optional Zenodo record search query"`
	Limit int    `json:"limit,omitempty" jsonschema:"maximum results from 1 to 100; zero defaults to 10"`
}

type ZenodoRecordInput struct {
	ID string `json:"id" jsonschema:"native decimal id, zenodo:record:id, or canonical Zenodo record URL"`
}

type ZenodoRecordOutput struct {
	QualifiedID        string                  `json:"qualifiedId"`
	Provider           repository.Provider     `json:"provider"`
	Kind               repository.ResourceKind `json:"kind"`
	ID                 string                  `json:"id"`
	ConceptID          string                  `json:"conceptId,omitempty"`
	Title              string                  `json:"title"`
	Description        string                  `json:"description,omitempty"`
	DOI                string                  `json:"doi,omitempty"`
	ConceptDOI         string                  `json:"conceptDoi,omitempty"`
	Creators           []zenodoapi.Creator     `json:"creators,omitempty"`
	Keywords           []string                `json:"keywords,omitempty"`
	AccessRight        string                  `json:"accessRight,omitempty"`
	License            string                  `json:"license,omitempty"`
	Created            string                  `json:"created,omitempty"`
	Updated            string                  `json:"updated,omitempty"`
	Links              map[string]string       `json:"links,omitempty"`
	NativeMetadataJSON string                  `json:"nativeMetadataJson"`
}

type ZenodoFileOutput struct {
	QualifiedID       string            `json:"qualifiedId,omitempty"`
	RecordQualifiedID string            `json:"recordQualifiedId"`
	ID                string            `json:"id,omitempty"`
	Key               string            `json:"key"`
	Size              int64             `json:"size,omitempty"`
	Checksum          string            `json:"checksum,omitempty"`
	DownloadURL       string            `json:"downloadUrl,omitempty"`
	Links             map[string]string `json:"links,omitempty"`
}

type RepositoryCapabilitiesResult struct {
	Contract repository.Contract `json:"contract"`
}
type ZenodoRecordsResult struct {
	Records []ZenodoRecordOutput `json:"records"`
}
type ZenodoRecordResult struct {
	Record ZenodoRecordOutput `json:"record"`
}
type ZenodoFilesResult struct {
	Files []ZenodoFileOutput `json:"files"`
}

func (s *Server) GetRepositoryCapabilities(ctx context.Context, _ *mcp.CallToolRequest, in ProviderInput) (*mcp.CallToolResult, RepositoryCapabilitiesResult, error) {
	var contract repository.Contract
	switch strings.ToLower(strings.TrimSpace(in.Provider)) {
	case string(repository.ProviderOSF):
		contract = repository.OSFContract()
	case string(repository.ProviderZenodo):
		contract = repository.ZenodoContract()
	default:
		return nil, RepositoryCapabilitiesResult{}, s.mcpError(ctx, errors.New("provider must be osf or zenodo"))
	}
	return nil, RepositoryCapabilitiesResult{Contract: contract}, nil
}

func (s *Server) SearchZenodoRecords(ctx context.Context, _ *mcp.CallToolRequest, in ZenodoSearchInput) (*mcp.CallToolResult, ZenodoRecordsResult, error) {
	query := strings.TrimSpace(in.Query)
	if len(query) > 2048 {
		return nil, ZenodoRecordsResult{}, s.mcpError(ctx, errors.New("query must be 2048 bytes or fewer"))
	}
	limit, err := boundedSearchLimit(in.Limit)
	if err != nil {
		return nil, ZenodoRecordsResult{}, s.mcpError(ctx, err)
	}
	records, err := s.zenodo.SearchRecords(ctx, query, limit)
	if err != nil {
		return nil, ZenodoRecordsResult{}, s.mcpError(ctx, err)
	}
	output := make([]ZenodoRecordOutput, 0, len(records))
	for _, record := range records {
		item, err := toZenodoRecordOutput(record)
		if err != nil {
			return nil, ZenodoRecordsResult{}, s.mcpError(ctx, err)
		}
		output = append(output, item)
	}
	return nil, ZenodoRecordsResult{Records: output}, nil
}

func (s *Server) GetZenodoRecord(ctx context.Context, _ *mcp.CallToolRequest, in ZenodoRecordInput) (*mcp.CallToolResult, ZenodoRecordResult, error) {
	id, err := zenodoid.ParseRecord(in.ID)
	if err != nil {
		return nil, ZenodoRecordResult{}, s.mcpError(ctx, err)
	}
	record, err := s.zenodo.GetRecord(ctx, id)
	if err != nil {
		return nil, ZenodoRecordResult{}, s.mcpError(ctx, err)
	}
	output, err := toZenodoRecordOutput(record)
	if err != nil {
		return nil, ZenodoRecordResult{}, s.mcpError(ctx, err)
	}
	return nil, ZenodoRecordResult{Record: output}, nil
}

func (s *Server) ListZenodoFiles(ctx context.Context, _ *mcp.CallToolRequest, in ZenodoRecordInput) (*mcp.CallToolResult, ZenodoFilesResult, error) {
	id, err := zenodoid.ParseRecord(in.ID)
	if err != nil {
		return nil, ZenodoFilesResult{}, s.mcpError(ctx, err)
	}
	files, err := s.zenodo.ListRecordFiles(ctx, id)
	if err != nil {
		return nil, ZenodoFilesResult{}, s.mcpError(ctx, err)
	}
	recordQualified, _ := qualifiedZenodoID(repository.KindRecord, id)
	output := make([]ZenodoFileOutput, 0, len(files))
	for _, file := range files {
		fileQualified := ""
		if strings.TrimSpace(file.ID) != "" {
			fileQualified, _ = qualifiedZenodoID(repository.KindFile, file.ID)
		}
		output = append(output, ZenodoFileOutput{QualifiedID: fileQualified, RecordQualifiedID: recordQualified, ID: file.ID, Key: file.Key, Size: file.Size, Checksum: file.Checksum, DownloadURL: file.ContentURL(), Links: file.Links})
	}
	return nil, ZenodoFilesResult{Files: output}, nil
}

func toZenodoRecordOutput(record zenodoapi.Record) (ZenodoRecordOutput, error) {
	qualified, err := qualifiedZenodoID(repository.KindRecord, record.ID)
	if err != nil {
		return ZenodoRecordOutput{}, err
	}
	native := record.NativeJSON()
	return ZenodoRecordOutput{QualifiedID: qualified, Provider: repository.ProviderZenodo, Kind: repository.KindRecord, ID: record.ID, ConceptID: record.ConceptRecID, Title: record.Metadata.Title, Description: record.Metadata.Description, DOI: record.DOI, ConceptDOI: record.ConceptDOI, Creators: append([]zenodoapi.Creator(nil), record.Metadata.Creators...), Keywords: append([]string(nil), record.Metadata.Keywords...), AccessRight: record.Metadata.AccessRight, License: record.Metadata.License.ID, Created: record.Created, Updated: record.Updated, Links: record.Links, NativeMetadataJSON: string(native)}, nil
}

func qualifiedZenodoID(kind repository.ResourceKind, native string) (string, error) {
	return (repository.QualifiedID{Provider: repository.ProviderZenodo, Kind: kind, NativeID: native}).Key()
}
