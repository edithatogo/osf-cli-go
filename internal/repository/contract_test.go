package repository

import (
	"bytes"
	"encoding/json"
	"errors"
	"slices"
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
	var unsupported *UnsupportedCapabilityError
	if !errors.As(err, &unsupported) || unsupported.Provider != ProviderOSF || unsupported.Capability != CapabilityOAIHarvest {
		t.Fatalf("Require() error = %#v", err)
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
