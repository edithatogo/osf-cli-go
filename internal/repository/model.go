package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net/url"
	"strings"
)

// Provider identifies a repository service.
type Provider string

const (
	ProviderOSF    Provider = "osf"
	ProviderZenodo Provider = "zenodo"
)

// ResourceKind preserves a provider resource's semantic kind.
type ResourceKind string

const (
	KindComponent    ResourceKind = "component"
	KindDeposition   ResourceKind = "deposition"
	KindFile         ResourceKind = "file"
	KindProject      ResourceKind = "project"
	KindRecord       ResourceKind = "record"
	KindRegistration ResourceKind = "registration"
)

var (
	// ErrInvalidIdentity identifies malformed provider-qualified identities.
	ErrInvalidIdentity = errors.New("invalid provider-qualified identity")
	validKinds         = map[ResourceKind]bool{
		KindComponent: true, KindDeposition: true, KindFile: true,
		KindProject: true, KindRecord: true, KindRegistration: true,
	}
	providerKinds = map[Provider]map[ResourceKind]bool{
		ProviderOSF: {
			KindComponent: true, KindFile: true, KindProject: true, KindRegistration: true,
		},
		ProviderZenodo: {
			KindDeposition: true, KindFile: true, KindRecord: true,
		},
	}
)

// QualifiedID keeps a native identifier attached to its provider and kind.
type QualifiedID struct {
	Provider Provider     `json:"provider"`
	Kind     ResourceKind `json:"kind"`
	NativeID string       `json:"native_id"`
}

// Validate checks that the identity is explicit and recognized.
func (id QualifiedID) Validate() error {
	if id.Provider != ProviderOSF && id.Provider != ProviderZenodo {
		return fmt.Errorf("%w: unknown provider %q", ErrInvalidIdentity, id.Provider)
	}
	if !validKinds[id.Kind] {
		return fmt.Errorf("%w: unknown resource kind %q", ErrInvalidIdentity, id.Kind)
	}
	if !providerKinds[id.Provider][id.Kind] {
		return fmt.Errorf("%w: kind %q is not valid for provider %q", ErrInvalidIdentity, id.Kind, id.Provider)
	}
	if strings.TrimSpace(id.NativeID) == "" {
		return fmt.Errorf("%w: native id is required", ErrInvalidIdentity)
	}
	return nil
}

// Key returns an unambiguous, reversible provider-qualified key.
func (id QualifiedID) Key() (string, error) {
	if err := id.Validate(); err != nil {
		return "", err
	}
	return string(id.Provider) + ":" + string(id.Kind) + ":" + url.QueryEscape(id.NativeID), nil
}

// ParseQualifiedID parses a key created by QualifiedID.Key.
func ParseQualifiedID(value string) (QualifiedID, error) {
	parts := strings.SplitN(value, ":", 3)
	if len(parts) != 3 {
		return QualifiedID{}, fmt.Errorf("%w: expected provider:kind:native-id", ErrInvalidIdentity)
	}
	nativeID, err := url.QueryUnescape(parts[2])
	if err != nil {
		return QualifiedID{}, fmt.Errorf("%w: decode native id: %v", ErrInvalidIdentity, err)
	}
	id := QualifiedID{Provider: Provider(parts[0]), Kind: ResourceKind(parts[1]), NativeID: nativeID}
	if err := id.Validate(); err != nil {
		return QualifiedID{}, err
	}
	return id, nil
}

// NativeMetadata retains a provider's original metadata representation.
type NativeMetadata struct {
	mediaType string
	data      []byte
}

type nativeMetadataJSON struct {
	MediaType string `json:"media_type"`
	Data      []byte `json:"data"`
}

// NewNativeMetadata copies and validates provider-native metadata.
func NewNativeMetadata(mediaType string, data []byte) (NativeMetadata, error) {
	mediaType = strings.TrimSpace(mediaType)
	if mediaType == "" || len(data) == 0 {
		return NativeMetadata{}, errors.New("native metadata requires a media type and data")
	}
	parsedType, _, err := mime.ParseMediaType(mediaType)
	if err != nil {
		return NativeMetadata{}, fmt.Errorf("native metadata media type: %w", err)
	}
	if (parsedType == "application/json" || strings.HasSuffix(parsedType, "+json")) && !json.Valid(data) {
		return NativeMetadata{}, errors.New("native JSON metadata is invalid")
	}
	return NativeMetadata{mediaType: mediaType, data: append([]byte(nil), data...)}, nil
}

// MediaType reports the provider-native metadata media type.
func (m NativeMetadata) MediaType() string { return m.mediaType }

// Bytes returns an independent copy of the provider-native metadata.
func (m NativeMetadata) Bytes() []byte { return append([]byte(nil), m.data...) }

// MarshalJSON encodes native bytes losslessly using JSON's base64 byte encoding.
func (m NativeMetadata) MarshalJSON() ([]byte, error) {
	if m.mediaType == "" || len(m.data) == 0 {
		return nil, errors.New("native metadata is empty")
	}
	return json.Marshal(nativeMetadataJSON{MediaType: m.mediaType, Data: m.data})
}

// UnmarshalJSON restores and validates a lossless native metadata envelope.
func (m *NativeMetadata) UnmarshalJSON(data []byte) error {
	var envelope nativeMetadataJSON
	if err := json.Unmarshal(data, &envelope); err != nil {
		return err
	}
	metadata, err := NewNativeMetadata(envelope.MediaType, envelope.Data)
	if err != nil {
		return err
	}
	*m = metadata
	return nil
}

// LifecycleState is a deliberately small common lifecycle vocabulary.
type LifecycleState string

const (
	LifecycleActive     LifecycleState = "active"
	LifecycleDraft      LifecycleState = "draft"
	LifecycleEmbargoed  LifecycleState = "embargoed"
	LifecyclePublished  LifecycleState = "published"
	LifecycleRegistered LifecycleState = "registered"
	LifecycleWithdrawn  LifecycleState = "withdrawn"
)

var validLifecycleStates = map[LifecycleState]bool{
	LifecycleActive: true, LifecycleDraft: true, LifecycleEmbargoed: true,
	LifecyclePublished: true, LifecycleRegistered: true, LifecycleWithdrawn: true,
}

// Lifecycle combines a workflow-oriented state with the lossless native state.
type Lifecycle struct {
	Common LifecycleState `json:"common"`
	Native string         `json:"native"`
}

// Permission represents knowledge about a concrete provider permission.
type Permission string

const (
	PermissionAllowed     Permission = "allowed"
	PermissionDenied      Permission = "denied"
	PermissionUnknown     Permission = "unknown"
	PermissionUnsupported Permission = "unsupported"
)

var validPermissions = map[Permission]bool{
	PermissionAllowed: true, PermissionDenied: true,
	PermissionUnknown: true, PermissionUnsupported: true,
}

// Permissions preserves action-specific authorization knowledge.
type Permissions struct {
	Read    Permission `json:"read"`
	Write   Permission `json:"write"`
	Delete  Permission `json:"delete"`
	Publish Permission `json:"publish"`
}

// Link retains a provider-native relation and target.
type Link struct {
	Relation string `json:"relation"`
	URL      string `json:"url"`
}

// Checksum preserves a provider-selected checksum algorithm and value.
type Checksum struct {
	Algorithm string `json:"algorithm"`
	Value     string `json:"value"`
}

// VersionIdentity keeps native and persistent version identifiers separate.
type VersionIdentity struct {
	NativeVersionID string `json:"native_version_id,omitempty"`
	DOI             string `json:"doi,omitempty"`
	ConceptDOI      string `json:"concept_doi,omitempty"`
}

// RecordEnvelope contains normalized workflow fields and lossless native data.
type RecordEnvelope struct {
	Identity       QualifiedID     `json:"identity"`
	Title          string          `json:"title,omitempty"`
	Lifecycle      Lifecycle       `json:"lifecycle"`
	NativeMetadata NativeMetadata  `json:"native_metadata"`
	Links          []Link          `json:"links,omitempty"`
	Permissions    Permissions     `json:"permissions"`
	Version        VersionIdentity `json:"version,omitempty"`
	Checksums      []Checksum      `json:"checksums,omitempty"`
}

// Validate checks the common envelope without interpreting native metadata.
func (record RecordEnvelope) Validate() error {
	if err := record.Identity.Validate(); err != nil {
		return err
	}
	if !validLifecycleStates[record.Lifecycle.Common] || strings.TrimSpace(record.Lifecycle.Native) == "" {
		return errors.New("record lifecycle requires recognized common and non-empty native states")
	}
	if record.NativeMetadata.mediaType == "" || len(record.NativeMetadata.data) == 0 {
		return errors.New("record requires native metadata")
	}
	permissions := []struct {
		name  string
		value Permission
	}{
		{name: "read", value: record.Permissions.Read},
		{name: "write", value: record.Permissions.Write},
		{name: "delete", value: record.Permissions.Delete},
		{name: "publish", value: record.Permissions.Publish},
	}
	for _, permission := range permissions {
		if !validPermissions[permission.value] {
			return fmt.Errorf("record %s permission %q is invalid", permission.name, permission.value)
		}
	}
	return nil
}
