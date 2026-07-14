package zenodooai

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/edithatogo/osf-cli-go/internal/repository"
)

var (
	// ErrResponseTooLarge indicates that an XML response exceeded its memory budget.
	ErrResponseTooLarge = errors.New("zenodo OAI-PMH response exceeds configured size limit")
	// ErrPageLimit indicates that harvesting exceeded its configured page budget.
	ErrPageLimit = errors.New("zenodo OAI-PMH harvest exceeds configured page limit")
	// ErrRecordLimit indicates that harvesting exceeded its configured record budget.
	ErrRecordLimit = errors.New("zenodo OAI-PMH harvest exceeds configured record limit")
	// ErrTokenExpired indicates that a resumption token is no longer valid.
	ErrTokenExpired = errors.New("zenodo OAI-PMH resumption token has expired")
	// ErrCrossOrigin indicates that an OAI endpoint escaped the configured origin.
	ErrCrossOrigin = errors.New("zenodo OAI-PMH endpoint leaves configured origin")
)

// Request selects an OAI-PMH ListRecords page. A resumption token is exclusive
// with all other selectors, as required by the protocol.
type Request struct {
	MetadataPrefix string
	Set            string
	From           time.Time
	Until          time.Time
	Token          ResumptionToken
}

// ResumptionToken carries both the opaque continuation and server expiry data.
type ResumptionToken struct {
	Value            string    `json:"value"`
	ExpiresAt        time.Time `json:"expiresAt,omitempty"`
	Cursor           int       `json:"cursor,omitempty"`
	CompleteListSize int       `json:"completeListSize,omitempty"`
	MetadataPrefix   string    `json:"metadataPrefix,omitempty"`
	Set              string    `json:"set,omitempty"`
}

// Empty reports whether there is no continuation.
func (token ResumptionToken) Empty() bool { return strings.TrimSpace(token.Value) == "" }

// Provenance describes the exact OAI harvest context for a native record.
type Provenance struct {
	BaseURL        string    `json:"baseUrl"`
	ResponseDate   time.Time `json:"responseDate"`
	MetadataPrefix string    `json:"metadataPrefix"`
	Set            string    `json:"set,omitempty"`
	Datestamp      string    `json:"datestamp"`
}

// Header is the OAI-PMH identity and selective-harvesting metadata.
type Header struct {
	Identifier string   `json:"identifier"`
	Datestamp  string   `json:"datestamp"`
	SetSpecs   []string `json:"setSpecs,omitempty"`
	Deleted    bool     `json:"deleted,omitempty"`
}

// Record preserves the native metadata XML and its harvest provenance.
type Record struct {
	Header         Header                     `json:"header"`
	NativeMetadata *repository.NativeMetadata `json:"nativeMetadata,omitempty"`
	AboutXML       []byte                     `json:"aboutXml,omitempty"`
	Provenance     Provenance                 `json:"provenance"`
}

// Page is one deterministic ListRecords response.
type Page struct {
	Records []Record        `json:"records"`
	Next    ResumptionToken `json:"next,omitempty"`
}

// Set is an OAI selective-harvesting set.
type Set struct {
	Spec        string `json:"spec"`
	Name        string `json:"name"`
	Description []byte `json:"descriptionXml,omitempty"`
}

// MetadataFormat is an OAI metadata schema advertised by the repository.
type MetadataFormat struct {
	Prefix       string `json:"prefix"`
	Schema       string `json:"schema"`
	NamespaceURL string `json:"namespaceUrl"`
}

// ProtocolError represents an OAI-PMH error element.
type ProtocolError struct {
	Code    string
	Message string
	Verb    string
}

func (err *ProtocolError) Error() string {
	if err == nil {
		return "<nil>"
	}
	message := fmt.Sprintf("zenodo OAI-PMH %s error %s", err.Verb, err.Code)
	if err.Message != "" {
		message += ": " + err.Message
	}
	return message
}

// Unwrap maps the protocol's invalid continuation to the stable expiry sentinel.
func (err *ProtocolError) Unwrap() error {
	if err != nil && err.Code == "badResumptionToken" {
		return ErrTokenExpired
	}
	return nil
}

// HTTPError is a redacted transport-level OAI failure.
type HTTPError struct {
	StatusCode int
	RetryAfter time.Duration
}

func (err *HTTPError) Error() string {
	if err == nil {
		return "<nil>"
	}
	return fmt.Sprintf("zenodo OAI-PMH request returned HTTP %d", err.StatusCode)
}
