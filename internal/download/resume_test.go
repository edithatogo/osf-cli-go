package download

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/edithatogo/osf-cli-go/internal/observability"
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

func TestResumeStreamAtomicallyValidatesInputsAndConflicts(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "result.txt")
	open := func(int64) (io.ReadCloser, error) { return io.NopCloser(strings.NewReader("x")), nil }
	for name, opts := range map[string]ResumeOptions{
		"missing opener":      {Destination: dst, Source: "source", Policy: ConflictOverwrite},
		"missing destination": {Source: "source", Policy: ConflictOverwrite},
		"missing source":      {Destination: dst, Policy: ConflictOverwrite},
		"invalid policy":      {Destination: dst, Source: "source", Policy: ConflictPolicy("unknown")},
	} {
		var opener StreamOpener = open
		if name == "missing opener" {
			opener = nil
		}
		if _, err := ResumeStreamAtomically(opener, opts); err == nil {
			t.Fatalf("%s returned nil error", name)
		}
	}
	if err := os.WriteFile(dst, []byte("existing"), 0o600); err != nil {
		t.Fatal(err)
	}
	result, err := ResumeStreamAtomically(open, ResumeOptions{Destination: dst, Source: "source", Policy: ConflictSkip})
	if err != nil || !result.Completed || result.CheckpointPath != "" {
		t.Fatalf("skip result=%+v err=%v", result, err)
	}
	if _, err := ResumeStreamAtomically(open, ResumeOptions{Destination: dst, Source: "source", Policy: ConflictFail}); !errors.Is(err, errDestinationExists) {
		t.Fatalf("fail conflict error=%v, want destination exists", err)
	}
}

func TestResumeStreamAtomicallyValidatesSizeAndMD5(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "result.txt")
	wrongSize := int64(4)
	result, err := ResumeStreamAtomically(func(int64) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader("abc")), nil
	}, ResumeOptions{Destination: dst, Source: "source", ExpectedSize: &wrongSize, Policy: ConflictOverwrite})
	if err == nil || result.Completed {
		t.Fatalf("size result=%+v err=%v, want mismatch", result, err)
	}
	correctMD5 := "md5:900150983cd24fb0d6963f7d28e17f72"
	result, err = ResumeStreamAtomically(func(int64) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader("abc")), nil
	}, ResumeOptions{Destination: dst, Source: "source-2", ExpectedChecksum: correctMD5, Policy: ConflictOverwrite})
	if err != nil || !result.Completed || result.Checksum != correctMD5 {
		t.Fatalf("md5 result=%+v err=%v", result, err)
	}
}

func TestResumeStreamAtomicallyOverwritesExistingDestination(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "result.txt")
	if err := os.WriteFile(dst, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	result, err := ResumeStreamAtomically(stringsReaderOpener("new"), ResumeOptions{
		Destination: dst,
		Source:      "source-overwrite",
		Policy:      ConflictOverwrite,
	})
	if err != nil || !result.Completed {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	got, err := os.ReadFile(dst)
	if err != nil || string(got) != "new" {
		t.Fatalf("destination=%q err=%v", got, err)
	}
}

func TestResumeStreamAtomicallyRejectsUnremovableDestination(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "result.txt")
	if err := os.Mkdir(dst, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dst, "keep"), []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := ResumeStreamAtomically(stringsReaderOpener("new"), ResumeOptions{Destination: dst, Source: "source", Policy: ConflictOverwrite})
	if err == nil || !strings.Contains(err.Error(), "remove existing destination") {
		t.Fatalf("error=%v, want destination removal error", err)
	}
}

func TestResumeStreamAtomicallyPreservesReadFailure(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "result.txt")
	result, err := ResumeStreamAtomically(func(int64) (io.ReadCloser, error) {
		return io.NopCloser(&errorReader{data: "partial", err: errors.New("remote read failed")}), nil
	}, ResumeOptions{Destination: dst, Source: "source-read-error", Policy: ConflictOverwrite})
	if err == nil || !strings.Contains(err.Error(), "remote read failed") || result.Completed {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}

func TestResumeStreamAtomicallyClassifiesCancellation(t *testing.T) {
	var events strings.Builder
	ctx := observability.WithEmitter(context.Background(), observability.NewJSONEmitter(&events, observability.LevelInfo))
	ctx, cancel := context.WithCancel(ctx)
	cancel()
	_, err := ResumeStreamAtomically(func(int64) (io.ReadCloser, error) {
		return nil, context.Canceled
	}, ResumeOptions{Destination: filepath.Join(t.TempDir(), "result.txt"), Source: "source-cancel", Policy: ConflictOverwrite, Context: ctx})
	if !errors.Is(err, context.Canceled) || !strings.Contains(events.String(), `"outcome":"canceled"`) {
		t.Fatalf("err=%v events=%q", err, events.String())
	}
}

func TestResumeStreamAtomicallyPropagatesSourceAndPathErrors(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "result.txt")
	_, err := ResumeStreamAtomically(func(int64) (io.ReadCloser, error) {
		return nil, errors.New("source unavailable")
	}, ResumeOptions{Destination: dst, Source: "source", Policy: ConflictOverwrite})
	if err == nil || !strings.Contains(err.Error(), "source unavailable") {
		t.Fatalf("source error=%v", err)
	}
	parent := filepath.Join(t.TempDir(), "parent")
	if err := os.WriteFile(parent, []byte("file"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err = ResumeStreamAtomically(stringsReaderOpener("x"), ResumeOptions{Destination: filepath.Join(parent, "result.txt"), Source: "source", Policy: ConflictOverwrite})
	if err == nil {
		t.Fatal("path error returned nil")
	}
}

func TestResumeStreamAtomicallyHandlesMalformedCheckpointAndPartialPath(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "result.txt")
	if err := os.WriteFile(resumeCheckpointPath(dst), []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	result, err := ResumeStreamAtomically(stringsReaderOpener("fresh"), ResumeOptions{Destination: dst, Source: "source", Policy: ConflictOverwrite})
	if err != nil || !result.Completed {
		t.Fatalf("malformed checkpoint result=%+v err=%v", result, err)
	}

	dst = filepath.Join(dir, "directory-part.txt")
	if err := os.Mkdir(resumePartialPath(dst), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(resumePartialPath(dst), "keep"), []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ResumeStreamAtomically(stringsReaderOpener("content"), ResumeOptions{Destination: dst, Source: "source", Policy: ConflictOverwrite}); err == nil || !strings.Contains(err.Error(), "open resume partial") {
		t.Fatalf("partial directory error=%v, want open partial error", err)
	}
}

func TestResumeStreamAtomicallyRetainsCheckpointWriteFailure(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "result.txt")
	checkpoint := resumeCheckpointPath(dst)
	first := true
	_, err := ResumeStreamAtomically(func(int64) (io.ReadCloser, error) {
		if first {
			first = false
			if err := os.Remove(checkpoint); err != nil {
				t.Fatal(err)
			}
			if err := os.Mkdir(checkpoint, 0o700); err != nil {
				t.Fatal(err)
			}
		}
		return io.NopCloser(strings.NewReader("content")), nil
	}, ResumeOptions{Destination: dst, Source: "source", Policy: ConflictOverwrite})
	if err == nil || !strings.Contains(err.Error(), "finalize resume checkpoint") {
		t.Fatalf("checkpoint write error=%v, want finalize error", err)
	}
}

func TestResumeStreamAtomicallyRejectsStalledReaders(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "result.txt")
	result, err := ResumeStreamAtomically(func(int64) (io.ReadCloser, error) {
		return io.NopCloser(zeroReader{}), nil
	}, ResumeOptions{Destination: dst, Source: "source", Policy: ConflictOverwrite})
	if !errors.Is(err, io.ErrNoProgress) || result.Completed {
		t.Fatalf("result=%+v err=%v, want no-progress failure", result, err)
	}
	if _, statErr := os.Stat(resumeCheckpointPath(dst)); statErr != nil {
		t.Fatalf("checkpoint stat=%v, want retained checkpoint", statErr)
	}
}

func TestResumeStreamAtomicallyEmitsTransferEvent(t *testing.T) {
	var events strings.Builder
	ctx := observability.WithOperationID(context.Background(), "op-transfer-test")
	ctx = observability.WithEmitter(ctx, observability.NewJSONEmitter(&events, observability.LevelInfo))
	dst := filepath.Join(t.TempDir(), "result.txt")
	if _, err := ResumeStreamAtomically(stringsReaderOpener("content"), ResumeOptions{Destination: dst, Source: "https://files.osf.io/download", Policy: ConflictOverwrite, Context: ctx}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(events.String(), `"name":"transfer.download"`) || !strings.Contains(events.String(), `"operationId":"op-transfer-test"`) {
		t.Fatalf("events=%q", events.String())
	}
}

func stringsReaderOpener(value string) StreamOpener {
	return func(int64) (io.ReadCloser, error) { return io.NopCloser(strings.NewReader(value)), nil }
}

func TestResumeCheckpointHelpersHandleFilesystemErrors(t *testing.T) {
	checkpoint := resumeCheckpoint{Version: resumeCheckpointVersion, Source: "source", Destination: "destination"}
	if err := writeResumeCheckpoint(filepath.Join(t.TempDir(), "missing", "checkpoint.json"), checkpoint); err == nil {
		t.Fatal("missing checkpoint directory returned nil error")
	}
	parent := filepath.Join(t.TempDir(), "parent")
	if err := os.WriteFile(parent, []byte("file"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := writeResumeCheckpoint(filepath.Join(parent, "checkpoint.json"), checkpoint); err == nil {
		t.Fatal("file checkpoint parent returned nil error")
	}
	valid := filepath.Join(t.TempDir(), "checkpoint.json")
	if err := writeResumeCheckpoint(valid, checkpoint); err != nil {
		t.Fatalf("writeResumeCheckpoint(valid): %v", err)
	}
	if _, err := os.ReadFile(valid); err != nil {
		t.Fatalf("checkpoint read: %v", err)
	}
	finalizeDir := filepath.Join(t.TempDir(), "existing")
	if err := os.Mkdir(finalizeDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := writeResumeCheckpoint(finalizeDir, checkpoint); err == nil {
		t.Fatal("existing checkpoint directory returned nil error")
	}
	if _, err := checksumFile(filepath.Join(t.TempDir(), "missing"), ""); err == nil {
		t.Fatal("missing checksum source returned nil error")
	}
}

func TestResumeHelperContracts(t *testing.T) {
	t.Parallel()
	if got := transferEventLevel(observability.OutcomeOK); got != observability.LevelInfo {
		t.Fatalf("success event level = %q", got)
	}
	if got := transferEventLevel(observability.OutcomeError); got != observability.LevelError {
		t.Fatalf("error event level = %q", got)
	}
	if got, _ := checksumHash("md5:abc"); got == nil {
		t.Fatal("md5 checksum hash is nil")
	}
	for _, test := range []struct {
		input, want string
	}{
		{"", ""},
		{"ABC", "sha256:abc"},
		{"MD5:ABC", "md5:abc"},
	} {
		if got := expectedChecksumValue(test.input); got != test.want {
			t.Fatalf("expectedChecksumValue(%q) = %q, want %q", test.input, got, test.want)
		}
	}
	if !equalInt64Ptr(nil, nil) || equalInt64Ptr(nil, new(int64)) || !equalInt64Ptr(newInt64(2), newInt64(2)) || equalInt64Ptr(newInt64(2), newInt64(3)) {
		t.Fatal("equalInt64Ptr contract mismatch")
	}
	if sourceFingerprint("source") == sourceFingerprint("other") || !strings.HasPrefix(sourceFingerprint("source"), "sha256:") {
		t.Fatal("source fingerprint contract mismatch")
	}
	if got := resumePartialPath("file"); got != "file.part" || resumeCheckpointPath("file") != "file.resume.json" {
		t.Fatalf("checkpoint paths = %q/%q", got, resumeCheckpointPath("file"))
	}

	dir := t.TempDir()
	checkpointPath := filepath.Join(dir, "missing", "checkpoint.json")
	if _, _, err := loadResumeCheckpoint(checkpointPath, filepath.Join(dir, "partial"), ResumeOptions{Destination: filepath.Join(dir, "file"), Source: "source"}); err != nil {
		t.Fatalf("missing checkpoint = %v", err)
	}
	if _, _, err := loadResumeCheckpoint(dir, filepath.Join(dir, "partial"), ResumeOptions{Destination: filepath.Join(dir, "file"), Source: "source"}); err == nil {
		t.Fatal("directory checkpoint returned nil error")
	}
	if err := writeResumeCheckpoint(filepath.Join(dir, "missing", "checkpoint.json"), resumeCheckpoint{Version: resumeCheckpointVersion}); err == nil {
		t.Fatal("checkpoint write to missing directory returned nil")
	}
	if exists, err := pathExists(filepath.Join(dir, "missing")); err != nil || exists {
		t.Fatalf("missing path exists=%v err=%v", exists, err)
	}
	if _, err := checksumFile(filepath.Join(dir, "missing"), ""); err == nil {
		t.Fatal("checksumFile on missing path returned nil")
	}
}

func newInt64(value int64) *int64 { return &value }

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

type zeroReader struct{}

type errorReader struct {
	data string
	err  error
	done bool
}

func (zeroReader) Read([]byte) (int, error) { return 0, nil }

func (r *errorReader) Read(p []byte) (int, error) {
	if r.done {
		return 0, r.err
	}
	r.done = true
	return copy(p, r.data), nil
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
