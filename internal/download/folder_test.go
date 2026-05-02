package download

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewFolderDownloadPlanRejectsTraversal(t *testing.T) {
	dir := t.TempDir()

	_, err := NewFolderDownloadPlan(dir, ConflictFail, []FolderDownloadFile{
		{RemotePath: "../secret.txt", Reader: strings.NewReader("x")},
	})
	if !errors.Is(err, errPathTraversal) {
		t.Fatalf("NewFolderDownloadPlan error = %v, want traversal error", err)
	}
}

func TestNewFolderDownloadPlanRejectsNoReaderOrOpener(t *testing.T) {
	dir := t.TempDir()
	_, err := NewFolderDownloadPlan(dir, ConflictFail, []FolderDownloadFile{
		{RemotePath: "no-source.txt"},
	})
	if err == nil {
		t.Fatal("NewFolderDownloadPlan returned nil error, want error")
	}
	if !strings.Contains(err.Error(), "requires a reader or opener") {
		t.Fatalf("error = %q, want reader or opener message", err.Error())
	}
}

func TestNewFolderDownloadPlanRejectsEmptyDestRoot(t *testing.T) {
	_, err := NewFolderDownloadPlan("", ConflictFail, []FolderDownloadFile{
		{RemotePath: "file.txt", Reader: strings.NewReader("x")},
	})
	if err == nil {
		t.Fatal("NewFolderDownloadPlan returned nil error, want error")
	}
}

func TestFolderDownloadPlanWritesNestedPaths(t *testing.T) {
	dir := t.TempDir()
	expectedBytes := int64(len("beta"))

	plan, err := NewFolderDownloadPlan(dir, ConflictFail, []FolderDownloadFile{
		{RemotePath: "alpha.txt", Reader: strings.NewReader("alpha")},
		{RemotePath: "nested/dir/bravo.txt", Open: func() (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("beta")), nil
		}, KnownBytes: &expectedBytes},
	})
	if err != nil {
		t.Fatalf("NewFolderDownloadPlan: %v", err)
	}

	manifest, err := plan.Execute()
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if got, want := len(manifest.Records), 2; got != want {
		t.Fatalf("manifest records = %d, want %d", got, want)
	}

	alpha := manifest.Records[0]
	if alpha.Status != FolderDownloadWritten {
		t.Fatalf("alpha status = %q, want %q", alpha.Status, FolderDownloadWritten)
	}
	if alpha.Bytes == nil || *alpha.Bytes != int64(len("alpha")) {
		t.Fatalf("alpha bytes = %v, want %d", alpha.Bytes, len("alpha"))
	}

	bravo := manifest.Records[1]
	if bravo.Status != FolderDownloadWritten {
		t.Fatalf("bravo status = %q, want %q", bravo.Status, FolderDownloadWritten)
	}
	if bravo.Bytes == nil || *bravo.Bytes != int64(len("beta")) {
		t.Fatalf("bravo bytes = %v, want %d", bravo.Bytes, len("beta"))
	}

	for rel, want := range map[string]string{
		"alpha.txt": "alpha",
		filepath.Join("nested", "dir", "bravo.txt"): "beta",
	} {
		got, readErr := os.ReadFile(filepath.Join(dir, rel))
		if readErr != nil {
			t.Fatalf("ReadFile(%q): %v", rel, readErr)
		}
		if string(got) != want {
			t.Fatalf("file %q = %q, want %q", rel, string(got), want)
		}
	}
}

func TestFolderDownloadPlanSkipsExistingFile(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "existing.txt")
	if err := os.WriteFile(dst, []byte("original"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	expectedBytes := int64(4)

	plan, err := NewFolderDownloadPlan(dir, ConflictSkip, []FolderDownloadFile{
		{RemotePath: "existing.txt", Reader: failingReader{}, KnownBytes: &expectedBytes},
	})
	if err != nil {
		t.Fatalf("NewFolderDownloadPlan: %v", err)
	}

	manifest, err := plan.Execute()
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if len(manifest.Records) != 1 {
		t.Fatalf("manifest records = %d, want 1", len(manifest.Records))
	}
	record := manifest.Records[0]
	if record.Status != FolderDownloadSkipped {
		t.Fatalf("record status = %q, want %q", record.Status, FolderDownloadSkipped)
	}
	if record.Bytes == nil || *record.Bytes != expectedBytes {
		t.Fatalf("record bytes = %v, want %d", record.Bytes, expectedBytes)
	}

	got, readErr := os.ReadFile(dst)
	if readErr != nil {
		t.Fatalf("ReadFile: %v", readErr)
	}
	if string(got) != "original" {
		t.Fatalf("destination contents = %q, want %q", string(got), "original")
	}
}

func TestFolderDownloadPlanOverwritesExistingFile(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "existing.txt")
	if err := os.WriteFile(dst, []byte("original"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	plan, err := NewFolderDownloadPlan(dir, ConflictOverwrite, []FolderDownloadFile{
		{RemotePath: "existing.txt", Open: func() (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("replacement")), nil
		}},
	})
	if err != nil {
		t.Fatalf("NewFolderDownloadPlan: %v", err)
	}

	manifest, err := plan.Execute()
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	record := manifest.Records[0]
	if record.Status != FolderDownloadWritten {
		t.Fatalf("record status = %q, want %q", record.Status, FolderDownloadWritten)
	}
	if record.Bytes == nil || *record.Bytes != int64(len("replacement")) {
		t.Fatalf("record bytes = %v, want %d", record.Bytes, len("replacement"))
	}

	got, readErr := os.ReadFile(dst)
	if readErr != nil {
		t.Fatalf("ReadFile: %v", readErr)
	}
	if string(got) != "replacement" {
		t.Fatalf("destination contents = %q, want %q", string(got), "replacement")
	}
}

func TestFolderDownloadPlanRecordsFailureAndCleansTemp(t *testing.T) {
	dir := t.TempDir()
	failingDst := filepath.Join(dir, "fail.txt")

	plan, err := NewFolderDownloadPlan(dir, ConflictFail, []FolderDownloadFile{
		{RemotePath: "ok.txt", Reader: strings.NewReader("ok")},
		{RemotePath: "fail.txt", Reader: failingReader{}},
	})
	if err != nil {
		t.Fatalf("NewFolderDownloadPlan: %v", err)
	}

	manifest, execErr := plan.Execute()
	if execErr == nil {
		t.Fatal("Execute returned nil error, want failure")
	}
	if len(manifest.Records) != 2 {
		t.Fatalf("manifest records = %d, want 2", len(manifest.Records))
	}
	if manifest.Records[0].Status != FolderDownloadWritten {
		t.Fatalf("first record status = %q, want written", manifest.Records[0].Status)
	}
	if manifest.Records[1].Status != FolderDownloadFailed {
		t.Fatalf("second record status = %q, want failed", manifest.Records[1].Status)
	}
	if manifest.Records[1].Error == "" {
		t.Fatal("second record error = empty, want failure details")
	}

	matches, globErr := filepath.Glob(filepath.Join(dir, ".*.tmp"))
	if globErr != nil {
		t.Fatalf("Glob: %v", globErr)
	}
	if len(matches) != 0 {
		t.Fatalf("temporary files left behind: %v", matches)
	}

	got, readErr := os.ReadFile(filepath.Join(dir, "ok.txt"))
	if readErr != nil {
		t.Fatalf("ReadFile(ok.txt): %v", readErr)
	}
	if string(got) != "ok" {
		t.Fatalf("ok.txt contents = %q, want %q", string(got), "ok")
	}
	if _, statErr := os.Stat(failingDst); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("fail.txt stat error = %v, want not exist", statErr)
	}
}

func TestNewFolderDownloadPlanRejectsSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	link := filepath.Join(root, "link")

	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlink creation unavailable on this platform: %v", err)
	}

	_, err := NewFolderDownloadPlan(root, ConflictFail, []FolderDownloadFile{
		{RemotePath: "link/secret.txt", Reader: strings.NewReader("secret")},
	})
	if !errors.Is(err, errPathTraversal) {
		t.Fatalf("NewFolderDownloadPlan error = %v, want traversal error", err)
	}
}

func TestFolderDownloadPlanRecordsCloseFailure(t *testing.T) {
	dir := t.TempDir()

	plan, err := NewFolderDownloadPlan(dir, ConflictFail, []FolderDownloadFile{
		{RemotePath: "close-fail.txt", Open: func() (io.ReadCloser, error) {
			return closeErrorReadCloser{Reader: strings.NewReader("payload")}, nil
		}},
	})
	if err != nil {
		t.Fatalf("NewFolderDownloadPlan: %v", err)
	}

	manifest, execErr := plan.Execute()
	if execErr == nil {
		t.Fatal("Execute returned nil error, want close failure")
	}
	if len(manifest.Records) != 1 {
		t.Fatalf("manifest records = %d, want 1", len(manifest.Records))
	}
	record := manifest.Records[0]
	if record.Status != FolderDownloadFailed {
		t.Fatalf("record status = %q, want failed", record.Status)
	}
	if !strings.Contains(record.Error, "close download source") {
		t.Fatalf("record error = %q, want close failure message", record.Error)
	}
}

func TestFolderDownloadPlanRejectsNilOpenerReader(t *testing.T) {
	dir := t.TempDir()

	plan, err := NewFolderDownloadPlan(dir, ConflictFail, []FolderDownloadFile{
		{RemotePath: "nil.txt", Open: func() (io.ReadCloser, error) {
			return nil, nil
		}},
	})
	if err != nil {
		t.Fatalf("NewFolderDownloadPlan: %v", err)
	}

	manifest, execErr := plan.Execute()
	if execErr == nil {
		t.Fatal("Execute returned nil error, want nil opener failure")
	}
	if len(manifest.Records) != 1 {
		t.Fatalf("manifest records = %d, want 1", len(manifest.Records))
	}
	if manifest.Records[0].Status != FolderDownloadFailed {
		t.Fatalf("record status = %q, want failed", manifest.Records[0].Status)
	}
}

func TestExecuteWithNilPlan(t *testing.T) {
	var p *FolderDownloadPlan
	_, err := p.Execute()
	if err == nil {
		t.Fatal("Execute returned nil error for nil plan, want error")
	}
	if !strings.Contains(err.Error(), "folder download plan is required") {
		t.Fatalf("error = %q, want plan required message", err.Error())
	}
}

func TestOpenReaderWithReadCloserReader(t *testing.T) {
	file := plannedFolderDownloadFile{
		remotePath: "rc.txt",
		reader:     io.NopCloser(strings.NewReader("data")),
	}
	src, closeFn, err := file.openReader()
	if err != nil {
		t.Fatalf("openReader returned error: %v", err)
	}
	if closeFn == nil {
		t.Fatal("openReader returned nil closeFn, want non-nil")
	}
	body, err := io.ReadAll(src)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if string(body) != "data" {
		t.Fatalf("body = %q, want %q", string(body), "data")
	}
}

func TestNormalizeRemotePathRejectsDotOnly(t *testing.T) {
	for _, path := range []string{".", "/", "./", "/."} {
		_, err := NormalizeRemotePath(path)
		if err == nil {
			t.Fatalf("NormalizeRemotePath(%q) returned nil error, want error", path)
		}
	}
}

func TestFolderExecuteWithOpenError(t *testing.T) {
	dir := t.TempDir()
	plan, err := NewFolderDownloadPlan(dir, ConflictFail, []FolderDownloadFile{
		{RemotePath: "fail.bin", Open: func() (io.ReadCloser, error) {
			return nil, io.ErrUnexpectedEOF
		}},
	})
	if err != nil {
		t.Fatalf("NewFolderDownloadPlan: %v", err)
	}
	manifest, execErr := plan.Execute()
	if execErr == nil {
		t.Fatal("Execute returned nil error, want open failure")
	}
	if len(manifest.Records) != 1 || manifest.Records[0].Status != FolderDownloadFailed {
		t.Fatalf("manifest = %+v, want failed record", manifest)
	}
}

func TestFolderDownloadPlanSkipsWhenNoOpenAndNoReader(t *testing.T) {
	dir := t.TempDir()
	file := plannedFolderDownloadFile{
		remotePath: "orphan.txt",
	}
	_, _, err := file.openReader()
	if err == nil {
		t.Fatal("openReader returned nil error, want error for no reader or opener")
	}
	if !strings.Contains(err.Error(), "requires a reader or opener") {
		t.Fatalf("error = %q, want reader or opener message", err.Error())
	}

	// Also test that the plan rejects an empty file list entry
	_, planErr := NewFolderDownloadPlan(dir, ConflictFail, nil)
	if planErr != nil {
		t.Fatalf("NewFolderDownloadPlan with nil files returned error: %v", planErr)
	}
}

func TestOpenReaderWithPlainReader(t *testing.T) {
	file := plannedFolderDownloadFile{
		remotePath: "plain.txt",
		reader:     strings.NewReader("plain-data"),
	}
	src, closeFn, err := file.openReader()
	if err != nil {
		t.Fatalf("openReader returned error: %v", err)
	}
	if closeFn != nil {
		t.Fatal("openReader returned non-nil closeFn for plain reader, want nil")
	}
	body, err := io.ReadAll(src)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if string(body) != "plain-data" {
		t.Fatalf("body = %q, want %q", string(body), "plain-data")
	}
}

type closeErrorReadCloser struct {
	io.Reader
}

func (closeErrorReadCloser) Close() error {
	return io.ErrClosedPipe
}
