package repository

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

// Capability identifies a provider operation without implying support.
type Capability string

const (
	CapabilityFileDelete     Capability = "files.delete"
	CapabilityFileDownload   Capability = "files.download"
	CapabilityFileList       Capability = "files.list"
	CapabilityFileUpload     Capability = "files.upload"
	CapabilityPublish        Capability = "lifecycle.publish"
	CapabilityMetadataUpdate Capability = "metadata.update"
	CapabilityOAIHarvest     Capability = "oai.harvest"
	CapabilityRecordCreate   Capability = "records.create"
	CapabilityRecordDelete   Capability = "records.delete"
	CapabilityRecordGet      Capability = "records.get"
	CapabilityRecordSearch   Capability = "records.search"
	CapabilityRecordUpdate   Capability = "records.update"
	CapabilityVersionCreate  Capability = "versions.create"
)

var capabilities = []Capability{
	CapabilityFileDelete,
	CapabilityFileDownload,
	CapabilityFileList,
	CapabilityFileUpload,
	CapabilityPublish,
	CapabilityMetadataUpdate,
	CapabilityOAIHarvest,
	CapabilityRecordCreate,
	CapabilityRecordDelete,
	CapabilityRecordGet,
	CapabilityRecordSearch,
	CapabilityRecordUpdate,
	CapabilityVersionCreate,
}

// AllCapabilities returns a copy of the complete shared vocabulary.
func AllCapabilities() []Capability { return append([]Capability(nil), capabilities...) }

// SupportLevel distinguishes full, constrained, and absent behavior.
type SupportLevel string

const (
	SupportPartial     SupportLevel = "partial"
	SupportSupported   SupportLevel = "supported"
	SupportUnsupported SupportLevel = "unsupported"
)

// CapabilitySupport describes a concrete provider's operation semantics.
type CapabilitySupport struct {
	Capability  Capability   `json:"capability"`
	Level       SupportLevel `json:"level"`
	Constraints []string     `json:"constraints,omitempty"`
}

// Contract is a concrete provider capability descriptor, not a client interface.
type Contract struct {
	Provider     Provider            `json:"provider"`
	ModelVersion int                 `json:"model_version"`
	Capabilities []CapabilitySupport `json:"capabilities"`
}

// Support returns explicit support details, including for unknown capabilities.
func (contract Contract) Support(capability Capability) CapabilitySupport {
	index, found := slices.BinarySearchFunc(contract.Capabilities, capability, func(detail CapabilitySupport, target Capability) int {
		return strings.Compare(string(detail.Capability), string(target))
	})
	if found {
		return contract.Capabilities[index]
	}
	return CapabilitySupport{Capability: capability, Level: SupportUnsupported, Constraints: []string{"capability is outside the reviewed vocabulary"}}
}

var (
	// ErrUnsupportedCapability identifies operations absent from a provider.
	ErrUnsupportedCapability = errors.New("repository capability is unsupported")
	// ErrPartialCapability identifies operations requiring provider-specific handling.
	ErrPartialCapability = errors.New("repository capability is only partially supported")
)

// CapabilitySupportError provides capability-aware user guidance.
type CapabilitySupportError struct {
	Provider    Provider
	Capability  Capability
	Level       SupportLevel
	Constraints []string
}

func (e *CapabilitySupportError) Error() string {
	return fmt.Sprintf("provider %s capability %s is %s: %s", e.Provider, e.Capability, e.Level, strings.Join(e.Constraints, "; "))
}

func (e *CapabilitySupportError) Unwrap() error {
	if e.Level == SupportPartial {
		return ErrPartialCapability
	}
	return ErrUnsupportedCapability
}

// Require succeeds only for fully supported behavior.
func (contract Contract) Require(capability Capability) error {
	detail := contract.Support(capability)
	if detail.Level == SupportSupported {
		return nil
	}
	return &CapabilitySupportError{
		Provider: contract.Provider, Capability: capability,
		Level: detail.Level, Constraints: append([]string(nil), detail.Constraints...),
	}
}

// Validate ensures every vocabulary entry has exactly one explicit decision.
func (contract Contract) Validate() error {
	if contract.Provider != ProviderOSF && contract.Provider != ProviderZenodo {
		return fmt.Errorf("invalid contract provider %q", contract.Provider)
	}
	if contract.ModelVersion != 1 {
		return fmt.Errorf("model version = %d, want 1", contract.ModelVersion)
	}
	if len(contract.Capabilities) != len(capabilities) {
		return fmt.Errorf("capability count = %d, want %d", len(contract.Capabilities), len(capabilities))
	}
	for i, detail := range contract.Capabilities {
		if detail.Capability != capabilities[i] {
			return fmt.Errorf("capability %d = %q, want %q", i+1, detail.Capability, capabilities[i])
		}
		if detail.Level != SupportSupported && detail.Level != SupportPartial && detail.Level != SupportUnsupported {
			return fmt.Errorf("capability %q has invalid support level %q", detail.Capability, detail.Level)
		}
		if detail.Level != SupportSupported && len(detail.Constraints) == 0 {
			return fmt.Errorf("capability %q requires a reason for %s support", detail.Capability, detail.Level)
		}
		for _, constraint := range detail.Constraints {
			if strings.TrimSpace(constraint) == "" {
				return fmt.Errorf("capability %q has an empty constraint", detail.Capability)
			}
		}
	}
	return nil
}

func support(capability Capability, level SupportLevel, constraints ...string) CapabilitySupport {
	return CapabilitySupport{Capability: capability, Level: level, Constraints: constraints}
}

// OSFContract reports reviewed OSF behavior without changing the existing API client.
func OSFContract() Contract {
	return Contract{Provider: ProviderOSF, ModelVersion: 1, Capabilities: []CapabilitySupport{
		support(CapabilityFileDelete, SupportSupported, "requires explicit destructive command and authorization"),
		support(CapabilityFileDownload, SupportSupported, "OSF Storage uses WaterButler links and checksums"),
		support(CapabilityFileList, SupportSupported, "storage provider remains visible"),
		support(CapabilityFileUpload, SupportSupported, "conflict behavior is explicit"),
		support(CapabilityPublish, SupportPartial, "project visibility and registration are distinct OSF workflows"),
		support(CapabilityMetadataUpdate, SupportSupported, "fields depend on OSF entity and permissions"),
		support(CapabilityOAIHarvest, SupportUnsupported, "OSF adapter does not expose Zenodo OAI-PMH"),
		support(CapabilityRecordCreate, SupportSupported, "creates OSF projects or components with explicit kinds"),
		support(CapabilityRecordDelete, SupportPartial, "OSF deletion and withdrawal semantics depend on entity kind"),
		support(CapabilityRecordGet, SupportSupported, "provider-qualified OSF GUID and entity kind are retained"),
		support(CapabilityRecordSearch, SupportSupported, "OSF query semantics and pagination remain native"),
		support(CapabilityRecordUpdate, SupportSupported, "mutable fields depend on entity kind and permissions"),
		support(CapabilityVersionCreate, SupportUnsupported, "OSF projects do not share Zenodo record version semantics"),
	}}
}

// ZenodoContract reports behavior evidenced by the pinned Zenodo snapshot.
func ZenodoContract() Contract {
	return Contract{Provider: ProviderZenodo, ModelVersion: 1, Capabilities: []CapabilitySupport{
		support(CapabilityFileDelete, SupportPartial, "files can be removed only while the deposition is unpublished"),
		support(CapabilityFileDownload, SupportSupported, "access follows the record file policy and native links"),
		support(CapabilityFileList, SupportSupported, "draft and published records expose different native shapes"),
		support(CapabilityFileUpload, SupportPartial, "uploads target a draft bucket and enforce record limits"),
		support(CapabilityPublish, SupportPartial, "publication is irreversible and requires deposit:actions"),
		support(CapabilityMetadataUpdate, SupportPartial, "published metadata changes use provider lifecycle actions"),
		support(CapabilityOAIHarvest, SupportSupported, "OAI-PMH has separate resumption and rate-limit semantics"),
		support(CapabilityRecordCreate, SupportPartial, "creation produces an unpublished deposition"),
		support(CapabilityRecordDelete, SupportPartial, "only unpublished depositions can be deleted"),
		support(CapabilityRecordGet, SupportSupported, "published record and deposition identities remain distinct"),
		support(CapabilityRecordSearch, SupportSupported, "published record search is public"),
		support(CapabilityRecordUpdate, SupportPartial, "updates depend on draft or edit state"),
		support(CapabilityVersionCreate, SupportPartial, "new versions require the latest published version identity"),
	}}
}
