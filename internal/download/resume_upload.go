package download

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// UploadSession sends one provider-supported chunk beginning at offset and
// returns the next acknowledged offset. A provider may return an error after
// acknowledging bytes; those bytes are checkpointed before the error returns.
type UploadSession func(context.Context, int64, int64, io.Reader) (nextOffset int64, done bool, err error)

// UploadOptions identifies the immutable local source for resumable upload.
type UploadOptions struct {
	SourcePath     string
	SourceIdentity string
	CheckpointPath string
}

// UploadResult reports provider acknowledgements and checkpoint cleanup.
type UploadResult struct {
	Bytes          int64  `json:"bytes"`
	Resumed        bool   `json:"resumed"`
	Completed      bool   `json:"completed"`
	CheckpointPath string `json:"checkpointPath,omitempty"`
}

type uploadCheckpoint struct {
	Version                   int    `json:"version"`
	SourcePathFingerprint     string `json:"sourcePathFingerprint"`
	SourceIdentityFingerprint string `json:"sourceIdentityFingerprint"`
	Total                     int64  `json:"total"`
	Completed                 int64  `json:"completed"`
}

// ResumeFileUpload coordinates a provider-supported resumable upload. It does
// not emulate resumability for one-shot providers: the session callback must
// acknowledge each provider chunk explicitly.
func ResumeFileUpload(ctx context.Context, opts UploadOptions, session UploadSession) (result UploadResult, err error) {
	if session == nil {
		return result, errors.New("upload session is required")
	}
	if opts.SourcePath == "" || opts.SourceIdentity == "" {
		return result, errors.New("upload source path and identity are required")
	}
	info, err := os.Stat(opts.SourcePath)
	if err != nil {
		return result, fmt.Errorf("stat upload source: %w", err)
	}
	if !info.Mode().IsRegular() {
		return result, fmt.Errorf("upload source %q is not a regular file", opts.SourcePath)
	}
	if opts.CheckpointPath == "" {
		opts.CheckpointPath = opts.SourcePath + ".upload.resume.json"
	}
	result.CheckpointPath = opts.CheckpointPath
	checkpoint, resumed, err := loadUploadCheckpoint(opts, info.Size())
	if err != nil {
		return result, err
	}
	if !resumed {
		checkpoint = uploadCheckpoint{Version: 1, SourcePathFingerprint: sourceFingerprint(opts.SourcePath), SourceIdentityFingerprint: sourceFingerprint(opts.SourceIdentity), Total: info.Size()}
		_ = os.Remove(opts.CheckpointPath)
	}
	result.Resumed = resumed
	if err := writeUploadCheckpoint(opts.CheckpointPath, checkpoint); err != nil {
		return result, err
	}

	f, err := os.Open(opts.SourcePath)
	if err != nil {
		return result, fmt.Errorf("open upload source: %w", err)
	}
	defer func() {
		if closeErr := f.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("close upload source: %w", closeErr)
		}
	}()

	for checkpoint.Completed < checkpoint.Total {
		if _, err := f.Seek(checkpoint.Completed, io.SeekStart); err != nil {
			return result, fmt.Errorf("seek upload source: %w", err)
		}
		previous := checkpoint.Completed
		next, done, sessionErr := session(ctx, previous, checkpoint.Total, f)
		if next < previous || next > checkpoint.Total {
			return result, fmt.Errorf("upload session acknowledged invalid offset %d", next)
		}
		checkpoint.Completed = next
		if checkpointErr := writeUploadCheckpoint(opts.CheckpointPath, checkpoint); checkpointErr != nil {
			return result, checkpointErr
		}
		result.Bytes = next
		if sessionErr != nil {
			return result, fmt.Errorf("resumable upload session: %w", sessionErr)
		}
		if done && next != checkpoint.Total {
			return result, fmt.Errorf("upload session marked complete at %d of %d bytes", next, checkpoint.Total)
		}
		if next == previous {
			return result, errors.New("upload session made no progress")
		}
	}
	if err := os.Remove(opts.CheckpointPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return result, fmt.Errorf("remove upload checkpoint: %w", err)
	}
	result.Completed = true
	result.CheckpointPath = ""
	return result, nil
}

func loadUploadCheckpoint(opts UploadOptions, total int64) (uploadCheckpoint, bool, error) {
	data, err := os.ReadFile(opts.CheckpointPath)
	if errors.Is(err, os.ErrNotExist) {
		return uploadCheckpoint{}, false, nil
	}
	if err != nil {
		return uploadCheckpoint{}, false, fmt.Errorf("read upload checkpoint: %w", err)
	}
	var checkpoint uploadCheckpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		_ = os.Remove(opts.CheckpointPath)
		return uploadCheckpoint{}, false, nil
	}
	valid := checkpoint.Version == 1 && checkpoint.SourcePathFingerprint == sourceFingerprint(opts.SourcePath) && checkpoint.SourceIdentityFingerprint == sourceFingerprint(opts.SourceIdentity) && checkpoint.Total == total && checkpoint.Completed >= 0 && checkpoint.Completed <= total
	if !valid {
		_ = os.Remove(opts.CheckpointPath)
		return uploadCheckpoint{}, false, nil
	}
	return checkpoint, checkpoint.Completed > 0, nil
}

func writeUploadCheckpoint(path string, checkpoint uploadCheckpoint) error {
	data, err := json.MarshalIndent(checkpoint, "", "  ")
	if err != nil {
		return fmt.Errorf("encode upload checkpoint: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create upload checkpoint directory: %w", err)
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".upload-resume-*.tmp")
	if err != nil {
		return fmt.Errorf("create upload checkpoint: %w", err)
	}
	tempName := temp.Name()
	defer func() { _ = os.Remove(tempName) }()
	if err := temp.Chmod(0o600); err != nil {
		_ = temp.Close()
		return fmt.Errorf("set upload checkpoint permissions: %w", err)
	}
	if _, err := temp.Write(data); err != nil {
		_ = temp.Close()
		return fmt.Errorf("write upload checkpoint: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close upload checkpoint: %w", err)
	}
	if err := os.Rename(tempName, path); err != nil {
		return fmt.Errorf("finalize upload checkpoint: %w", err)
	}
	return nil
}
