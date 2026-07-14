package mcpserver

import (
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestZenodoOAIToolsReturnSeparateProtocolResults(t *testing.T) {
	session := connectTestServer(t, &fakeOSFClient{})
	tests := []struct {
		name      string
		arguments map[string]any
		field     string
	}{
		{name: "zenodo_oai_records_list", arguments: map[string]any{"metadataPrefix": "oai_dc"}, field: "records"},
		{name: "zenodo_oai_sets_list", field: "sets"},
		{name: "zenodo_oai_formats_list", arguments: map[string]any{"identifier": "oai:zenodo.org:1001"}, field: "formats"},
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
			if test.name == "zenodo_oai_records_list" && content["next"] != nil {
				t.Fatalf("completed page has continuation: %#v", content)
			}
		})
	}
}

func TestZenodoOAIRecordToolValidatesSelectors(t *testing.T) {
	session := connectTestServer(t, &fakeOSFClient{})
	for _, arguments := range []map[string]any{
		{"from": "not-a-date"},
		{"resumptionToken": "token", "set": "demo"},
	} {
		result, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: "zenodo_oai_records_list", Arguments: arguments})
		if err != nil || !result.IsError || !strings.Contains(contentText(result.Content), "must") && !strings.Contains(contentText(result.Content), "cannot") {
			t.Fatalf("result=%#v err=%v", result, err)
		}
	}
}
