package download

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResumeStreamAtomicallyRecoversAfterInterruptedRead(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "result.txt")
	var offsets []int64
	first := true
	opener := func(offset int64) (io.ReadCloser, error) {
		offsets = append(offsets, offset)
		if first {
			first = false
			return io.NopCloser(&interruptedReader{data: "abcdef", limit: 3}), nil
		}
		return io.NopCloser(strings.NewReader("def")), nil
	}

	_, err := ResumeStreamAtomically(opener, ResumeOptions{Destination: dst, Source: "https://osf.example/file", Policy: ConflictOverwrite})
	if err == nil {
		t.Fatal("first transfer returned nil error, want interruption")
	}
	if _, err := os.Stat(resumeCheckpointPath(dst)); err != nil {
		t.Fatalf("checkpoint stat: %v", err)
	}

	result, err := ResumeStreamAtomically(opener, ResumeOptions{Destination: dst, Source: "https://osf.example/file", Policy: ConflictOverwrite})
	if err != nil {
		t.Fatalf("resume returned error: %v", err)
	}
	if !result.Resumed || len(offsets) != 2 || offsets[0] != 0 || offsets[1] != 3 {
		t.Fatalf("resume result=%+v offsets=%v", result, offsets)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "abcdef" {
		t.Fatalf("destination = %q, want abcdef", got)
	}
	if _, err := os.Stat(resumeCheckpointPath(dst)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("checkpoint still exists: %v", err)
	}
}

func TestResumeStreamAtomicallyInvalidatesMismatchedCheckpoint(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "result.txt")
	first := true
	opener := func(offset int64) (io.ReadCloser, error) {
		if first {
			first = false
			return io.NopCloser(&interruptedReader{data: "old", limit: 1}), nil
		}
		if offset != 0 {
			t.Fatalf("mismatched source resumed at offset %d", offset)
		}
		return io.NopCloser(strings.NewReader("new")), nil
	}
	_, _ = ResumeStreamAtomically(opener, ResumeOptions{Destination: dst, Source: "old-source", Policy: ConflictOverwrite})
	result, err := ResumeStreamAtomically(opener, ResumeOptions{Destination: dst, Source: "new-source", Policy: ConflictOverwrite})
	if err != nil {
		t.Fatalf("mismatched resume returned error: %v", err)
	}
	if result.Resumed {
		t.Fatal("mismatched checkpoint reported resumed")
	}
	got, err := os.ReadFile(dst)
	if err != nil || string(got) != "new" {
		t.Fatalf("destination=%q err=%v, want new", got, err)
	}
}

func TestResumeStreamAtomicallyRejectsChecksumMismatch(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "result.txt")
	result, err := ResumeStreamAtomically(func(int64) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader("content")), nil
	}, ResumeOptions{Destination: dst, Source: "source", Policy: ConflictOverwrite, ExpectedChecksum: "sha256:0000"})
	if err == nil || result.Completed {
		t.Fatalf("checksum result=%+v err=%v, want failure", result, err)
	}
	if _, statErr := os.Stat(dst); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("destination stat=%v, want absent", statErr)
	}
}

func TestResumeStreamAtomicallyRestartsWhenRangeUnsupported(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "result.txt")
	first := true
	var offsets []int64
	opener := func(offset int64) (io.ReadCloser, error) {
		offsets = append(offsets, offset)
		if first {
			first = false
			return io.NopCloser(&interruptedReader{data: "abc", limit: 1}), nil
		}
		if offset > 0 {
			return nil, ErrRangeUnsupported
		}
		return io.NopCloser(strings.NewReader("abc")), nil
	}
	_, _ = ResumeStreamAtomically(opener, ResumeOptions{Destination: dst, Source: "source", Policy: ConflictOverwrite})
	result, err := ResumeStreamAtomically(opener, ResumeOptions{Destination: dst, Source: "source", Policy: ConflictOverwrite})
	if err != nil || !result.Completed || result.Resumed {
		t.Fatalf("result=%+v err=%v, want clean restart", result, err)
	}
	if len(offsets) != 3 || offsets[1] != 1 || offsets[2] != 0 {
		t.Fatalf("offsets=%v, want [0 1 0]", offsets)
	}
}

func FuzzLoadResumeCheckpointRejectsMalformedData(f *testing.F) {
	f.Add("{")
	f.Add("null")
	f.Add(`{"version":"wrong"}`)
	f.Fuzz(func(t *testing.T, data string) {
		dir := t.TempDir()
		checkpointPath := filepath.Join(dir, "file.resume.json")
		partialPath := filepath.Join(dir, "file.part")
		if err := os.WriteFile(checkpointPath, []byte(data), 0o600); err != nil {
			t.Fatal(err)
		}
		_, _, err := loadResumeCheckpoint(checkpointPath, partialPath, ResumeOptions{
			Destination: filepath.Join(dir, "file"),
			Source:      "source",
		})
		if err != nil {
			t.Fatalf("loadResumeCheckpoint returned error for malformed data: %v", err)
		}
	})
}

type interruptedReader struct {
	data  string
	limit int
	read  int
}

func (r *interruptedReader) Read(p []byte) (int, error) {
	if r.read >= r.limit {
		return 0, errors.New("simulated interruption")
	}
	remaining := r.limit - r.read
	if remaining > len(p) {
		remaining = len(p)
	}
	copy(p[:remaining], r.data[r.read:r.read+remaining])
	r.read += remaining
	return remaining, nil
}
