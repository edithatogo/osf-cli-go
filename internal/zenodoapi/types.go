package zenodoapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/edithatogo/osf-cli-go/internal/repository"
)

// Creator is a Zenodo creator entry.
type Creator struct {
	Name        string `json:"name"`
	Affiliation string `json:"affiliation,omitempty"`
	ORCID       string `json:"orcid,omitempty"`
}

// License preserves Zenodo's license identifier.
type License struct {
	ID string `json:"id,omitempty"`
}

// RecordMetadata contains stable discovery fields while raw record JSON remains available.
type RecordMetadata struct {
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Creators    []Creator `json:"creators,omitempty"`
	Keywords    []string  `json:"keywords,omitempty"`
	AccessRight string    `json:"access_right,omitempty"`
	License     License   `json:"license,omitempty"`
}

// File models both current record-file entries and the legacy file array shape.
type File struct {
	ID       string            `json:"id,omitempty"`
	Key      string            `json:"key"`
	Size     int64             `json:"size,omitempty"`
	Checksum string            `json:"checksum,omitempty"`
	Links    map[string]string `json:"links,omitempty"`
}

// ContentURL returns the preferred read URL without inferring authorization.
func (file File) ContentURL() string {
	for _, relation := range []string{"content", "download", "self"} {
		if value := strings.TrimSpace(file.Links[relation]); value != "" {
			return value
		}
	}
	return ""
}

// Record models a published Zenodo record and preserves its original JSON.
type Record struct {
	ID           string            `json:"id"`
	ConceptRecID string            `json:"conceptrecid,omitempty"`
	DOI          string            `json:"doi,omitempty"`
	ConceptDOI   string            `json:"conceptdoi,omitempty"`
	Created      string            `json:"created,omitempty"`
	Updated      string            `json:"updated,omitempty"`
	Metadata     RecordMetadata    `json:"metadata"`
	Files        []File            `json:"files,omitempty"`
	Links        map[string]string `json:"links,omitempty"`
	raw          []byte
}

func decodeStringLike(data []byte) (string, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return "", nil
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return "", err
	}
	switch typed := value.(type) {
	case nil:
		return "", nil
	case string:
		return typed, nil
	case json.Number:
		return typed.String(), nil
	default:
		return "", fmt.Errorf("expected string or number, got %T", value)
	}
}

func decodeStringMap(data map[string]json.RawMessage) (map[string]string, error) {
	values := make(map[string]string, len(data))
	for key, raw := range data {
		value, err := decodeStringLike(raw)
		if err != nil {
			// Zenodo exposes nested link groups such as thumbnails. Keep the
			// scalar links used by this client and retain the complete response
			// in Record.NativeJSON for callers that need nested provider data.
			continue
		}
		values[key] = value
	}
	return values, nil
}

// NativeJSON returns an independent copy of the original provider response.
func (record Record) NativeJSON() []byte { return append([]byte(nil), record.raw...) }

// UnmarshalJSON accepts the current entries map and the legacy files array.
func (record *Record) UnmarshalJSON(data []byte) error {
	var envelope struct {
		ID           json.RawMessage            `json:"id"`
		ConceptRecID json.RawMessage            `json:"conceptrecid"`
		DOI          string                     `json:"doi"`
		ConceptDOI   string                     `json:"conceptdoi"`
		Created      string                     `json:"created"`
		Updated      string                     `json:"updated"`
		Metadata     RecordMetadata             `json:"metadata"`
		Files        json.RawMessage            `json:"files"`
		Links        map[string]json.RawMessage `json:"links"`
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&envelope); err != nil {
		return err
	}
	id, err := decodeStringLike(envelope.ID)
	if err != nil {
		return fmt.Errorf("decode id: %w", err)
	}
	conceptRecID, err := decodeStringLike(envelope.ConceptRecID)
	if err != nil {
		return fmt.Errorf("decode conceptrecid: %w", err)
	}
	links, err := decodeStringMap(envelope.Links)
	if err != nil {
		return fmt.Errorf("decode links: %w", err)
	}
	files, err := decodeFiles(envelope.Files)
	if err != nil {
		return fmt.Errorf("decode files: %w", err)
	}
	*record = Record{
		ID: id, ConceptRecID: conceptRecID, DOI: envelope.DOI,
		ConceptDOI: envelope.ConceptDOI, Created: envelope.Created, Updated: envelope.Updated,
		Metadata: envelope.Metadata, Files: files, Links: links,
		raw: append([]byte(nil), data...),
	}
	return nil
}

func decodeFiles(data []byte) ([]File, error) {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.Equal(data, []byte("null")) {
		return nil, nil
	}
	if data[0] == '[' {
		var files []File
		if err := json.Unmarshal(data, &files); err != nil {
			return nil, err
		}
		return files, nil
	}
	var collection struct {
		Order   []string        `json:"order"`
		Entries map[string]File `json:"entries"`
	}
	if err := json.Unmarshal(data, &collection); err != nil {
		return nil, err
	}
	files := make([]File, 0, len(collection.Entries))
	seen := make(map[string]bool, len(collection.Entries))
	for _, key := range collection.Order {
		if file, ok := collection.Entries[key]; ok {
			if file.Key == "" {
				file.Key = key
			}
			files = append(files, file)
			seen[key] = true
		}
	}
	remaining := make([]string, 0, len(collection.Entries)-len(seen))
	for key := range collection.Entries {
		if !seen[key] {
			remaining = append(remaining, key)
		}
	}
	sort.Strings(remaining)
	for _, key := range remaining {
		file := collection.Entries[key]
		if file.Key == "" {
			file.Key = key
		}
		files = append(files, file)
	}
	return files, nil
}

// Envelope converts a record without dropping native fields or identifiers.
func (record Record) Envelope() (repository.RecordEnvelope, error) {
	if strings.TrimSpace(record.ID) == "" || len(record.raw) == 0 {
		return repository.RecordEnvelope{}, errors.New("zenodo record requires id and native JSON")
	}
	metadata, err := repository.NewNativeMetadata("application/vnd.zenodo.record+json", record.raw)
	if err != nil {
		return repository.RecordEnvelope{}, err
	}
	state := repository.LifecyclePublished
	nativeState := "published"
	if record.Metadata.AccessRight != "" {
		nativeState += ":" + record.Metadata.AccessRight
	}
	if record.Metadata.AccessRight == "embargoed" {
		state = repository.LifecycleEmbargoed
	}
	links := make([]repository.Link, 0, len(record.Links))
	relations := make([]string, 0, len(record.Links))
	for relation := range record.Links {
		relations = append(relations, relation)
	}
	sort.Strings(relations)
	for _, relation := range relations {
		links = append(links, repository.Link{Relation: relation, URL: record.Links[relation]})
	}
	checksums := make([]repository.Checksum, 0, len(record.Files))
	for _, file := range record.Files {
		algorithm, value, ok := strings.Cut(file.Checksum, ":")
		if ok && algorithm != "" && value != "" {
			checksums = append(checksums, repository.Checksum{Algorithm: algorithm, Value: value})
		}
	}
	return repository.RecordEnvelope{
		Identity:       repository.QualifiedID{Provider: repository.ProviderZenodo, Kind: repository.KindRecord, NativeID: record.ID},
		Title:          record.Metadata.Title,
		Lifecycle:      repository.Lifecycle{Common: state, Native: nativeState},
		NativeMetadata: metadata,
		Links:          links,
		Permissions:    repository.Permissions{Read: repository.PermissionAllowed, Write: repository.PermissionUnknown, Delete: repository.PermissionUnknown, Publish: repository.PermissionUnsupported},
		Version:        repository.VersionIdentity{NativeVersionID: record.ID, DOI: record.DOI, ConceptDOI: record.ConceptDOI},
		Checksums:      checksums,
	}, nil
}

// RateLimit captures the reviewed Zenodo response headers.
type RateLimit struct {
	Limit     int64
	Remaining int64
	ResetUnix int64
}

// APIError preserves safe Zenodo HTTP failure details.
type APIError struct {
	StatusCode int
	Method     string
	Path       string
	Message    string
	RetryAfter time.Duration
	RateLimit  RateLimit
}

func (e *APIError) Error() string {
	if e == nil {
		return "<nil>"
	}
	message := fmt.Sprintf("zenodo api error: %s %s returned %d", e.Method, e.Path, e.StatusCode)
	if e.Message != "" {
		message += ": " + e.Message
	}
	return message
}
