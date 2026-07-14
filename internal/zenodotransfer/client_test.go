package zenodotransfer

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/edithatogo/osf-cli-go/internal/download"
)

func TestNewRequiresDedicatedTokenAndRejectsProduction(t *testing.T) {
	t.Parallel()
	for name, input := range map[string]struct {
		baseURL string
		token   string
	}{
		"missing token": {baseURL: "https://sandbox.zenodo.org/api/"},
		"production":    {baseURL: "https://zenodo.org/api/", token: "secret"},
		"other host":    {baseURL: "https://example.com/api/", token: "secret"},
		"url token":     {baseURL: "https://secret@sandbox.zenodo.org/api/", token: "secret"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := New(input.baseURL, input.token); err == nil {
				t.Fatal("New returned nil error")
			}
		})
	}
}

func TestDraftTransferRetriesVerifiesAndCleansUp(t *testing.T) {
	t.Parallel()
	const token = "sandbox-secret-token"
	content := []byte("verified sandbox bytes")
	digest := md5.Sum(content)
	checksum := "md5:" + hex.EncodeToString(digest[:])
	var server *httptest.Server
	var uploadAttempts atomic.Int32
	var deleted atomic.Bool
	var stored []byte

	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+token {
			t.Errorf("authorization = %q", r.Header.Get("Authorization"))
		}
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/deposit/depositions":
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintf(w, `{"id":123,"links":{"bucket":%q}}`, server.URL+"/api/files/bucket-1")
		case r.Method == http.MethodGet && r.URL.Path == "/api/deposit/depositions/123/files":
			_, _ = io.WriteString(w, `[]`)
		case r.Method == http.MethodPut && r.URL.Path == "/api/files/bucket-1/result.txt":
			attempt := uploadAttempts.Add(1)
			body, _ := io.ReadAll(r.Body)
			if attempt == 1 {
				w.WriteHeader(http.StatusServiceUnavailable)
				_, _ = io.WriteString(w, `{"message":"retry upload"}`)
				return
			}
			stored = append([]byte(nil), body...)
			_, _ = fmt.Fprintf(w, `{"id":"file-1","filename":"result.txt","filesize":%d,"checksum":%q,"links":{"download":%q}}`, len(stored), checksum, server.URL+"/api/files/bucket-1/result.txt")
		case r.Method == http.MethodGet && r.URL.Path == "/api/files/bucket-1/result.txt":
			w.Header().Set("Content-Length", fmt.Sprint(len(stored)))
			_, _ = w.Write(stored)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/deposit/depositions/123":
			deleted.Store(true)
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "unexpected request", http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := New(server.URL+"/api/", token, WithRetryPolicy(2, 0))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	draft, err := client.CreateDraft(t.Context())
	if err != nil {
		t.Fatalf("CreateDraft: %v", err)
	}

	dir := t.TempDir()
	source := filepath.Join(dir, "source.txt")
	if err := os.WriteFile(source, content, 0o600); err != nil {
		t.Fatal(err)
	}
	upload, err := client.UploadFile(t.Context(), draft, source, "result.txt", download.ConflictFail)
	if err != nil {
		t.Fatalf("UploadFile: %v", err)
	}
	if upload.Remote.Checksum != checksum || upload.Bytes != int64(len(content)) || upload.RetryCount != 1 || !upload.Completed {
		t.Fatalf("upload = %+v", upload)
	}

	destination := filepath.Join(dir, "download.txt")
	result, err := client.DownloadFile(t.Context(), upload.Remote, destination, download.ConflictFail)
	if err != nil {
		t.Fatalf("DownloadFile: %v", err)
	}
	got, _ := os.ReadFile(destination)
	if string(got) != string(content) || result.Checksum != checksum || !result.Completed {
		t.Fatalf("download = %+v bytes=%q", result, got)
	}
	if err := client.DeleteDraft(t.Context(), draft.ID); err != nil {
		t.Fatalf("DeleteDraft: %v", err)
	}
	if !deleted.Load() {
		t.Fatal("draft was not deleted")
	}
}

func TestUploadConflictPoliciesAndLimits(t *testing.T) {
	t.Parallel()
	var puts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/files"):
			_, _ = fmt.Fprintf(w, `[{"id":"existing","filename":"result.txt","filesize":3,"checksum":"md5:900150983cd24fb0d6963f7d28e17f72","links":{"download":%q}}]`, "http://"+r.Host+"/api/files/result.txt")
		case r.Method == http.MethodPut:
			puts.Add(1)
			_, _ = io.Copy(io.Discard, r.Body)
			_, _ = fmt.Fprintf(w, `{"id":"existing","filename":"result.txt","filesize":3,"checksum":"md5:900150983cd24fb0d6963f7d28e17f72","links":{"download":%q}}`, "http://"+r.Host+"/api/files/result.txt")
		default:
			http.Error(w, "unexpected", http.StatusNotFound)
		}
	}))
	defer server.Close()
	client, err := New(server.URL+"/api/", "secret", WithMaxFileBytes(3))
	if err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(t.TempDir(), "source")
	if err := os.WriteFile(source, []byte("abc"), 0o600); err != nil {
		t.Fatal(err)
	}
	draft := Draft{ID: "123", BucketURL: server.URL + "/api/files/bucket"}
	if _, err := client.UploadFile(t.Context(), draft, source, "result.txt", download.ConflictFail); err == nil {
		t.Fatal("conflict fail returned nil error")
	}
	skipped, err := client.UploadFile(t.Context(), draft, source, "result.txt", download.ConflictSkip)
	if err != nil || !skipped.Skipped || puts.Load() != 0 {
		t.Fatalf("skip = %+v err=%v puts=%d", skipped, err, puts.Load())
	}
	if _, err := client.UploadFile(t.Context(), draft, source, "result.txt", download.ConflictOverwrite); err != nil || puts.Load() != 1 {
		t.Fatalf("overwrite err=%v puts=%d", err, puts.Load())
	}
	if err := os.WriteFile(source, []byte("abcd"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := client.UploadFile(t.Context(), draft, source, "other.txt", download.ConflictFail); err == nil {
		t.Fatal("oversize upload returned nil error")
	}
}

func TestUploadRejectsRecordLimitAndCorruptAcknowledgement(t *testing.T) {
	t.Parallel()
	var listBody = `[{"id":"one","filename":"one.txt","filesize":1,"checksum":"md5:0cc175b9c0f1b6a831c399e269772661","links":{"download":"PLACEHOLDER"}}]`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = io.WriteString(w, strings.ReplaceAll(listBody, "PLACEHOLDER", "http://"+r.Host+"/api/file"))
		case http.MethodPut:
			_, _ = io.Copy(io.Discard, r.Body)
			_, _ = fmt.Fprintf(w, `{"id":"bad","filename":"result.txt","filesize":3,"checksum":"md5:00000000000000000000000000000000","links":{"download":%q}}`, "http://"+r.Host+"/api/file")
		}
	}))
	defer server.Close()
	source := filepath.Join(t.TempDir(), "source")
	if err := os.WriteFile(source, []byte("abc"), 0o600); err != nil {
		t.Fatal(err)
	}
	draft := Draft{ID: "123", BucketURL: server.URL + "/api/bucket"}
	limited, err := New(server.URL+"/api/", "secret", WithMaxFiles(1))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := limited.UploadFile(t.Context(), draft, source, "result.txt", download.ConflictFail); !errors.Is(err, ErrFileCountLimit) {
		t.Fatalf("record limit error = %v", err)
	}

	listBody = `[]`
	client, err := New(server.URL+"/api/", "secret")
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.UploadFile(t.Context(), draft, source, "result.txt", download.ConflictFail)
	if err == nil || result.Completed || result.Bytes != 0 || result.CheckpointPath == "" {
		t.Fatalf("corrupt upload = %+v err=%v", result, err)
	}
}

func TestDownloadResumesAfterInterruptedResponse(t *testing.T) {
	t.Parallel()
	content := []byte("abcdef")
	digest := md5.Sum(content)
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call := calls.Add(1)
		if call == 1 {
			w.Header().Set("Content-Length", fmt.Sprint(len(content)))
			_, _ = w.Write(content[:3])
			return
		}
		if r.Header.Get("Range") != "bytes=3-" {
			t.Errorf("Range = %q, want bytes=3-", r.Header.Get("Range"))
		}
		w.Header().Set("Content-Range", "bytes 3-5/6")
		w.WriteHeader(http.StatusPartialContent)
		_, _ = w.Write(content[3:])
	}))
	defer server.Close()
	client, err := New(server.URL+"/api/", "secret", WithRetryPolicy(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	destination := filepath.Join(t.TempDir(), "download")
	remote := RemoteFile{ID: "file", Name: "result.txt", Size: int64(len(content)), Checksum: "md5:" + hex.EncodeToString(digest[:]), DownloadURL: server.URL + "/api/file"}
	if _, err := client.DownloadFile(t.Context(), remote, destination, download.ConflictOverwrite); err == nil {
		t.Fatal("interrupted download returned nil error")
	}
	result, err := client.DownloadFile(t.Context(), remote, destination, download.ConflictOverwrite)
	if err != nil || !result.Resumed {
		t.Fatalf("resumed download = %+v err=%v", result, err)
	}
}

func TestDownloadRejectsWrongContentRange(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Range", "bytes 4-5/6")
		w.WriteHeader(http.StatusPartialContent)
		_, _ = io.WriteString(w, "ef")
	}))
	defer server.Close()
	client, err := New(server.URL+"/api/", "secret")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.openDownload(t.Context(), mustURL(t, server.URL+"/api/file"), 3); !errors.Is(err, ErrInvalidContentRange) {
		t.Fatalf("openDownload error = %v", err)
	}
}

func TestCancellationAndErrorsAreTruthfulAndRedacted(t *testing.T) {
	t.Parallel()
	const token = "sandbox-secret-token"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = fmt.Fprintf(w, `{"message":"token %s rejected"}`, token)
	}))
	defer server.Close()
	client, err := New(server.URL+"/api/", token, WithRetryPolicy(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.CreateDraft(t.Context()); err == nil || strings.Contains(err.Error(), token) {
		t.Fatalf("CreateDraft error = %v", err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := client.CreateDraft(ctx); err == nil || !strings.Contains(err.Error(), context.Canceled.Error()) {
		t.Fatalf("cancelled CreateDraft error = %v", err)
	}
}

func TestCrossOriginRedirectDoesNotForwardAuthorization(t *testing.T) {
	t.Parallel()
	var received atomic.Bool
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			received.Store(true)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer target.Close()
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL+"/api/deposit/depositions", http.StatusTemporaryRedirect)
	}))
	defer source.Close()
	client, err := New(source.URL+"/api/", "secret")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.CreateDraft(t.Context()); !errors.Is(err, ErrCrossOrigin) {
		t.Fatalf("CreateDraft error = %v", err)
	}
	if received.Load() {
		t.Fatal("authorization reached cross-origin redirect target")
	}
}

func TestControlResponseLimitAndIdempotentCleanup(t *testing.T) {
	t.Parallel()
	var deletes atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			if deletes.Add(1) == 1 {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = io.WriteString(w, strings.Repeat("x", 33))
	}))
	defer server.Close()
	client, err := New(server.URL+"/api/", "secret", WithMaxResponseBytes(32), WithRetryPolicy(1, 0))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListDraftFiles(t.Context(), "123"); !errors.Is(err, ErrResponseTooLarge) {
		t.Fatalf("ListDraftFiles error = %v", err)
	}
	if err := client.DeleteDraft(t.Context(), "123"); err != nil {
		t.Fatalf("DeleteDraft: %v", err)
	}
	if deletes.Load() != 2 {
		t.Fatalf("delete attempts = %d, want 2", deletes.Load())
	}
}

func TestInvalidConfigurationAndTransferInputs(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()
	for name, option := range map[string]Option{
		"response limit": WithMaxResponseBytes(0),
		"file limit":     WithMaxFileBytes(0),
		"file count":     WithMaxFiles(0),
		"retry count":    WithRetryPolicy(-1, 0),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := New(server.URL+"/api/", "secret", option); err == nil {
				t.Fatal("New returned nil error")
			}
		})
	}
	client, err := New(server.URL+"/api/", "secret")
	if err != nil {
		t.Fatal(err)
	}
	if err := client.DeleteDraft(t.Context(), ""); err == nil {
		t.Fatal("DeleteDraft accepted empty id")
	}
	if _, err := client.DownloadFile(t.Context(), RemoteFile{}, filepath.Join(t.TempDir(), "out"), download.ConflictFail); err == nil {
		t.Fatal("DownloadFile accepted empty remote")
	}
	if _, err := client.UploadFile(t.Context(), Draft{}, "missing", "../escape", download.ConflictFail); err == nil {
		t.Fatal("UploadFile accepted invalid draft and filename")
	}
}

func mustURL(t *testing.T, value string) *url.URL {
	t.Helper()
	parsed, err := url.Parse(value)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}
