package crossprovider

import (
	"errors"
	"testing"
	"time"

	"github.com/edithatogo/osf-cli-go/internal/repository"
)

func TestBuildMappingRequiresExplicitBoundaries(t *testing.T) {
	t.Parallel()
	valid := validRequest(t)
	tests := []struct {
		name   string
		mutate func(*Request)
	}{
		{name: "direction", mutate: func(r *Request) { r.Direction = "" }},
		{name: "destination", mutate: func(r *Request) { r.Destination = Destination{} }},
		{name: "same provider", mutate: func(r *Request) { r.Destination.Provider = repository.ProviderOSF }},
		{name: "authorization", mutate: func(r *Request) { r.Authorized = false }},
		{name: "publish intent", mutate: func(r *Request) { r.PublishIntent = "" }},
		{name: "conflict", mutate: func(r *Request) { r.Conflict = "" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			request := valid
			test.mutate(&request)
			if _, err := BuildMapping(request, time.Now()); !errors.Is(err, ErrInvalidRequest) {
				t.Fatalf("BuildMapping error = %v", err)
			}
		})
	}
}

func TestOSFToZenodoDryRunAccountsForEveryField(t *testing.T) {
	t.Parallel()
	request := validRequest(t)
	report, err := BuildMapping(request, time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if !report.Executable || report.IdempotencyKey == "" || report.Provenance.NativeMetadataSHA256 == "" {
		t.Fatalf("report = %+v", report)
	}
	wantFields := []string{"title", "description", "upload_type", "creators", "keywords", "access", "license", "embargo", "identifiers", "version", "native_metadata", "files"}
	if len(report.Fields) != len(wantFields) {
		t.Fatalf("field count = %d, want %d", len(report.Fields), len(wantFields))
	}
	for i, field := range wantFields {
		if report.Fields[i].SourceField != field || report.Fields[i].Disposition == "" {
			t.Fatalf("field %d = %+v", i, report.Fields[i])
		}
	}
	if report.Target.Access.Kind != AccessOpen || report.Target.License != "cc-by-4.0" || report.Target.Title != request.Source.Metadata.Title {
		t.Fatalf("target = %+v", report.Target)
	}
}

func TestPrivateOSFSourceRequiresExplicitZenodoAccess(t *testing.T) {
	t.Parallel()
	request := validRequest(t)
	request.Source.Metadata.Access = AccessPolicy{Kind: AccessPrivate}
	request.TargetAccess = nil
	report, err := BuildMapping(request, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if report.Executable || len(report.Blockers) == 0 || mappingFor(report, "access").Disposition != DispositionBlocked {
		t.Fatalf("report = %+v", report)
	}
	request.TargetAccess = &AccessPolicy{Kind: AccessRestricted, Conditions: "available after review"}
	report, err = BuildMapping(request, time.Now())
	if err != nil || !report.Executable || report.Target.Access.Kind != AccessRestricted {
		t.Fatalf("explicit access report = %+v err=%v", report, err)
	}
}

func TestZenodoToOSFPreservesUnsupportedSemantics(t *testing.T) {
	t.Parallel()
	request := validRequest(t)
	request.Direction = DirectionZenodoToOSF
	request.Source.Identity.Provider = repository.ProviderZenodo
	request.Source.Identity.Kind = repository.KindRecord
	request.Destination.Provider = repository.ProviderOSF
	request.Source.Metadata.Access = AccessPolicy{Kind: AccessEmbargoed, EmbargoUntil: datePtr(time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC))}
	report, err := BuildMapping(request, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if !report.Executable || report.Target.Access.Kind != AccessPrivate {
		t.Fatalf("report = %+v", report)
	}
	for _, field := range []string{"license", "embargo", "identifiers", "native_metadata"} {
		mapping := mappingFor(report, field)
		if mapping.Disposition != DispositionPreservedNative || mapping.Reason == "" {
			t.Fatalf("mapping %s = %+v", field, mapping)
		}
	}
}

func TestIdempotencyKeyIsDeterministicAndContentSensitive(t *testing.T) {
	t.Parallel()
	request := validRequest(t)
	first, err := BuildMapping(request, time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	second, err := BuildMapping(request, time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if first.IdempotencyKey != second.IdempotencyKey || first.Provenance.CapturedAt.Equal(second.Provenance.CapturedAt) {
		t.Fatalf("keys/times first=%+v second=%+v", first, second)
	}
	request.Source.Files[0].Checksum = "sha256:changed"
	changed, err := BuildMapping(request, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if changed.IdempotencyKey == first.IdempotencyKey {
		t.Fatal("content change did not change idempotency key")
	}
}

func validRequest(t *testing.T) Request {
	t.Helper()
	native, err := repository.NewNativeMetadata("application/json", []byte(`{"id":"abc","providerOnly":true}`))
	if err != nil {
		t.Fatal(err)
	}
	return Request{
		Direction: DirectionOSFToZenodo,
		Source: Snapshot{
			Identity:       repository.QualifiedID{Provider: repository.ProviderOSF, Kind: repository.KindProject, NativeID: "abc"},
			Metadata:       Metadata{Title: "Source project", Description: "Preserved description", UploadType: "dataset", Creators: []Creator{{Name: "Researcher, Example"}}, Keywords: []string{"open-science"}, Access: AccessPolicy{Kind: AccessPublic}, License: "cc-by-4.0", Identifiers: []Identifier{{Scheme: "doi", Value: "10.1234/source"}}, Version: "v2"},
			Files:          []File{{Path: "data.csv", Size: 42, Checksum: "sha256:abc", MediaType: "text/csv"}},
			NativeMetadata: native,
		},
		Destination: Destination{Provider: repository.ProviderZenodo, CreateNew: true},
		Authorized:  true, PublishIntent: PublishDraftOnly, Conflict: ConflictFail,
	}
}

func mappingFor(report Report, field string) FieldMapping {
	for _, mapping := range report.Fields {
		if mapping.SourceField == field {
			return mapping
		}
	}
	return FieldMapping{}
}

func datePtr(value time.Time) *time.Time { return &value }
