package osfapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestResolveDOIAcceptsFormsAndOSFDestination(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodHead || req.URL.Host != "doi.org" {
			return nil, fmt.Errorf("request = %s %s, want HEAD doi.org", req.Method, req.URL)
		}
		resolved, _ := url.Parse("https://osf.io/abc12/")
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("")), Request: &http.Request{Method: req.Method, URL: resolved}}, nil
	})}

	for _, input := range []string{"10.1234/example", "doi:10.1234/example", "https://doi.org/10.1234/example"} {
		got, err := resolveDOIWithHTTPClient(context.Background(), input, client)
		if err != nil {
			t.Fatalf("resolveDOI(%q) error: %v", input, err)
		}
		if got.DOI != "10.1234/example" || got.ResolvedURL != "https://osf.io/abc12/" {
			t.Fatalf("resolveDOI(%q) = %#v", input, got)
		}
	}
}

func TestResolveDOIRejectsInvalidAndNonOSFDestination(t *testing.T) {
	for _, input := range []string{"", "not-a-doi", "https://example.org/10.1234/example"} {
		if _, err := resolveDOIWithHTTPClient(context.Background(), input, nil); err == nil {
			t.Fatalf("resolveDOI(%q) returned nil error", input)
		}
	}
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		resolved, _ := url.Parse("https://example.org/dataset")
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("")), Request: &http.Request{Method: req.Method, URL: resolved}}, nil
	})}
	if _, err := resolveDOIWithHTTPClient(context.Background(), "10.1234/example", client); err == nil || !strings.Contains(err.Error(), "non-OSF") {
		t.Fatalf("non-OSF resolution error = %v", err)
	}
}
