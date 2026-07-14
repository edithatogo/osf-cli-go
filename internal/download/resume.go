package download

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const resumeCheckpointVersion = 1

// ErrRangeUnsupported asks a resumable caller to restart from byte zero.
var ErrRangeUnsupported = errors.New("download range request was not honored")

// StreamOpener opens a remote stream positioned at offset. An implementation
// may restart from zero when the provider does not support range requests.
type StreamOpener func(offset int64) (io.ReadCloser, error)

// ResumeOptions controls a checkpointed atomic stream transfer.
type ResumeOptions struct {
	Destination      string
	Source           string
	ExpectedSize     *int64
	ExpectedChecksum string
	Policy           ConflictPolicy
	Perm             fs.FileMode
}

// ResumeResult reports the verified transfer and checkpoint state.
type ResumeResult struct {
	Bytes          int64  `json:"bytes"`
	Checksum       string `json:"checksum,omitempty"`
	Resumed        bool   `json:"resumed"`
	Completed      bool   `json:"completed"`
	CheckpointPath string `json:"checkpointPath,omitempty"`
}

type resumeCheckpoint struct {
	Version          int    `json:"version"`
	Source           string `json:"source"`
	Destination      string `json:"destination"`
	ExpectedSize     *int64 `json:"expectedSize,omitempty"`
	ExpectedChecksum string `json:"expectedChecksum,omitempty"`
	Completed        int64  `json:"completed"`
}

// ResumeStreamAtomically appends to a validated checkpoint and atomically
// finalizes the destination only after the stream size and checksum verify.
// Failed or cancelled transfers retain their .part and .resume.json files so
// the next invocation can continue from the recorded byte offset.
func ResumeStreamAtomically(open StreamOpener, opts ResumeOptions) (result ResumeResult, err error) {
	if open == nil {
		return result, errors.New("resume stream opener is required")
	}
	if strings.TrimSpace(opts.Destination) == "" {
		return result, errors.New("resume destination is required")
	}
	if strings.TrimSpace(opts.Source) == "" {
		return result, errors.New("resume source is required")
	}
	if err := opts.Policy.Validate(); err != nil {
		return result, err
	}
	if opts.Perm == 0 {
		opts.Perm = 0o644
	}

	destination := opts.Destination
	checkpointPath := resumeCheckpointPath(destination)
	partialPath := resumePartialPath(destination)
	result.CheckpointPath = checkpointPath

	if exists, statErr := pathExists(destination); statErr != nil {
		return result, fmt.Errorf("check resume destination: %w", statErr)
	} else if exists {
		switch opts.Policy {
		case ConflictFail:
			return result, errDestinationExists
		case ConflictSkip:
			result.Completed = true
			result.CheckpointPath = ""
			return result, nil
		}
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return result, fmt.Errorf("create resume destination directory: %w", err)
	}

	checkpoint, resumed, err := loadResumeCheckpoint(checkpointPath, partialPath, opts)
	if err != nil {
		return result, err
	}
	if !resumed {
		checkpoint = resumeCheckpoint{
			Version:          resumeCheckpointVersion,
			Source:           sourceFingerprint(opts.Source),
			Destination:      destination,
			ExpectedSize:     opts.ExpectedSize,
			ExpectedChecksum: opts.ExpectedChecksum,
		}
		_ = os.Remove(partialPath)
		_ = os.Remove(checkpointPath)
	}
	result.Resumed = resumed

	flags := os.O_CREATE | os.O_WRONLY
	if resumed {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}
	partial, err := os.OpenFile(partialPath, flags, opts.Perm)
	if err != nil {
		return result, fmt.Errorf("open resume partial: %w", err)
	}
	partialClosed := false
	defer func() {
		if !partialClosed {
			if closeErr := partial.Close(); err == nil && closeErr != nil {
				err = fmt.Errorf("close resume partial: %w", closeErr)
			}
		}
	}()

	if err := writeResumeCheckpoint(checkpointPath, checkpoint); err != nil {
		return result, err
	}
	src, err := open(checkpoint.Completed)
	if err != nil && checkpoint.Completed > 0 && errors.Is(err, ErrRangeUnsupported) {
		checkpoint.Completed = 0
		result.Resumed = false
		if truncateErr := partial.Truncate(0); truncateErr != nil {
			return result, fmt.Errorf("restart unsupported resume: %w", truncateErr)
		}
		if seekErr := func() error { _, seekErr := partial.Seek(0, io.SeekStart); return seekErr }(); seekErr != nil {
			return result, fmt.Errorf("rewind unsupported resume: %w", seekErr)
		}
		if checkpointErr := writeResumeCheckpoint(checkpointPath, checkpoint); checkpointErr != nil {
			return result, checkpointErr
		}
		src, err = open(0)
	}
	if err != nil {
		return result, fmt.Errorf("open resume source at %d: %w", checkpoint.Completed, err)
	}
	defer func() { _ = src.Close() }()

	buf := make([]byte, 32*1024)
	emptyReads := 0
	for {
		n, readErr := src.Read(buf)
		if n > 0 {
			emptyReads = 0
			written, writeErr := partial.Write(buf[:n])
			checkpoint.Completed += int64(written)
			if writeErr != nil {
				_ = writeResumeCheckpoint(checkpointPath, checkpoint)
				return result, fmt.Errorf("write resume partial: %w", writeErr)
			}
			if written != n {
				_ = writeResumeCheckpoint(checkpointPath, checkpoint)
				return result, io.ErrShortWrite
			}
			if err := writeResumeCheckpoint(checkpointPath, checkpoint); err != nil {
				return result, err
			}
		}
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				break
			}
			return result, fmt.Errorf("read resume source: %w", readErr)
		}
		emptyReads++
		if emptyReads >= 100 {
			return result, io.ErrNoProgress
		}
	}
	if err := partial.Sync(); err != nil {
		return result, fmt.Errorf("sync resume partial: %w", err)
	}
	if err := partial.Close(); err != nil {
		return result, fmt.Errorf("close resume partial: %w", err)
	}
	partialClosed = true

	if opts.ExpectedSize != nil && checkpoint.Completed != *opts.ExpectedSize {
		return result, fmt.Errorf("resume size mismatch: got %d bytes, want %d", checkpoint.Completed, *opts.ExpectedSize)
	}
	checksum, err := checksumFile(partialPath, opts.ExpectedChecksum)
	if err != nil {
		return result, err
	}
	result.Bytes = checkpoint.Completed
	result.Checksum = checksum
	if expected := expectedChecksumValue(opts.ExpectedChecksum); expected != "" && checksum != expected {
		return result, fmt.Errorf("resume checksum mismatch: got %s, want %s", checksum, expected)
	}

	if opts.Policy == ConflictOverwrite {
		if exists, statErr := pathExists(destination); statErr != nil {
			return result, fmt.Errorf("check destination before resume finalize: %w", statErr)
		} else if exists {
			if err := os.Remove(destination); err != nil {
				return result, fmt.Errorf("remove existing destination: %w", err)
			}
		}
	}
	if err := os.Rename(partialPath, destination); err != nil {
		return result, fmt.Errorf("finalize resumed transfer: %w", err)
	}
	if err := os.Remove(checkpointPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return result, fmt.Errorf("remove resume checkpoint: %w", err)
	}
	result.Completed = true
	result.CheckpointPath = ""
	return result, nil
}

func loadResumeCheckpoint(checkpointPath, partialPath string, opts ResumeOptions) (resumeCheckpoint, bool, error) {
	data, err := os.ReadFile(checkpointPath)
	if errors.Is(err, os.ErrNotExist) {
		return resumeCheckpoint{}, false, nil
	}
	if err != nil {
		return resumeCheckpoint{}, false, fmt.Errorf("read resume checkpoint: %w", err)
	}
	var checkpoint resumeCheckpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		_ = os.Remove(partialPath)
		_ = os.Remove(checkpointPath)
		return resumeCheckpoint{}, false, nil
	}
	valid := checkpoint.Version == resumeCheckpointVersion && checkpoint.Source == sourceFingerprint(opts.Source) && checkpoint.Destination == opts.Destination && equalInt64Ptr(checkpoint.ExpectedSize, opts.ExpectedSize) && checkpoint.ExpectedChecksum == opts.ExpectedChecksum
	partialInfo, statErr := os.Stat(partialPath)
	valid = valid && statErr == nil && partialInfo.Size() == checkpoint.Completed && checkpoint.Completed >= 0
	if !valid {
		_ = os.Remove(partialPath)
		_ = os.Remove(checkpointPath)
		return resumeCheckpoint{}, false, nil
	}
	return checkpoint, checkpoint.Completed > 0, nil
}

func writeResumeCheckpoint(path string, checkpoint resumeCheckpoint) error {
	data, err := json.MarshalIndent(checkpoint, "", "  ")
	if err != nil {
		return fmt.Errorf("encode resume checkpoint: %w", err)
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".resume-*.tmp")
	if err != nil {
		return fmt.Errorf("create resume checkpoint: %w", err)
	}
	tempName := temp.Name()
	defer func() { _ = os.Remove(tempName) }()
	if err := temp.Chmod(0o600); err != nil {
		_ = temp.Close()
		return fmt.Errorf("set resume checkpoint permissions: %w", err)
	}
	if _, err := temp.Write(data); err != nil {
		_ = temp.Close()
		return fmt.Errorf("write resume checkpoint: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close resume checkpoint: %w", err)
	}
	if err := os.Rename(tempName, path); err != nil {
		return fmt.Errorf("finalize resume checkpoint: %w", err)
	}
	return nil
}

func checksumFile(path, expected string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open resume checksum source: %w", err)
	}
	defer func() { _ = f.Close() }()
	h, label := checksumHash(expected)
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("calculate resume checksum: %w", err)
	}
	return label + ":" + hex.EncodeToString(h.Sum(nil)), nil
}

func checksumHash(expected string) (hash.Hash, string) {
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(expected)), "md5:") {
		return md5.New(), "md5"
	}
	return sha256.New(), "sha256"
}

func expectedChecksumValue(expected string) string {
	trimmed := strings.TrimSpace(expected)
	if trimmed == "" {
		return ""
	}
	if strings.Contains(trimmed, ":") {
		return strings.ToLower(trimmed)
	}
	return "sha256:" + strings.ToLower(trimmed)
}

func equalInt64Ptr(a, b *int64) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func resumePartialPath(destination string) string {
	return destination + ".part"
}

func resumeCheckpointPath(destination string) string {
	return destination + ".resume.json"
}

func sourceFingerprint(source string) string {
	sum := sha256.Sum256([]byte(source))
	return "sha256:" + hex.EncodeToString(sum[:])
}
