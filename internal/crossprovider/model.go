package crossprovider

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/edithatogo/osf-cli-go/internal/repository"
)

var ErrInvalidRequest = errors.New("invalid cross-provider transfer request")

// Direction names the only two supported provider-copy directions.
type Direction string

const (
	DirectionOSFToZenodo Direction = "osf_to_zenodo"
	DirectionZenodoToOSF Direction = "zenodo_to_osf"
)

// PublishIntent makes destination publication intent explicit during planning.
type PublishIntent string

const (
	PublishDraftOnly PublishIntent = "draft_only"
	PublishAfterCopy PublishIntent = "publish_after_copy"
)

// ConflictPolicy controls destination collisions without implying mirroring.
type ConflictPolicy string

const (
	ConflictFail          ConflictPolicy = "fail"
	ConflictSkipIdentical ConflictPolicy = "skip_identical"
	ConflictReplaceDraft  ConflictPolicy = "replace_draft"
)

// AccessKind preserves both OSF visibility and Zenodo access vocabularies.
type AccessKind string

const (
	AccessPublic     AccessKind = "public"
	AccessPrivate    AccessKind = "private"
	AccessOpen       AccessKind = "open"
	AccessEmbargoed  AccessKind = "embargoed"
	AccessRestricted AccessKind = "restricted"
	AccessClosed     AccessKind = "closed"
)

// AccessPolicy records access semantics that must not be inferred silently.
type AccessPolicy struct {
	Kind         AccessKind `json:"kind"`
	EmbargoUntil *time.Time `json:"embargoUntil,omitempty"`
	Conditions   string     `json:"conditions,omitempty"`
}

// Creator is a normalized creator identity retained in provenance.
type Creator struct {
	Name string `json:"name"`
}

// Identifier preserves a typed persistent or provider-native identifier.
type Identifier struct {
	Scheme string `json:"scheme"`
	Value  string `json:"value"`
}

// Metadata is the common mapping vocabulary. Provider-native metadata remains separate.
type Metadata struct {
	Title       string       `json:"title"`
	Description string       `json:"description"`
	UploadType  string       `json:"uploadType"`
	Creators    []Creator    `json:"creators"`
	Keywords    []string     `json:"keywords,omitempty"`
	Access      AccessPolicy `json:"access"`
	License     string       `json:"license,omitempty"`
	Identifiers []Identifier `json:"identifiers,omitempty"`
	Version     string       `json:"version,omitempty"`
}

// File is a logical source file with integrity metadata.
type File struct {
	Path      string `json:"path"`
	Size      int64  `json:"size"`
	Checksum  string `json:"checksum"`
	MediaType string `json:"mediaType,omitempty"`
}

// Snapshot is an immutable transfer input captured by a concrete provider adapter.
type Snapshot struct {
	Identity       repository.QualifiedID    `json:"identity"`
	Metadata       Metadata                  `json:"metadata"`
	Files          []File                    `json:"files"`
	NativeMetadata repository.NativeMetadata `json:"nativeMetadata"`
}

// Destination identifies either a new destination record or a concrete draft.
type Destination struct {
	Provider  repository.Provider `json:"provider"`
	NativeID  string              `json:"nativeId,omitempty"`
	CreateNew bool                `json:"createNew"`
}

// Request is the complete caller intent required before planning.
type Request struct {
	Direction     Direction      `json:"direction"`
	Source        Snapshot       `json:"source"`
	Destination   Destination    `json:"destination"`
	Authorized    bool           `json:"authorized"`
	PublishIntent PublishIntent  `json:"publishIntent"`
	Conflict      ConflictPolicy `json:"conflict"`
	TargetAccess  *AccessPolicy  `json:"targetAccess,omitempty"`
	TargetLicense string         `json:"targetLicense,omitempty"`
}

// Disposition states how one source semantic is handled.
type Disposition string

const (
	DispositionExact           Disposition = "exact"
	DispositionTransformed     Disposition = "transformed"
	DispositionPreservedNative Disposition = "preserved_native"
	DispositionBlocked         Disposition = "blocked"
)

// FieldMapping makes semantic preservation or loss explicit.
type FieldMapping struct {
	SourceField string      `json:"sourceField"`
	TargetField string      `json:"targetField,omitempty"`
	Disposition Disposition `json:"disposition"`
	Reason      string      `json:"reason,omitempty"`
}

// Provenance binds a report to the source and its native metadata digest.
type Provenance struct {
	SourceIdentity       repository.QualifiedID `json:"sourceIdentity"`
	DestinationProvider  repository.Provider    `json:"destinationProvider"`
	CapturedAt           time.Time              `json:"capturedAt"`
	NativeMetadataSHA256 string                 `json:"nativeMetadataSha256"`
	Transformations      []FieldMapping         `json:"transformations"`
}

// Report is a dry-run mapping decision. It never performs a provider write.
type Report struct {
	Direction      Direction      `json:"direction"`
	Destination    Destination    `json:"destination"`
	PublishIntent  PublishIntent  `json:"publishIntent"`
	Conflict       ConflictPolicy `json:"conflict"`
	Target         Metadata       `json:"target"`
	Files          []File         `json:"files"`
	Fields         []FieldMapping `json:"fields"`
	NativeFields   []FieldMapping `json:"nativeFields,omitempty"`
	Blockers       []string       `json:"blockers,omitempty"`
	Executable     bool           `json:"executable"`
	IdempotencyKey string         `json:"idempotencyKey"`
	Provenance     Provenance     `json:"provenance"`
}

var mappedFields = []string{"title", "description", "upload_type", "creators", "keywords", "access", "license", "embargo", "identifiers", "version", "native_metadata", "files"}

// BuildMapping validates explicit intent and returns a deterministic dry-run report.
func BuildMapping(request Request, capturedAt time.Time) (Report, error) {
	if err := request.validate(capturedAt); err != nil {
		return Report{}, err
	}
	target := request.Source.Metadata.clone()
	if strings.TrimSpace(request.TargetLicense) != "" {
		target.License = strings.TrimSpace(request.TargetLicense)
	}
	fields := exactMappings()
	var blockers []string
	if request.Direction == DirectionOSFToZenodo {
		mapOSFToZenodo(request, &target, fields, &blockers)
	} else {
		mapZenodoToOSF(request, &target, fields)
	}
	digest := sha256.Sum256(request.Source.NativeMetadata.Bytes())
	transformations := make([]FieldMapping, 0, len(fields))
	for _, mapping := range fields {
		if mapping.Disposition != DispositionExact {
			transformations = append(transformations, mapping)
		}
	}
	nativeFields := nativeFieldMappings(request.Source.NativeMetadata)
	transformations = append(transformations, nativeFields...)
	key, err := idempotencyKey(request, target)
	if err != nil {
		return Report{}, err
	}
	return Report{
		Direction: request.Direction, Destination: request.Destination,
		PublishIntent: request.PublishIntent, Conflict: request.Conflict,
		Target: target, Files: sortedFiles(request.Source.Files), Fields: fields, NativeFields: nativeFields,
		Blockers: blockers, Executable: len(blockers) == 0, IdempotencyKey: key,
		Provenance: Provenance{
			SourceIdentity: request.Source.Identity, DestinationProvider: request.Destination.Provider,
			CapturedAt: capturedAt, NativeMetadataSHA256: "sha256:" + hex.EncodeToString(digest[:]),
			Transformations: transformations,
		},
	}, nil
}

func (request Request) validate(capturedAt time.Time) error {
	if err := request.Source.Identity.Validate(); err != nil {
		return fmt.Errorf("%w: source identity: %v", ErrInvalidRequest, err)
	}
	wantDirection := DirectionOSFToZenodo
	wantDestination := repository.ProviderZenodo
	if request.Source.Identity.Provider == repository.ProviderZenodo {
		wantDirection = DirectionZenodoToOSF
		wantDestination = repository.ProviderOSF
	}
	if request.Direction != wantDirection {
		return fmt.Errorf("%w: direction %q does not match source provider %q", ErrInvalidRequest, request.Direction, request.Source.Identity.Provider)
	}
	if request.Destination.Provider != wantDestination {
		return fmt.Errorf("%w: destination provider must be %q", ErrInvalidRequest, wantDestination)
	}
	if request.Destination.CreateNew == (strings.TrimSpace(request.Destination.NativeID) != "") {
		return fmt.Errorf("%w: destination requires exactly one of createNew or nativeId", ErrInvalidRequest)
	}
	if !request.Authorized {
		return fmt.Errorf("%w: explicit authorization is required", ErrInvalidRequest)
	}
	if request.PublishIntent != PublishDraftOnly && request.PublishIntent != PublishAfterCopy {
		return fmt.Errorf("%w: publish intent must be draft_only or publish_after_copy", ErrInvalidRequest)
	}
	if request.Conflict != ConflictFail && request.Conflict != ConflictSkipIdentical && request.Conflict != ConflictReplaceDraft {
		return fmt.Errorf("%w: conflict policy is invalid", ErrInvalidRequest)
	}
	if err := request.Source.validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidRequest, err)
	}
	if request.TargetAccess != nil {
		if err := request.TargetAccess.validateFor(request.Destination.Provider); err != nil {
			return fmt.Errorf("%w: target access: %v", ErrInvalidRequest, err)
		}
		if err := request.TargetAccess.validateTarget(capturedAt); err != nil {
			return fmt.Errorf("%w: target access: %v", ErrInvalidRequest, err)
		}
	}
	return nil
}

func (snapshot Snapshot) validate() error {
	if len(snapshot.NativeMetadata.Bytes()) == 0 || strings.TrimSpace(snapshot.NativeMetadata.MediaType()) == "" {
		return errors.New("source native metadata is required")
	}
	if strings.TrimSpace(snapshot.Metadata.Title) == "" {
		return errors.New("source title is required")
	}
	seen := make(map[string]bool, len(snapshot.Files))
	for _, file := range snapshot.Files {
		clean := path.Clean(strings.TrimSpace(file.Path))
		if clean == "." || strings.HasPrefix(clean, "../") || path.IsAbs(clean) || file.Size < 0 || strings.TrimSpace(file.Checksum) == "" {
			return fmt.Errorf("source file %q has invalid path, size, or checksum", file.Path)
		}
		if seen[clean] {
			return fmt.Errorf("source file path %q is duplicated", clean)
		}
		seen[clean] = true
	}
	return snapshot.Metadata.Access.validateFor(snapshot.Identity.Provider)
}

func (access AccessPolicy) validateFor(provider repository.Provider) error {
	valid := false
	switch provider {
	case repository.ProviderOSF:
		valid = access.Kind == AccessPublic || access.Kind == AccessPrivate
	case repository.ProviderZenodo:
		valid = access.Kind == AccessOpen || access.Kind == AccessEmbargoed || access.Kind == AccessRestricted || access.Kind == AccessClosed
	}
	if !valid {
		return fmt.Errorf("access %q is invalid for provider %q", access.Kind, provider)
	}
	if access.Kind == AccessEmbargoed && access.EmbargoUntil == nil {
		return errors.New("embargoed access requires an embargo date")
	}
	if access.Kind == AccessRestricted && strings.TrimSpace(access.Conditions) == "" {
		return errors.New("restricted access requires conditions")
	}
	return nil
}

func mapOSFToZenodo(request Request, target *Metadata, fields []FieldMapping, blockers *[]string) {
	if request.TargetAccess != nil {
		target.Access = cloneAccess(*request.TargetAccess)
		setMapping(fields, "access", DispositionTransformed, "caller selected an explicit Zenodo access policy")
	} else if request.Source.Metadata.Access.Kind == AccessPublic {
		target.Access = AccessPolicy{Kind: AccessOpen}
		setMapping(fields, "access", DispositionTransformed, "public OSF visibility maps to open Zenodo access")
	} else {
		setMapping(fields, "access", DispositionBlocked, "private OSF visibility has no safe implicit Zenodo access mapping")
		*blockers = append(*blockers, "select an explicit Zenodo access policy for the private OSF source")
	}
	if (target.Access.Kind == AccessOpen || target.Access.Kind == AccessEmbargoed) && strings.TrimSpace(target.License) == "" {
		setMapping(fields, "license", DispositionBlocked, "open or embargoed Zenodo access requires an explicit license")
		*blockers = append(*blockers, "supply a license for open or embargoed Zenodo access")
	}
	setMapping(fields, "identifiers", DispositionTransformed, "source identifiers become related identifiers and never replace destination identity")
	setMapping(fields, "version", DispositionTransformed, "source version is descriptive metadata, not a Zenodo version identity")
	setMapping(fields, "native_metadata", DispositionPreservedNative, "OSF-native metadata is attached to provenance, not flattened into Zenodo fields")
}

func mapZenodoToOSF(request Request, target *Metadata, fields []FieldMapping) {
	if request.TargetAccess != nil {
		target.Access = cloneAccess(*request.TargetAccess)
	} else if request.Source.Metadata.Access.Kind == AccessOpen {
		target.Access = AccessPolicy{Kind: AccessPublic}
	} else {
		target.Access = AccessPolicy{Kind: AccessPrivate}
	}
	setMapping(fields, "access", DispositionTransformed, "Zenodo access maps only to OSF public or private visibility")
	for _, field := range []string{"license", "embargo", "identifiers", "version", "native_metadata"} {
		setMapping(fields, field, DispositionPreservedNative, "OSF has no equivalent stable field in the reviewed cross-provider contract")
	}
}

func exactMappings() []FieldMapping {
	result := make([]FieldMapping, 0, len(mappedFields))
	for _, field := range mappedFields {
		result = append(result, FieldMapping{SourceField: field, TargetField: field, Disposition: DispositionExact})
	}
	return result
}

func setMapping(fields []FieldMapping, source string, disposition Disposition, reason string) {
	index := slices.IndexFunc(fields, func(mapping FieldMapping) bool { return mapping.SourceField == source })
	if index >= 0 {
		fields[index].Disposition = disposition
		fields[index].Reason = reason
	}
}

func idempotencyKey(request Request, target Metadata) (string, error) {
	request.Source.Files = sortedFiles(request.Source.Files)
	canonical := struct {
		Direction     Direction      `json:"direction"`
		Source        Snapshot       `json:"source"`
		Destination   Destination    `json:"destination"`
		PublishIntent PublishIntent  `json:"publishIntent"`
		Conflict      ConflictPolicy `json:"conflict"`
		Target        Metadata       `json:"target"`
	}{request.Direction, request.Source, request.Destination, request.PublishIntent, request.Conflict, target}
	encoded, err := json.Marshal(canonical)
	if err != nil {
		return "", fmt.Errorf("encode cross-provider idempotency input: %w", err)
	}
	digest := sha256.Sum256(encoded)
	return "xfer-v1-" + hex.EncodeToString(digest[:]), nil
}

func sortedFiles(files []File) []File {
	result := append([]File(nil), files...)
	slices.SortFunc(result, func(left, right File) int { return strings.Compare(left.Path, right.Path) })
	return result
}

func (metadata Metadata) clone() Metadata {
	result := metadata
	result.Creators = append([]Creator(nil), metadata.Creators...)
	result.Keywords = append([]string(nil), metadata.Keywords...)
	result.Identifiers = append([]Identifier(nil), metadata.Identifiers...)
	if metadata.Access.EmbargoUntil != nil {
		value := *metadata.Access.EmbargoUntil
		result.Access.EmbargoUntil = &value
	}
	return result
}

func cloneAccess(access AccessPolicy) AccessPolicy {
	result := access
	if access.EmbargoUntil != nil {
		value := *access.EmbargoUntil
		result.EmbargoUntil = &value
	}
	return result
}

func (access AccessPolicy) validateTarget(capturedAt time.Time) error {
	switch access.Kind {
	case AccessPublic, AccessPrivate, AccessOpen, AccessClosed:
		if access.EmbargoUntil != nil || strings.TrimSpace(access.Conditions) != "" {
			return fmt.Errorf("access %q cannot set embargo or conditions", access.Kind)
		}
	case AccessEmbargoed:
		if access.EmbargoUntil == nil || !access.EmbargoUntil.After(capturedAt) || strings.TrimSpace(access.Conditions) != "" {
			return errors.New("embargoed target access requires a future date and no restricted conditions")
		}
	case AccessRestricted:
		if strings.TrimSpace(access.Conditions) == "" || access.EmbargoUntil != nil {
			return errors.New("restricted target access requires conditions and no embargo date")
		}
	}
	return nil
}

func nativeFieldMappings(metadata repository.NativeMetadata) []FieldMapping {
	var object map[string]json.RawMessage
	if json.Unmarshal(metadata.Bytes(), &object) != nil {
		return []FieldMapping{{SourceField: "<opaque>", Disposition: DispositionPreservedNative, Reason: "non-JSON provider metadata remains in the lossless provenance envelope"}}
	}
	names := make([]string, 0, len(object))
	for name := range object {
		names = append(names, name)
	}
	slices.Sort(names)
	result := make([]FieldMapping, 0, len(names))
	for _, name := range names {
		result = append(result, FieldMapping{SourceField: name, Disposition: DispositionPreservedNative, Reason: "provider-native field remains in the lossless provenance envelope"})
	}
	return result
}
