package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/edithatogo/osf-cli-go/internal/zenodooai"
)

type fakeOAIClient struct {
	request zenodooai.Request
	all     bool
	err     error
}

func (fake *fakeOAIClient) ListRecords(_ context.Context, request zenodooai.Request) (zenodooai.Page, error) {
	fake.request = request
	if fake.err != nil {
		return zenodooai.Page{}, fake.err
	}
	return zenodooai.Page{Records: []zenodooai.Record{{Header: zenodooai.Header{Identifier: "oai:zenodo.org:1001", Datestamp: "2026-07-15", SetSpecs: []string{"demo"}}}}, Next: zenodooai.ResumptionToken{Value: "next", MetadataPrefix: request.MetadataPrefix}}, nil
}
func (fake *fakeOAIClient) Harvest(_ context.Context, request zenodooai.Request) ([]zenodooai.Record, error) {
	fake.request, fake.all = request, true
	if fake.err != nil {
		return nil, fake.err
	}
	return []zenodooai.Record{{Header: zenodooai.Header{Identifier: "oai:zenodo.org:1001"}}}, nil
}
func (fake *fakeOAIClient) ListSets(context.Context) ([]zenodooai.Set, error) {
	if fake.err != nil {
		return nil, fake.err
	}
	return []zenodooai.Set{{Spec: "demo", Name: "Demo"}}, nil
}
func (fake *fakeOAIClient) ListMetadataFormats(context.Context, string) ([]zenodooai.MetadataFormat, error) {
	if fake.err != nil {
		return nil, fake.err
	}
	return []zenodooai.MetadataFormat{{Prefix: "oai_dc", Schema: "https://example.test/schema"}}, nil
}

func executeOAICommand(t *testing.T, fake *fakeOAIClient, args ...string) (string, error) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	root := newRootCommandWithClients(&stdout, &stderr, &fakeReadonlyClient{}, fake)
	root.SetArgs(args)
	err := root.ExecuteContext(context.Background())
	return stdout.String(), err
}

func TestZenodoOAIHarvestCLI(t *testing.T) {
	fake := &fakeOAIClient{}
	output, err := executeOAICommand(t, fake, "zenodo", "oai", "harvest", "--metadata-prefix", "datacite", "--set", "user-demo", "--from", "2026-07-01", "--until", "2026-07-15T00:00:00Z", "--output", "json")
	if err != nil {
		t.Fatal(err)
	}
	if fake.request.MetadataPrefix != "datacite" || fake.request.Set != "user-demo" || fake.request.From.IsZero() || fake.request.Until.IsZero() {
		t.Fatalf("request = %#v", fake.request)
	}
	if !strings.Contains(output, `"identifier":"oai:zenodo.org:1001"`) || !strings.Contains(output, `"value":"next"`) {
		t.Fatalf("output = %s", output)
	}

	fake = &fakeOAIClient{}
	output, err = executeOAICommand(t, fake, "zenodo", "oai", "harvest", "--all", "--json")
	if err != nil || !fake.all || !strings.Contains(output, "oai:zenodo.org:1001") {
		t.Fatalf("output=%q all=%v err=%v", output, fake.all, err)
	}
	if strings.Contains(output, `"next"`) {
		t.Fatalf("completed harvest has continuation: %s", output)
	}
}

func TestZenodoOAIAuxiliaryCLIAndValidation(t *testing.T) {
	fake := &fakeOAIClient{}
	for _, args := range [][]string{{"zenodo", "oai", "sets"}, {"zenodo", "oai", "formats", "--identifier", "oai:zenodo.org:1001", "--json"}} {
		output, err := executeOAICommand(t, fake, args...)
		if err != nil || strings.TrimSpace(output) == "" {
			t.Fatalf("args=%v output=%q err=%v", args, output, err)
		}
	}
	invalid := [][]string{
		{"zenodo", "oai", "harvest", "--resume-token", "token", "--set", "demo"},
		{"zenodo", "oai", "harvest", "--from", "yesterday"},
	}
	for _, args := range invalid {
		if _, err := executeOAICommand(t, fake, args...); err == nil {
			t.Fatalf("args %v succeeded", args)
		}
	}
	fake.err = errors.New("harvest unavailable")
	if _, err := executeOAICommand(t, fake, "zenodo", "oai", "sets"); err == nil {
		t.Fatal("backend error not returned")
	}
}
