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
