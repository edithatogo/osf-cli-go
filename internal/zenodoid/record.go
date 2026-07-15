// Package zenodoid parses provider-qualified and canonical Zenodo identities.
package zenodoid

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/edithatogo/osf-cli-go/internal/repository"
)

// ParseRecord returns a native decimal record ID from a native ID, qualified
// ID, or canonical production/sandbox record URL.
func ParseRecord(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", errors.New("zenodo record id is required")
	}
	if strings.HasPrefix(value, "zenodo:") {
		id, err := repository.ParseQualifiedID(value)
		if err != nil {
			return "", fmt.Errorf("invalid Zenodo qualified id: %w", err)
		}
		if id.Provider != repository.ProviderZenodo || id.Kind != repository.KindRecord {
			return "", errors.New("qualified id must identify a Zenodo record")
		}
		value = id.NativeID
	} else if parsed, err := url.Parse(value); err == nil && parsed.Scheme != "" {
		host := strings.ToLower(parsed.Hostname())
		if parsed.Scheme != "https" || parsed.User != nil || parsed.Port() != "" || parsed.RawQuery != "" || parsed.Fragment != "" || (host != "zenodo.org" && host != "sandbox.zenodo.org") {
			return "", errors.New("zenodo record URL must use a canonical production or sandbox HTTPS origin")
		}
		parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
		if len(parts) != 2 || parts[0] != "records" || strings.TrimSpace(parts[1]) == "" {
			return "", errors.New("zenodo record URL must have /records/<id> form")
		}
		value = parts[1]
	}
	if strings.ContainsAny(value, "/?# \\") {
		return "", errors.New("zenodo record id must not contain paths, spaces, query, or fragment")
	}
	for _, character := range value {
		if character < '0' || character > '9' {
			return "", errors.New("zenodo record id must contain decimal digits only")
		}
	}
	return value, nil
}
