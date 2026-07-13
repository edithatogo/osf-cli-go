package osfapi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var doiPattern = regexp.MustCompile(`^10\.\d{4,9}/\S+$`)

// DOIResolution describes a DOI that resolved to an OSF web resource.
type DOIResolution struct {
	DOI         string `json:"doi"`
	ResolvedURL string `json:"resolved_url"`
}

// ResolveDOI resolves an OSF DOI or DOI URL and rejects non-OSF destinations.
// It follows normal HTTP redirects but never treats an arbitrary DOI target as
// an OSF resource.
func ResolveDOI(ctx context.Context, identifier string) (DOIResolution, error) {
	return resolveDOIWithHTTPClient(ctx, identifier, &http.Client{Timeout: 30 * time.Second})
}

func resolveDOIWithHTTPClient(ctx context.Context, identifier string, client *http.Client) (DOIResolution, error) {
	doi, err := parseDOI(identifier)
	if err != nil {
		return DOIResolution{}, err
	}
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}

	requestURL := "https://doi.org/" + doi
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, requestURL, nil)
	if err != nil {
		return DOIResolution{}, fmt.Errorf("create DOI request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return DOIResolution{}, fmt.Errorf("resolve DOI %q: %w", doi, err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode == http.StatusMethodNotAllowed || resp.StatusCode == http.StatusForbidden {
		resp, err = doDOIGet(ctx, requestURL, client)
		if err != nil {
			return DOIResolution{}, err
		}
		_ = resp.Body.Close()
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return DOIResolution{}, fmt.Errorf("DOI %q returned HTTP %d", doi, resp.StatusCode)
	}
	if resp.Request == nil || resp.Request.URL == nil {
		return DOIResolution{}, fmt.Errorf("DOI %q returned no final URL", doi)
	}
	if !isOSFHost(resp.Request.URL.Hostname()) {
		return DOIResolution{}, fmt.Errorf("DOI %q resolved to non-OSF host %q", doi, resp.Request.URL.Hostname())
	}
	return DOIResolution{DOI: doi, ResolvedURL: resp.Request.URL.String()}, nil
}

func doDOIGet(ctx context.Context, requestURL string, client *http.Client) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create DOI fallback request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("resolve DOI with GET: %w", err)
	}
	return resp, nil
}

func parseDOI(identifier string) (string, error) {
	value := strings.TrimSpace(identifier)
	if strings.HasPrefix(strings.ToLower(value), "doi:") {
		value = strings.TrimSpace(value[4:])
	}
	if parsed, err := url.Parse(value); err == nil && parsed.Scheme != "" && parsed.Host != "" {
		if !strings.EqualFold(parsed.Hostname(), "doi.org") && !strings.EqualFold(parsed.Hostname(), "www.doi.org") {
			return "", fmt.Errorf("identifier URL host %q is not doi.org", parsed.Hostname())
		}
		value = strings.TrimPrefix(parsed.Path, "/")
	}
	if !doiPattern.MatchString(value) || strings.ContainsAny(value, "\r\n") {
		return "", fmt.Errorf("%q is not a valid DOI", identifier)
	}
	return value, nil
}

func isOSFHost(host string) bool {
	host = strings.ToLower(strings.TrimSuffix(host, "."))
	return host == "osf.io" || strings.HasSuffix(host, ".osf.io")
}
