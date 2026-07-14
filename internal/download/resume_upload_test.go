package download

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResumeFileUploadRecoversAcknowledgedChunk(t *testing.T) {
	source := filepath.Join(t.TempDir(), "source.txt")
	if err := os.WriteFile(source, []byte("abcdef"), 0o600); err != nil {
		t.Fatal(err)
	}
	checkpoint := filepath.Join(t.TempDir(), "upload.resume.json")
	first := true
	var offsets []int64
	session := func(_ context.Context, offset, total int64, content io.Reader) (int64, bool, error) {
		offsets = append(offsets, offset)
		if first {
			first = false
			buf := make([]byte, 3)
			if _, err := io.ReadFull(content, buf); err != nil {
				return offset, false, err
			}
			return 3, false, errors.New("provider connection lost")
		}
		body, err := io.ReadAll(content)
		if err != nil {
			return offset, false, err
		}
		if string(body) != "def" || total != 6 {
			t.Fatalf("upload body=%q total=%d", body, total)
		}
		return offset + int64(len(body)), true, nil
	}

	result, err := ResumeFileUpload(context.Background(), UploadOptions{SourcePath: source, SourceIdentity: "osf:file-1", CheckpointPath: checkpoint}, session)
	if err == nil || result.Completed {
		t.Fatalf("first result=%+v err=%v, want checkpointed failure", result, err)
	}
	result, err = ResumeFileUpload(context.Background(), UploadOptions{SourcePath: source, SourceIdentity: "osf:file-1", CheckpointPath: checkpoint}, session)
	if err != nil || !result.Completed || !result.Resumed {
		t.Fatalf("resume result=%+v err=%v", result, err)
	}
	if len(offsets) != 2 || offsets[0] != 0 || offsets[1] != 3 {
		t.Fatalf("offsets=%v, want [0 3]", offsets)
	}
	if _, err := os.Stat(checkpoint); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("checkpoint stat=%v, want removed", err)
	}
}

func TestResumeFileUploadRejectsNoProgress(t *testing.T) {
	source := filepath.Join(t.TempDir(), "source.txt")
	if err := os.WriteFile(source, []byte(strings.Repeat("a", 3)), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := ResumeFileUpload(context.Background(), UploadOptions{SourcePath: source, SourceIdentity: "source"}, func(context.Context, int64, int64, io.Reader) (int64, bool, error) {
		return 0, false, nil
	})
	if err == nil || !strings.Contains(err.Error(), "no progress") {
		t.Fatalf("error=%v, want no-progress error", err)
	}
}

func TestResumeFileUploadAcceptsProgressAcrossMultipleChunks(t *testing.T) {
	source := filepath.Join(t.TempDir(), "source.txt")
	if err := os.WriteFile(source, []byte("abcdef"), 0o600); err != nil {
		t.Fatal(err)
	}
	calls := 0
	result, err := ResumeFileUpload(context.Background(), UploadOptions{SourcePath: source, SourceIdentity: "source"}, func(_ context.Context, offset, total int64, content io.Reader) (int64, bool, error) {
		calls++
		body, readErr := io.ReadAll(io.LimitReader(content, 3))
		if readErr != nil {
			return offset, false, readErr
		}
		if total != 6 || len(body) == 0 {
			return offset, false, errors.New("unexpected upload chunk")
		}
		next := offset + int64(len(body))
		return next, next == total, nil
	})
	if err != nil || !result.Completed || calls != 2 {
		t.Fatalf("result=%+v err=%v calls=%d, want completed in two chunks", result, err, calls)
	}
}

func TestResumeFileUploadValidatesSourceAndAcknowledgements(t *testing.T) {
	ctx := context.Background()
	if _, err := ResumeFileUpload(ctx, UploadOptions{}, func(context.Context, int64, int64, io.Reader) (int64, bool, error) {
		return 0, false, nil
	}); err == nil {
		t.Fatal("missing upload options returned nil error")
	}
	dir := t.TempDir()
	if _, err := ResumeFileUpload(ctx, UploadOptions{SourcePath: dir, SourceIdentity: "source"}, func(context.Context, int64, int64, io.Reader) (int64, bool, error) {
		return 0, false, nil
	}); err == nil {
		t.Fatal("directory upload source returned nil error")
	}
	source := filepath.Join(dir, "source.txt")
	if err := os.WriteFile(source, []byte("abc"), 0o600); err != nil {
		t.Fatal(err)
	}
	for name, session := range map[string]UploadSession{
		"invalid offset": func(context.Context, int64, int64, io.Reader) (int64, bool, error) {
			return -1, false, nil
		},
		"early completion": func(context.Context, int64, int64, io.Reader) (int64, bool, error) {
			return 1, true, nil
		},
		"provider failure": func(context.Context, int64, int64, io.Reader) (int64, bool, error) {
			return 1, false, errors.New("provider failed")
		},
	} {
		checkpoint := filepath.Join(dir, name+".json")
		result, err := ResumeFileUpload(ctx, UploadOptions{SourcePath: source, SourceIdentity: name, CheckpointPath: checkpoint}, session)
		if err == nil || result.Completed {
			t.Fatalf("%s result=%+v err=%v, want failure", name, result, err)
		}
		if _, statErr := os.Stat(checkpoint); statErr != nil {
			t.Fatalf("%s checkpoint stat=%v, want retained checkpoint", name, statErr)
		}
	}
}

func TestResumeFileUploadInvalidatesStaleCheckpoint(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "source.txt")
	checkpoint := filepath.Join(dir, "upload.resume.json")
	if err := os.WriteFile(source, []byte("abc"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(checkpoint, []byte(`{"version":1,"sourcePathFingerprint":"stale"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	result, err := ResumeFileUpload(context.Background(), UploadOptions{SourcePath: source, SourceIdentity: "source", CheckpointPath: checkpoint}, func(_ context.Context, offset, total int64, content io.Reader) (int64, bool, error) {
		body, readErr := io.ReadAll(content)
		return offset + int64(len(body)), total == int64(len(body)), readErr
	})
	if err != nil || !result.Completed || result.Resumed {
		t.Fatalf("result=%+v err=%v, want fresh completion", result, err)
	}
}

func TestResumeFileUploadValidatesSessionAndCheckpointPaths(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "source.txt")
	if err := os.WriteFile(source, []byte("abc"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ResumeFileUpload(context.Background(), UploadOptions{SourcePath: source, SourceIdentity: "source"}, nil); err == nil {
		t.Fatal("nil upload session returned nil error")
	}
	if _, err := ResumeFileUpload(context.Background(), UploadOptions{SourcePath: filepath.Join(dir, "missing"), SourceIdentity: "source"}, func(context.Context, int64, int64, io.Reader) (int64, bool, error) {
		return 0, false, nil
	}); err == nil {
		t.Fatal("missing upload source returned nil error")
	}
	checkpointDir := filepath.Join(dir, "checkpoint-dir")
	if err := os.Mkdir(checkpointDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := ResumeFileUpload(context.Background(), UploadOptions{SourcePath: source, SourceIdentity: "source", CheckpointPath: checkpointDir}, func(context.Context, int64, int64, io.Reader) (int64, bool, error) {
		return 3, true, nil
	}); err == nil {
		t.Fatal("directory checkpoint path returned nil error")
	}
}

func TestUploadCheckpointHelpersHandleFilesystemErrors(t *testing.T) {
	checkpoint := uploadCheckpoint{Version: 1, SourcePathFingerprint: "source", SourceIdentityFingerprint: "identity", Total: 3}
	parent := filepath.Join(t.TempDir(), "parent")
	if err := os.WriteFile(parent, []byte("file"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := writeUploadCheckpoint(filepath.Join(parent, "checkpoint.json"), checkpoint); err == nil {
		t.Fatal("file upload checkpoint parent returned nil error")
	}
	valid := filepath.Join(t.TempDir(), "checkpoint.json")
	if err := writeUploadCheckpoint(valid, checkpoint); err != nil {
		t.Fatalf("writeUploadCheckpoint(valid): %v", err)
	}
	if _, err := os.ReadFile(valid); err != nil {
		t.Fatalf("upload checkpoint read: %v", err)
	}
}
