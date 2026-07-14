package mcpserver

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/edithatogo/osf-cli-go/internal/zenodoapi"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type fakeZenodoREST struct {
	query string
	limit int
	id    string
	err   error
	empty bool
}

func mcpZenodoRecord() zenodoapi.Record {
	raw := []byte(`{"id":"1001","conceptrecid":"900","doi":"10.5281/zenodo.1001","metadata":{"title":"Open methods","creators":[{"name":"Researcher"}],"keywords":["open"],"access_right":"open","license":{"id":"cc-by-4.0"}},"files":[{"id":"file-1","key":"data.csv","size":42,"checksum":"md5:abc","links":{"content":"https://zenodo.org/records/1001/files/data.csv"}}],"provider_extension":{"retained":true}}`)
	var record zenodoapi.Record
	_ = json.Unmarshal(raw, &record)
	return record
}

func (fake *fakeZenodoREST) SearchRecords(_ context.Context, query string, limit int) ([]zenodoapi.Record, error) {
	fake.query, fake.limit = query, limit
	if fake.err != nil {
		return nil, fake.err
	}
	if fake.empty {
		return []zenodoapi.Record{{}}, nil
	}
	return []zenodoapi.Record{mcpZenodoRecord()}, nil
}

func (fake *fakeZenodoREST) GetRecord(_ context.Context, id string) (zenodoapi.Record, error) {
	fake.id = id
	if fake.err != nil {
		return zenodoapi.Record{}, fake.err
	}
	if fake.empty {
		return zenodoapi.Record{}, nil
	}
	return mcpZenodoRecord(), nil
}

func (fake *fakeZenodoREST) ListRecordFiles(_ context.Context, id string) ([]zenodoapi.File, error) {
	fake.id = id
	if fake.err != nil {
		return nil, fake.err
	}
	return mcpZenodoRecord().Files, nil
}

func TestZenodoRESTToolsAndCapabilityNegotiation(t *testing.T) {
	session := connectTestServer(t, &fakeOSFClient{})
	tests := []struct {
		name      string
		arguments map[string]any
		field     string
	}{
		{name: "repository_capabilities_get", arguments: map[string]any{"provider": "zenodo"}, field: "contract"},
		{name: "zenodo_records_search", arguments: map[string]any{"query": "open", "limit": 5}, field: "records"},
		{name: "zenodo_record_get", arguments: map[string]any{"id": "zenodo:record:1001"}, field: "record"},
		{name: "zenodo_files_list", arguments: map[string]any{"id": "https://zenodo.org/records/1001"}, field: "files"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: test.name, Arguments: test.arguments})
			if err != nil || result.IsError {
				t.Fatalf("result=%#v err=%v", result, err)
			}
			content, ok := result.StructuredContent.(map[string]any)
			if !ok || content[test.field] == nil {
				t.Fatalf("content=%#v", result.StructuredContent)
			}
			rendered, _ := json.Marshal(content)
			if test.name == "zenodo_record_get" && (!strings.Contains(string(rendered), "zenodo:record:1001") || !strings.Contains(string(rendered), "provider_extension")) {
				t.Fatalf("lossy record: %s", rendered)
			}
			if test.name == "zenodo_files_list" && !strings.Contains(string(rendered), "zenodo:file:file-1") {
				t.Fatalf("unqualified file: %s", rendered)
			}
		})
	}
}

func TestZenodoMCPValidationUnsupportedAndRedaction(t *testing.T) {
	session := connectTestServer(t, &fakeOSFClient{})
	invalid := []struct {
		name      string
		arguments map[string]any
	}{
		{name: "repository_capabilities_get", arguments: map[string]any{"provider": "unknown"}},
		{name: "zenodo_records_search", arguments: map[string]any{"limit": 101}},
		{name: "zenodo_records_search", arguments: map[string]any{"query": strings.Repeat("x", 2049)}},
		{name: "zenodo_record_get", arguments: map[string]any{"id": "not-an-id"}},
	}
	for _, test := range invalid {
		result, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: test.name, Arguments: test.arguments})
		if err != nil || !result.IsError {
			t.Fatalf("%s result=%#v err=%v", test.name, result, err)
		}
	}

	for tool, err := range session.Tools(t.Context(), nil) {
		if err != nil {
			t.Fatal(err)
		}
		if strings.HasPrefix(tool.Name, "zenodo_") && (strings.Contains(tool.Name, "publish") || strings.Contains(tool.Name, "upload") || strings.Contains(tool.Name, "delete")) {
			t.Fatalf("deferred write tool advertised: %s", tool.Name)
		}
	}

	secret := "zenodo-secret-token-123456789"
	server := New(&fakeOSFClient{}, Options{Version: "test", ZenodoREST: &fakeZenodoREST{err: errors.New("request failed Authorization: Bearer " + secret)}})
	t1, t2 := mcp.NewInMemoryTransports()
	if _, err := server.Connect(t.Context(), t1, nil); err != nil {
		t.Fatal(err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "test"}, nil)
	redactedSession, err := client.Connect(t.Context(), t2, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = redactedSession.Close() }()
	result, err := redactedSession.CallTool(t.Context(), &mcp.CallToolParams{Name: "zenodo_record_get", Arguments: map[string]any{"id": "1001"}})
	if err != nil || !result.IsError || strings.Contains(contentText(result.Content), secret) || !strings.Contains(contentText(result.Content), "[REDACTED]") {
		t.Fatalf("result=%#v err=%v", result, err)
	}
}

func TestZenodoRESTHandlerErrorBranches(t *testing.T) {
	ctx := context.Background()
	server := &Server{zenodo: &fakeZenodoREST{}}
	_, capabilities, err := server.GetRepositoryCapabilities(ctx, nil, ProviderInput{Provider: "OSF"})
	if err != nil || capabilities.Contract.Provider != "osf" {
		t.Fatalf("capabilities=%#v err=%v", capabilities, err)
	}
	_, records, err := server.SearchZenodoRecords(ctx, nil, ZenodoSearchInput{})
	if err != nil || len(records.Records) != 1 {
		t.Fatalf("records=%#v err=%v", records, err)
	}

	server.zenodo = &fakeZenodoREST{err: errors.New("offline")}
	if _, _, err := server.SearchZenodoRecords(ctx, nil, ZenodoSearchInput{}); err == nil {
		t.Fatal("search backend error hidden")
	}
	if _, _, err := server.GetZenodoRecord(ctx, nil, ZenodoRecordInput{ID: "1001"}); err == nil {
		t.Fatal("get backend error hidden")
	}
	if _, _, err := server.ListZenodoFiles(ctx, nil, ZenodoRecordInput{ID: "1001"}); err == nil {
		t.Fatal("files backend error hidden")
	}

	server.zenodo = &fakeZenodoREST{empty: true}
	if _, _, err := server.SearchZenodoRecords(ctx, nil, ZenodoSearchInput{}); err == nil {
		t.Fatal("search accepted missing record id")
	}
	if _, _, err := server.GetZenodoRecord(ctx, nil, ZenodoRecordInput{ID: "1001"}); err == nil {
		t.Fatal("get accepted missing record id")
	}
	if _, _, err := server.ListZenodoFiles(ctx, nil, ZenodoRecordInput{ID: "bad"}); err == nil {
		t.Fatal("files accepted invalid id")
	}
}
