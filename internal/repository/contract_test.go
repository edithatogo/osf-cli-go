package repository

import (
	"bytes"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestQualifiedIDRoundTrip(t *testing.T) {
	want := QualifiedID{Provider: ProviderZenodo, Kind: KindRecord, NativeID: "10.5281/zenodo:123 45"}
	encoded, err := want.Key()
	if err != nil {
		t.Fatalf("Key() error = %v", err)
	}
	got, err := ParseQualifiedID(encoded)
	if err != nil {
		t.Fatalf("ParseQualifiedID() error = %v", err)
	}
	if got != want {
		t.Fatalf("ParseQualifiedID() = %#v, want %#v", got, want)
	}
}

func TestQualifiedIDRejectsUnknownProvider(t *testing.T) {
	_, err := (QualifiedID{Provider: "other", Kind: KindRecord, NativeID: "123"}).Key()
	if err == nil || !errors.Is(err, ErrInvalidIdentity) {
		t.Fatalf("Key() error = %v, want ErrInvalidIdentity", err)
	}
}

func TestQualifiedIDRejectsProviderKindMismatch(t *testing.T) {
	_, err := (QualifiedID{Provider: ProviderZenodo, Kind: KindProject, NativeID: "123"}).Key()
	if err == nil || !errors.Is(err, ErrInvalidIdentity) {
		t.Fatalf("Key() error = %v, want provider/kind ErrInvalidIdentity", err)
	}
}

func TestParseQualifiedIDRejectsMalformedKeys(t *testing.T) {
	for _, value := range []string{"missing", "osf:project:%zz", "zenodo:project:123", "osf:project: "} {
		if _, err := ParseQualifiedID(value); err == nil || !errors.Is(err, ErrInvalidIdentity) {
			t.Errorf("ParseQualifiedID(%q) error = %v", value, err)
		}
	}
}

func TestNativeMetadataPreservesAndCopiesJSON(t *testing.T) {
	source := []byte(`{"provider_field":{"nested":[1,true,"x"]}}`)
	metadata, err := NewNativeMetadata("application/json", source)
	if err != nil {
		t.Fatalf("NewNativeMetadata() error = %v", err)
	}
	source[2] = 'X'
	first := metadata.Bytes()
	if string(first) != `{"provider_field":{"nested":[1,true,"x"]}}` {
		t.Fatalf("Bytes() = %s", first)
	}
	first[0] = 'X'
	if string(metadata.Bytes()) != `{"provider_field":{"nested":[1,true,"x"]}}` {
		t.Fatal("Bytes() exposed mutable native metadata")
	}
}

func TestNativeMetadataRejectsInvalidJSON(t *testing.T) {
	_, err := NewNativeMetadata("application/json; charset=utf-8", []byte(`{"broken"`))
	if err == nil {
		t.Fatal("NewNativeMetadata() returned nil error for invalid JSON")
	}
}

func TestNativeMetadataRejectsInvalidEnvelopeAndSupportsOpaqueBytes(t *testing.T) {
	for _, test := range []struct {
		mediaType string
		data      []byte
	}{
		{mediaType: "", data: []byte("x")},
		{mediaType: "application/json", data: nil},
		{mediaType: "not a media type", data: []byte("x")},
	} {
		if _, err := NewNativeMetadata(test.mediaType, test.data); err == nil {
			t.Errorf("NewNativeMetadata(%q) returned nil error", test.mediaType)
		}
	}
	opaque, err := NewNativeMetadata("application/octet-stream", []byte{0, 255})
	if err != nil || opaque.MediaType() != "application/octet-stream" {
		t.Fatalf("opaque metadata = %#v, %v", opaque, err)
	}
	if _, err := json.Marshal(NativeMetadata{}); err == nil {
		t.Fatal("MarshalJSON() accepted empty metadata")
	}
	for _, value := range []string{`{`, `{"media_type":"application/json","data":"eA=="}`} {
		var metadata NativeMetadata
		if err := json.Unmarshal([]byte(value), &metadata); err == nil {
			t.Errorf("UnmarshalJSON(%q) returned nil error", value)
		}
	}
}

func TestNativeMetadataJSONRoundTripIsLossless(t *testing.T) {
	want, err := NewNativeMetadata("application/vnd.zenodo+json", []byte(`{"native": [1, 2, 3]}`))
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	var got NativeMetadata
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if got.MediaType() != want.MediaType() || !bytes.Equal(got.Bytes(), want.Bytes()) {
		t.Fatalf("round trip = %q %s, want %q %s", got.MediaType(), got.Bytes(), want.MediaType(), want.Bytes())
	}
}

func TestConcreteContractsAreCompleteAndValid(t *testing.T) {
	for _, contract := range []Contract{OSFContract(), ZenodoContract()} {
		if err := contract.Validate(); err != nil {
			t.Errorf("%s contract: %v", contract.Provider, err)
		}
		got := make([]Capability, 0, len(contract.Capabilities))
		for _, detail := range contract.Capabilities {
			got = append(got, detail.Capability)
		}
		if !slices.Equal(got, AllCapabilities()) {
			t.Errorf("%s capabilities = %v, want %v", contract.Provider, got, AllCapabilities())
		}
	}
}

func TestConcreteContractsExposeNonEquivalence(t *testing.T) {
	osf := OSFContract()
	zenodo := ZenodoContract()
	if got := osf.Support(CapabilityOAIHarvest).Level; got != SupportUnsupported {
		t.Fatalf("OSF OAI support = %q", got)
	}
	if got := zenodo.Support(CapabilityOAIHarvest).Level; got != SupportSupported {
		t.Fatalf("Zenodo OAI support = %q", got)
	}
	if got := osf.Support(CapabilityPublish).Level; got != SupportPartial {
		t.Fatalf("OSF publish support = %q", got)
	}
	if got := zenodo.Support(CapabilityPublish).Level; got != SupportPartial {
		t.Fatalf("Zenodo publish support = %q", got)
	}
}

func TestRequireCapabilityReturnsTypedError(t *testing.T) {
	err := OSFContract().Require(CapabilityOAIHarvest)
	if err == nil || !errors.Is(err, ErrUnsupportedCapability) {
		t.Fatalf("Require() error = %v, want ErrUnsupportedCapability", err)
	}
	var unsupported *CapabilitySupportError
	if !errors.As(err, &unsupported) || unsupported.Provider != ProviderOSF || unsupported.Capability != CapabilityOAIHarvest {
		t.Fatalf("Require() error = %#v", err)
	}
}

func TestRequireCapabilityDistinguishesPartialSupport(t *testing.T) {
	err := ZenodoContract().Require(CapabilityPublish)
	if err == nil || !errors.Is(err, ErrPartialCapability) || errors.Is(err, ErrUnsupportedCapability) {
		t.Fatalf("Require() error = %v, want only ErrPartialCapability", err)
	}
}

func TestCapabilitySupportErrorFormattingAndSupportedRequirement(t *testing.T) {
	if err := OSFContract().Require(CapabilityFileDownload); err != nil {
		t.Fatalf("Require() supported error = %v", err)
	}
	err := &CapabilitySupportError{Provider: ProviderZenodo, Capability: CapabilityPublish, Level: SupportPartial, Constraints: []string{"draft only", "confirmation required"}}
	if got := err.Error(); !strings.Contains(got, "draft only; confirmation required") {
		t.Fatalf("Error() = %q", got)
	}
}

func TestContractValidationRejectsIncompleteOrUnexplainedDecisions(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Contract)
	}{
		{name: "missing capability", mutate: func(contract *Contract) { contract.Capabilities = contract.Capabilities[:len(contract.Capabilities)-1] }},
		{name: "invalid level", mutate: func(contract *Contract) { contract.Capabilities[0].Level = "maybe" }},
		{name: "partial without reason", mutate: func(contract *Contract) {
			contract.Capabilities[0].Level = SupportPartial
			contract.Capabilities[0].Constraints = nil
		}},
		{name: "wrong order", mutate: func(contract *Contract) {
			contract.Capabilities[0], contract.Capabilities[1] = contract.Capabilities[1], contract.Capabilities[0]
		}},
		{name: "provider", mutate: func(contract *Contract) { contract.Provider = "other" }},
		{name: "model version", mutate: func(contract *Contract) { contract.ModelVersion = 2 }},
		{name: "empty constraint", mutate: func(contract *Contract) { contract.Capabilities[0].Constraints = []string{""} }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			contract := OSFContract()
			test.mutate(&contract)
			if err := contract.Validate(); err == nil {
				t.Fatal("Validate() returned nil error")
			}
		})
	}
}

func TestUnknownCapabilityIsExplicitlyUnsupported(t *testing.T) {
	detail := ZenodoContract().Support("future.operation")
	if detail.Level != SupportUnsupported || len(detail.Constraints) == 0 {
		t.Fatalf("Support() = %#v", detail)
	}
}

func TestRecordEnvelopeValidatesNativeIdentityAndState(t *testing.T) {
	metadata, err := NewNativeMetadata("application/json", []byte(`{"id":"abc"}`))
	if err != nil {
		t.Fatal(err)
	}
	record := RecordEnvelope{
		Identity:       QualifiedID{Provider: ProviderOSF, Kind: KindProject, NativeID: "abc"},
		Lifecycle:      Lifecycle{Common: LifecycleActive, Native: "public"},
		NativeMetadata: metadata,
		Permissions:    Permissions{Read: PermissionAllowed, Write: PermissionUnknown, Delete: PermissionUnknown, Publish: PermissionUnsupported},
	}
	if err := record.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	record.Lifecycle.Native = ""
	if err := record.Validate(); err == nil {
		t.Fatal("Validate() accepted missing native lifecycle state")
	}
}

func TestRecordEnvelopeJSONRoundTripPreservesNativeMetadata(t *testing.T) {
	metadata, err := NewNativeMetadata("application/octet-stream", []byte{0, 1, 2, 255})
	if err != nil {
		t.Fatal(err)
	}
	want := RecordEnvelope{
		Identity:       QualifiedID{Provider: ProviderZenodo, Kind: KindRecord, NativeID: "123"},
		Lifecycle:      Lifecycle{Common: LifecyclePublished, Native: "done"},
		NativeMetadata: metadata,
		Permissions:    Permissions{Read: PermissionAllowed, Write: PermissionDenied, Delete: PermissionDenied, Publish: PermissionUnsupported},
	}
	encoded, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var got RecordEnvelope
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatal(err)
	}
	if err := got.Validate(); err != nil {
		t.Fatalf("round-trip Validate() error = %v", err)
	}
	if !bytes.Equal(got.NativeMetadata.Bytes(), want.NativeMetadata.Bytes()) {
		t.Fatalf("native metadata = %v, want %v", got.NativeMetadata.Bytes(), want.NativeMetadata.Bytes())
	}
}

func TestRecordEnvelopeValidationFailures(t *testing.T) {
	metadata, err := NewNativeMetadata("application/json", []byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	valid := RecordEnvelope{
		Identity:       QualifiedID{Provider: ProviderOSF, Kind: KindProject, NativeID: "abc"},
		Lifecycle:      Lifecycle{Common: LifecycleActive, Native: "public"},
		NativeMetadata: metadata,
		Permissions:    Permissions{Read: PermissionAllowed, Write: PermissionUnknown, Delete: PermissionUnknown, Publish: PermissionUnsupported},
	}
	tests := []struct {
		name   string
		mutate func(*RecordEnvelope)
	}{
		{name: "identity", mutate: func(record *RecordEnvelope) { record.Identity.NativeID = "" }},
		{name: "common lifecycle", mutate: func(record *RecordEnvelope) { record.Lifecycle.Common = "other" }},
		{name: "native lifecycle", mutate: func(record *RecordEnvelope) { record.Lifecycle.Native = "" }},
		{name: "metadata", mutate: func(record *RecordEnvelope) { record.NativeMetadata = NativeMetadata{} }},
		{name: "permission", mutate: func(record *RecordEnvelope) { record.Permissions.Delete = "other" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			record := valid
			test.mutate(&record)
			if err := record.Validate(); err == nil {
				t.Fatal("Validate() returned nil error")
			}
		})
	}
}
