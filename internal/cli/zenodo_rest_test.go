package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/edithatogo/osf-cli-go/internal/repository"
	"github.com/edithatogo/osf-cli-go/internal/zenodoapi"
)

type fakeZenodoREST struct {
	query string
	limit int
	id    string
	err   error
	empty bool
}

func sampleZenodoRecord() zenodoapi.Record {
	record := zenodoapi.Record{ID: "1001", ConceptRecID: "900", DOI: "10.5281/zenodo.1001", ConceptDOI: "10.5281/zenodo.900", Created: "2026-07-01", Updated: "2026-07-15", Metadata: zenodoapi.RecordMetadata{Title: "Open methods", Description: "Fixture record", Creators: []zenodoapi.Creator{{Name: "Researcher"}}, Keywords: []string{"open"}, AccessRight: "open", License: zenodoapi.License{ID: "cc-by-4.0"}}, Links: map[string]string{"self": "https://zenodo.org/api/records/1001"}, Files: []zenodoapi.File{{ID: "file-1", Key: "data.csv", Size: 42, Checksum: "md5:abc", Links: map[string]string{"content": "https://zenodo.org/records/1001/files/data.csv"}}}}
	data, _ := json.Marshal(record)
	var decoded zenodoapi.Record
	_ = json.Unmarshal(data, &decoded)
	return decoded
}

func (fake *fakeZenodoREST) SearchRecords(_ context.Context, query string, limit int) ([]zenodoapi.Record, error) {
	fake.query, fake.limit = query, limit
	if fake.err != nil {
		return nil, fake.err
	}
	if fake.empty {
		return []zenodoapi.Record{{}}, nil
	}
	return []zenodoapi.Record{sampleZenodoRecord()}, nil
}
func (fake *fakeZenodoREST) GetRecord(_ context.Context, id string) (zenodoapi.Record, error) {
	fake.id = id
	if fake.err != nil {
		return zenodoapi.Record{}, fake.err
	}
	if fake.empty {
		return zenodoapi.Record{}, nil
	}
	return sampleZenodoRecord(), nil
}
func (fake *fakeZenodoREST) ListRecordFiles(_ context.Context, id string) ([]zenodoapi.File, error) {
	fake.id = id
	if fake.err != nil {
		return nil, fake.err
	}
	return sampleZenodoRecord().Files, nil
}

func executeZenodoRESTCommand(t *testing.T, fake *fakeZenodoREST, args ...string) (string, error) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	root := newRootCommandWithProviders(&stdout, &stderr, &fakeReadonlyClient{}, fake, &fakeOAIClient{})
	root.SetArgs(args)
	err := root.ExecuteContext(context.Background())
	return stdout.String(), err
}

func TestZenodoRecordSearchGetAndFilesCLI(t *testing.T) {
	fake := &fakeZenodoREST{}
	output, err := executeZenodoRESTCommand(t, fake, "zenodo", "records", "search", "open science", "--limit", "5", "--json")
	if err != nil || fake.query != "open science" || fake.limit != 5 {
		t.Fatalf("output=%q fake=%#v err=%v", output, fake, err)
	}
	for _, want := range []string{`"qualifiedId":"zenodo:record:1001"`, `"provider":"zenodo"`, `"conceptDoi":"10.5281/zenodo.900"`, `"nativeMetadata":`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output missing %s: %s", want, output)
		}
	}

	output, err = executeZenodoRESTCommand(t, fake, "zenodo", "records", "get", "https://zenodo.org/records/1001")
	if err != nil || fake.id != "1001" || !strings.Contains(output, "Open methods") {
		t.Fatalf("output=%q id=%q err=%v", output, fake.id, err)
	}

	output, err = executeZenodoRESTCommand(t, fake, "zenodo", "files", "list", "zenodo:record:1001", "--json")
	if err != nil || fake.id != "1001" || !strings.Contains(output, `"qualifiedId":"zenodo:file:file-1"`) || !strings.Contains(output, `"recordQualifiedId":"zenodo:record:1001"`) || !strings.Contains(output, `"downloadUrl"`) {
		t.Fatalf("output=%q id=%q err=%v", output, fake.id, err)
	}
}

func TestZenodoCapabilitiesAndUnsupportedGuidance(t *testing.T) {
	output, err := executeZenodoRESTCommand(t, &fakeZenodoREST{}, "zenodo", "capabilities", "--json")
	if err != nil || !strings.Contains(output, `"provider":"zenodo"`) || !strings.Contains(output, `"capability":"records.search","level":"supported"`) {
		t.Fatalf("output=%q err=%v", output, err)
	}

	for _, args := range [][]string{{"zenodo", "records", "create"}, {"zenodo", "records", "update", "1001"}} {
		_, err := executeZenodoRESTCommand(t, &fakeZenodoREST{}, args...)
		var capabilityErr *repository.CapabilitySupportError
		if !errors.As(err, &capabilityErr) || !errors.Is(err, repository.ErrPartialCapability) || !strings.Contains(err.Error(), "provider zenodo capability") {
			t.Fatalf("args=%v error=%T %v", args, err, err)
		}
	}
}

func TestZenodoRecordIDAndCLIValidation(t *testing.T) {
	for input, want := range map[string]string{"1001": "1001", "zenodo:record:1001": "1001", "https://sandbox.zenodo.org/records/2002/": "2002"} {
		got, err := parseZenodoRecordID(input)
		if err != nil || got != want {
			t.Fatalf("parse(%q)=%q,%v want %q", input, got, err, want)
		}
	}
	for _, input := range []string{"", "abc", "zenodo:record:abc", "osf:project:abc", "https://example.com/records/1", "https://user@zenodo.org/records/1", "https://zenodo.org:444/records/1", "https://zenodo.org/deposits/1", "https://zenodo.org/records/1?q=x", "bad/id", "bad id"} {
		if _, err := parseZenodoRecordID(input); err == nil {
			t.Fatalf("parse(%q) succeeded", input)
		}
	}
	if _, err := executeZenodoRESTCommand(t, &fakeZenodoREST{}, "zenodo", "records", "search", "--limit", "0"); err == nil {
		t.Fatal("zero limit accepted")
	}
	if _, err := executeZenodoRESTCommand(t, &fakeZenodoREST{}, "zenodo", "records", "search", strings.Repeat("x", 2049)); err == nil {
		t.Fatal("oversized query accepted")
	}
	fake := &fakeZenodoREST{err: errors.New("backend unavailable")}
	if _, err := executeZenodoRESTCommand(t, fake, "zenodo", "records", "get", "1001"); err == nil {
		t.Fatal("backend error hidden")
	}
}

func TestZenodoHelpKeepsProviderScopeExplicit(t *testing.T) {
	output, err := executeZenodoRESTCommand(t, &fakeZenodoREST{}, "zenodo", "--help")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"records", "files", "deposits", "capabilities", "oai"} {
		if !strings.Contains(output, want) {
			t.Fatalf("help missing %q: %s", want, output)
		}
	}
}

func TestZenodoRESTOutputModesAndErrorBranches(t *testing.T) {
	fake := &fakeZenodoREST{}
	for _, args := range [][]string{
		{"zenodo", "records", "search"},
		{"zenodo", "records", "get", "zenodo:record:1001", "--json"},
		{"zenodo", "files", "list", "1001"},
		{"zenodo", "capabilities"},
	} {
		output, err := executeZenodoRESTCommand(t, fake, args...)
		if err != nil || strings.TrimSpace(output) == "" {
			t.Fatalf("args=%v output=%q err=%v", args, output, err)
		}
	}

	for _, args := range [][]string{{"zenodo", "records", "search"}, {"zenodo", "files", "list", "1001"}} {
		if _, err := executeZenodoRESTCommand(t, &fakeZenodoREST{err: errors.New("offline")}, args...); err == nil {
			t.Fatalf("args %v hid backend error", args)
		}
	}
	for _, args := range [][]string{{"zenodo", "records", "search"}, {"zenodo", "records", "get", "1001"}} {
		if _, err := executeZenodoRESTCommand(t, &fakeZenodoREST{empty: true}, args...); err == nil {
			t.Fatalf("args %v accepted missing record id", args)
		}
	}
	if _, err := parseZenodoRecordID("https://%zz"); err == nil {
		t.Fatal("malformed URL accepted")
	}
}
