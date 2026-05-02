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
