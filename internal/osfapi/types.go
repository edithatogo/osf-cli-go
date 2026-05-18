package osfapi

import (
	"fmt"
	"time"
)

type document[T any] struct {
	Data  T     `json:"data"`
	Links Links `json:"links"`
}

// Links carries the JSON:API link set used by OSF responses.
type Links struct {
	Self     string `json:"self,omitempty"`
	Next     string `json:"next,omitempty"`
	Prev     string `json:"prev,omitempty"`
	Related  string `json:"related,omitempty"`
	Download string `json:"download,omitempty"`
}

// APIError preserves OSF status and error detail fields.
type APIError struct {
	StatusCode int
	Method     string
	Path       string
	Title      string
	Detail     string
}

// Error returns a human-readable summary of the OSF API error.
func (e *APIError) Error() string {
	if e == nil {
		return "<nil>"
	}
	base := fmt.Sprintf("osf api error: %s %s returned %d", e.Method, e.Path, e.StatusCode)
	if e.Title != "" {
		base += ": " + e.Title
	}
	if e.Detail != "" {
		if e.Title != "" {
			base += " - " + e.Detail
		} else {
			base += ": " + e.Detail
		}
	}
	return base
}

// User models the OSF current user response.
type User struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"`
	Attributes UserAttributes `json:"attributes"`
	Links      Links          `json:"links"`
}

// UserAttributes carries profile fields returned for an OSF user.
type UserAttributes struct {
	FullName   string `json:"full_name"`
	GivenName  string `json:"given_name,omitempty"`
	FamilyName string `json:"family_name,omitempty"`
}

// Node models OSF projects and components.
type Node struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"`
	Attributes NodeAttributes `json:"attributes"`
	Links      Links          `json:"links"`
}

// NodeAttributes carries display and category fields for OSF nodes.
type NodeAttributes struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
}

// Contributor models a node contributor entry.
type Contributor struct {
	ID         string                `json:"id"`
	Type       string                `json:"type"`
	Attributes ContributorAttributes `json:"attributes"`
	Links      Links                 `json:"links"`
}

// ContributorAttributes carries contributor display and permission metadata.
type ContributorAttributes struct {
	FullName      string `json:"full_name"`
	Bibliographic bool   `json:"bibliographic"`
	Permission    string `json:"permission,omitempty"`
}

// StorageFile models OSF Storage file and folder entries.
type StorageFile struct {
	ID         string                `json:"id"`
	Type       string                `json:"type"`
	Attributes StorageFileAttributes `json:"attributes"`
	Links      Links                 `json:"links"`
}

// StorageFileAttributes carries OSF Storage file metadata.
type StorageFileAttributes struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
	Size int64  `json:"size,omitempty"`
}

// DownloadURL returns the file download URL when OSF provides one.
func (f StorageFile) DownloadURL() string {
	return f.Links.Download
}

// FileVersion represents a version of an OSF storage file.
type FileVersion struct {
	ID         string                `json:"id"`
	Type       string                `json:"type"`
	Attributes FileVersionAttributes `json:"attributes"`
	Links      Links                 `json:"links"`
}

// FileVersionAttributes carries size and timestamp metadata for a file version.
type FileVersionAttributes struct {
	Size         int64     `json:"size,omitempty"`
	DateCreated  time.Time `json:"date_created,omitempty"`
	DateModified time.Time `json:"date_modified,omitempty"`
}

// SearchResult models one OSF search result across supported resource types.
type SearchResult struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
	URL         string `json:"url,omitempty"`
}

// RegistrationRequest describes the draft registration fields submitted for a node.
type RegistrationRequest struct {
	SchemaID    string `json:"schemaId"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}
