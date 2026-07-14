package crossprovider

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type sandboxFile struct {
	id, name, checksum string
	content            []byte
}

func TestZenodoSandboxDestinationFailureReplayIntegrityAndCompensation(t *testing.T) {
	t.Parallel()
	const token = "sandbox-cross-provider-token"
	var server *httptest.Server
	var createCalls, metadataCalls, publishCalls atomic.Int32
	var deletedDraft atomic.Bool
	var mu sync.Mutex
	files := make(map[string]sandboxFile)

	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+token {
			t.Errorf("authorization = %q", r.Header.Get("Authorization"))
		}
		switch {
		case strings.Contains(r.URL.Path, "/actions/publish"):
			publishCalls.Add(1)
			http.Error(w, "publication is forbidden in this test", http.StatusForbidden)
		case r.Method == http.MethodPost && r.URL.Path == "/api/deposit/depositions":
			createCalls.Add(1)
			_, _ = fmt.Fprintf(w, `{"id":123,"links":{"bucket":%q}}`, server.URL+"/api/files/bucket-1")
		case r.Method == http.MethodPut && r.URL.Path == "/api/deposit/depositions/123":
			if metadataCalls.Add(1) == 1 {
				http.Error(w, "injected metadata failure", http.StatusServiceUnavailable)
				return
			}
			var payload struct {
				Metadata struct {
					Keywords []string `json:"keywords"`
				} `json:"metadata"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || len(payload.Metadata.Keywords) != 1 {
				t.Errorf("metadata payload=%+v err=%v", payload, err)
			}
			_, _ = io.WriteString(w, `{"id":123}`)
		case r.Method == http.MethodGet && r.URL.Path == "/api/deposit/depositions/123/files":
			mu.Lock()
			defer mu.Unlock()
			inventory := make([]map[string]any, 0, len(files))
			for _, file := range files {
				inventory = append(inventory, map[string]any{
					"id": file.id, "filename": file.name, "filesize": len(file.content), "checksum": file.checksum,
					"links": map[string]string{"download": server.URL + "/api/files/bucket-1/" + file.name},
				})
			}
			_ = json.NewEncoder(w).Encode(inventory)
		case r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/api/files/bucket-1/"):
			name := strings.TrimPrefix(r.URL.Path, "/api/files/bucket-1/")
			content, _ := io.ReadAll(r.Body)
			digest := md5.Sum(content)
			file := sandboxFile{id: "file-" + name, name: name, content: content, checksum: "md5:" + hex.EncodeToString(digest[:])}
			mu.Lock()
			files[name] = file
			mu.Unlock()
			_ = json.NewEncoder(w).Encode(map[string]any{"id": file.id, "filename": name, "filesize": len(content), "checksum": file.checksum})
		case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/api/deposit/depositions/123/files/"):
			id := strings.TrimPrefix(r.URL.Path, "/api/deposit/depositions/123/files/")
			mu.Lock()
			for name, file := range files {
				if file.id == id {
					delete(files, name)
				}
			}
			mu.Unlock()
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/deposit/depositions/123":
			deletedDraft.Store(true)
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "unexpected request", http.StatusNotFound)
		}
	}))
	defer server.Close()

	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "nested"), 0o700); err != nil {
		t.Fatal(err)
	}
	content := []byte("cross-provider verified bytes\n")
	if err := os.WriteFile(filepath.Join(root, "nested", "data.txt"), content, 0o600); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(content)
	request := validRequest(t)
	request.Source.Files = []File{{Path: "nested/data.txt", Size: int64(len(content)), Checksum: "sha256:" + hex.EncodeToString(digest[:])}}
	request.Source.Metadata.Identifiers = []Identifier{{Scheme: "url", Value: "https://osf.io/example/"}}
	request.Source.Metadata.Version = "2026.07"
	report, err := BuildMapping(request, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	checkpoint, err := NewCheckpoint(report)
	if err != nil {
		t.Fatal(err)
	}
	destination, err := NewZenodoSandboxDestination(server.URL+"/api/", token)
	if err != nil {
		t.Fatal(err)
	}
	failed, err := Execute(t.Context(), report, checkpoint, LocalSource{Root: root}, destination)
	if err == nil || failed.Checkpoint.Status != SagaPartial || createCalls.Load() != 1 {
		t.Fatalf("failed=%+v err=%v creates=%d", failed, err, createCalls.Load())
	}
	completed, err := Execute(t.Context(), report, failed.Checkpoint, LocalSource{Root: root}, destination)
	if err != nil || completed.Checkpoint.Status != SagaCompleted || completed.Partial.Published {
		t.Fatalf("completed=%+v err=%v", completed, err)
	}
	if createCalls.Load() != 1 || metadataCalls.Load() != 2 || publishCalls.Load() != 0 {
		t.Fatalf("creates=%d metadata=%d publish=%d", createCalls.Load(), metadataCalls.Load(), publishCalls.Load())
	}
	mu.Lock()
	provenanceFile, provenancePresent := files[provenanceFilename]
	_, dataPresent := files[zenodoRemoteName("nested/data.txt")]
	mu.Unlock()
	if !provenancePresent || !dataPresent {
		t.Fatalf("provenance=%t data=%t", provenancePresent, dataPresent)
	}
	var sidecar struct {
		Report Report `json:"report"`
	}
	if err := json.Unmarshal(provenanceFile.content, &sidecar); err != nil || sidecar.Report.Target.Version != "2026.07" || len(sidecar.Report.Target.Identifiers) != 1 {
		t.Fatalf("sidecar=%+v err=%v", sidecar, err)
	}
	mu.Lock()
	dataFile := files[zenodoRemoteName("nested/data.txt")]
	dataFile.checksum = "md5:00000000000000000000000000000000"
	files[dataFile.name] = dataFile
	mu.Unlock()
	if err := destination.VerifyDraft(t.Context(), completed.Checkpoint.DestinationRef, report, "verify-again"); err == nil || !strings.Contains(err.Error(), "checksum") {
		t.Fatalf("tampered verification error = %v", err)
	}
	compensated, err := Compensate(t.Context(), completed.Checkpoint, destination)
	if err != nil || compensated.Status != SagaCompensated || !deletedDraft.Load() {
		t.Fatalf("compensated=%+v deleted=%t err=%v", compensated, deletedDraft.Load(), err)
	}
}

func TestZenodoDestinationRejectsChangedSourceAndUnsafeProduction(t *testing.T) {
	t.Parallel()
	if _, err := NewZenodoSandboxDestination("https://zenodo.org/api/", "secret"); err == nil {
		t.Fatal("production destination returned nil error")
	}
	destination := &ZenodoSandboxDestination{receipts: make(map[string]FileReceipt)}
	temporary := filepath.Join(t.TempDir(), "changed")
	if err := os.WriteFile(temporary, []byte("changed"), 0o600); err != nil {
		t.Fatal(err)
	}
	reader, err := os.Open(temporary)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = reader.Close() }()
	_, err = destination.CopyFile(t.Context(), "draft", File{Path: "changed", Size: 7, Checksum: "sha256:0000"}, reader, ConflictFail, "step")
	if err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("error = %v", err)
	}
}

func TestZenodoDestinationRejectsMutationsWithoutRollbackSnapshots(t *testing.T) {
	t.Parallel()
	destination := &ZenodoSandboxDestination{
		created: make(map[string]bool), receipts: make(map[string]FileReceipt),
	}
	if _, err := destination.ApplyMetadata(t.Context(), "existing", Metadata{}, Provenance{}, "step"); err == nil || !strings.Contains(err.Error(), "rollback snapshot") {
		t.Fatalf("metadata error = %v", err)
	}
	if _, err := destination.CopyFile(t.Context(), "existing", File{}, strings.NewReader(""), ConflictReplaceDraft, "step"); err == nil || !strings.Contains(err.Error(), "rollback snapshot") {
		t.Fatalf("replace error = %v", err)
	}
}
