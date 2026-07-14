package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func TestDryRunAndEvidenceRedaction(t *testing.T) {
	t.Parallel()
	report, err := runValidation(t.Context(), validationEnv{token: "secret-token"}, false, 64)
	if err != nil || report.Mode != "dry-run" {
		t.Fatalf("runValidation = %+v err=%v", report, err)
	}
	path := filepath.Join(t.TempDir(), "evidence.md")
	if err := writeEvidence(path, report, "secret-token"); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	text := string(data)
	if strings.Contains(text, "secret-token") || !strings.Contains(text, "ZENODO_TOKEN: set") || !strings.Contains(text, "ZENODO_SANDBOX_VALIDATION=1") {
		t.Fatalf("evidence = %s", text)
	}
}

func TestLiveValidationVerifiesResumeAndCleanup(t *testing.T) {
	t.Parallel()
	const token = "sandbox-token"
	var server *httptest.Server
	var stored []byte
	var deleted atomic.Bool
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+token {
			t.Errorf("authorization = %q", r.Header.Get("Authorization"))
		}
		switch {
		case r.Method == http.MethodPost:
			_, _ = fmt.Fprintf(w, `{"id":123,"links":{"bucket":%q}}`, server.URL+"/api/bucket")
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/files"):
			_, _ = io.WriteString(w, `[]`)
		case r.Method == http.MethodPut:
			stored, _ = io.ReadAll(r.Body)
			digest := md5.Sum(stored)
			_, _ = fmt.Fprintf(w, `{"id":"file","filename":"validation.bin","filesize":%d,"checksum":"md5:%s","links":{"download":%q}}`, len(stored), hex.EncodeToString(digest[:]), server.URL+"/api/file")
		case r.Method == http.MethodGet && r.URL.Path == "/api/file":
			offset := 0
			if r.Header.Get("Range") != "" {
				_, _ = fmt.Sscanf(r.Header.Get("Range"), "bytes=%d-", &offset)
				w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", offset, len(stored)-1, len(stored)))
				w.WriteHeader(http.StatusPartialContent)
			}
			_, _ = w.Write(stored[offset:])
		case r.Method == http.MethodDelete:
			deleted.Store(true)
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "unexpected", http.StatusNotFound)
		}
	}))
	defer server.Close()
	report, err := runValidation(t.Context(), validationEnv{enabled: true, token: token, baseURL: server.URL + "/api/"}, true, 128)
	if err != nil {
		t.Fatalf("runValidation: %v report=%+v", err, report)
	}
	if report.Mode != "live" || !deleted.Load() || stepStatus(report, "resumable download") != "passed" || stepStatus(report, "cleanup") != "passed" {
		t.Fatalf("report=%+v deleted=%v", report, deleted.Load())
	}
}

func TestLiveValidationCleansUpAfterTransferFailure(t *testing.T) {
	t.Parallel()
	var server *httptest.Server
	var deleted atomic.Bool
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			_, _ = fmt.Fprintf(w, `{"id":123,"links":{"bucket":%q}}`, server.URL+"/api/bucket")
		case http.MethodGet:
			_, _ = io.WriteString(w, `[]`)
		case http.MethodPut:
			_, _ = io.Copy(io.Discard, r.Body)
			_, _ = fmt.Fprintf(w, `{"id":"file","filename":"validation.bin","filesize":64,"checksum":"md5:00000000000000000000000000000000","links":{"download":%q}}`, server.URL+"/api/file")
		case http.MethodDelete:
			deleted.Store(true)
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer server.Close()
	report, err := runValidation(t.Context(), validationEnv{enabled: true, token: "secret", baseURL: server.URL + "/api/"}, true, 64)
	if err == nil || !deleted.Load() || stepStatus(report, "cleanup") != "passed" {
		t.Fatalf("report=%+v err=%v deleted=%v", report, err, deleted.Load())
	}
}
