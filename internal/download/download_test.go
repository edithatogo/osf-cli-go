package download

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNormalizeRemotePathRejectsTraversal(t *testing.T) {
	for _, remote := range []string{"../secret.txt", "folder/../../secret.txt", `..\secret.txt`} {
		if _, err := NormalizeRemotePath(remote); !errors.Is(err, errPathTraversal) {
			t.Fatalf("NormalizeRemotePath(%q) error = %v, want traversal error", remote, err)
		}
	}
}

func TestResolveDestinationRejectsTraversal(t *testing.T) {
	root := t.TempDir()

	if _, err := ResolveDestination(root, "../secret.txt"); !errors.Is(err, errPathTraversal) {
		t.Fatalf("ResolveDestination error = %v, want traversal error", err)
	}
}

func TestResolveDestinationRejectsSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	link := filepath.Join(root, "link")

	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlink creation unavailable on this platform: %v", err)
	}

	if _, err := ResolveDestination(root, "link/secret.txt"); !errors.Is(err, errPathTraversal) {
		t.Fatalf("ResolveDestination error = %v, want traversal error", err)
	}
}

func TestWriteStreamAtomicallyFailWhenDestinationExists(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "existing.txt")
	if err := os.WriteFile(dst, []byte("original"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	written, err := WriteStreamAtomically(dst, strings.NewReader("new"), 0o600, ConflictFail)
	if written {
		t.Fatalf("written = true, want false")
	}
	if !errors.Is(err, errDestinationExists) {
		t.Fatalf("err = %v, want destination exists error", err)
	}

	got, readErr := os.ReadFile(dst)
	if readErr != nil {
		t.Fatalf("ReadFile: %v", readErr)
	}
	if string(got) != "original" {
		t.Fatalf("destination contents = %q, want %q", string(got), "original")
	}
}

func TestWriteStreamAtomicallySkipWhenDestinationExists(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "existing.txt")
	if err := os.WriteFile(dst, []byte("original"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	written, err := WriteStreamAtomically(dst, strings.NewReader("new"), 0o600, ConflictSkip)
	if err != nil {
		t.Fatalf("WriteStreamAtomically returned error: %v", err)
	}
	if written {
		t.Fatalf("written = true, want false")
	}

	got, readErr := os.ReadFile(dst)
	if readErr != nil {
		t.Fatalf("ReadFile: %v", readErr)
	}
	if string(got) != "original" {
		t.Fatalf("destination contents = %q, want %q", string(got), "original")
	}
}

func TestWriteStreamAtomicallyOverwriteWhenDestinationExists(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "existing.txt")
	if err := os.WriteFile(dst, []byte("original"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	written, err := WriteStreamAtomically(dst, strings.NewReader("replacement"), 0o600, ConflictOverwrite)
	if err != nil {
		t.Fatalf("WriteStreamAtomically returned error: %v", err)
	}
	if !written {
		t.Fatalf("written = false, want true")
	}

	got, readErr := os.ReadFile(dst)
	if readErr != nil {
		t.Fatalf("ReadFile: %v", readErr)
	}
	if string(got) != "replacement" {
		t.Fatalf("destination contents = %q, want %q", string(got), "replacement")
	}
}

func TestWriteStreamAtomicallyCleansTempOnReaderFailure(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "output.txt")
	r := failingReader{}

	_, err := WriteStreamAtomically(dst, r, 0o600, ConflictFail)
	if err == nil {
		t.Fatal("WriteStreamAtomically returned nil error, want reader failure")
	}

	matches, globErr := filepath.Glob(filepath.Join(dir, ".*.tmp"))
	if globErr != nil {
		t.Fatalf("Glob: %v", globErr)
	}
	if len(matches) != 0 {
		t.Fatalf("temporary files left behind: %v", matches)
	}
	if _, statErr := os.Stat(dst); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("destination stat error = %v, want not exist", statErr)
	}
}

func TestWriteStreamAtomicallyWritesFile(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "nested", "output.txt")

	written, err := WriteStreamAtomically(dst, strings.NewReader("hello world"), 0o640, ConflictFail)
	if err != nil {
		t.Fatalf("WriteStreamAtomically returned error: %v", err)
	}
	if !written {
		t.Fatalf("written = false, want true")
	}

	got, readErr := os.ReadFile(dst)
	if readErr != nil {
		t.Fatalf("ReadFile: %v", readErr)
	}
	if string(got) != "hello world" {
		t.Fatalf("destination contents = %q, want %q", string(got), "hello world")
	}
}

func TestConflictPolicyValidation(t *testing.T) {
	for _, policy := range []ConflictPolicy{ConflictFail, ConflictSkip, ConflictOverwrite} {
		if err := policy.Validate(); err != nil {
			t.Fatalf("Validate(%q) returned error: %v", policy, err)
		}
		got, err := ParseConflictPolicy(string(policy))
		if err != nil {
			t.Fatalf("ParseConflictPolicy(%q) returned error: %v", policy, err)
		}
		if got != policy {
			t.Fatalf("ParseConflictPolicy(%q) = %q, want %q", policy, got, policy)
		}
	}

	if err := ConflictPolicy("bad").Validate(); err == nil {
		t.Fatal("Validate(bad) returned nil error, want validation error")
	}
	if _, err := ParseConflictPolicy("bad"); err == nil {
		t.Fatal("ParseConflictPolicy(bad) returned nil error, want validation error")
	}
}

type failingReader struct{}

func (failingReader) Read(_ []byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}
