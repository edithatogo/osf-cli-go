package download

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

var errDestinationExists = errors.New("destination already exists")

// WriteStreamAtomically writes a single file to dst through a temporary file.
//
// It returns written=false when the destination exists and ConflictSkip is used.
func WriteStreamAtomically(dst string, src io.Reader, perm fs.FileMode, policy ConflictPolicy) (written bool, err error) {
	if err := policy.Validate(); err != nil {
		return false, err
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return false, fmt.Errorf("create destination directory: %w", err)
	}

	_, statErr := os.Stat(dst)
	exists := statErr == nil
	if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
		return false, fmt.Errorf("check destination: %w", statErr)
	}

	switch policy {
	case ConflictFail:
		if exists {
			return false, errDestinationExists
		}
	case ConflictSkip:
		if exists {
			return false, nil
		}
	}

	dir := filepath.Dir(dst)
	base := filepath.Base(dst)
	temp, err := os.CreateTemp(dir, "."+base+".*.tmp")
	if err != nil {
		return false, fmt.Errorf("create temporary file: %w", err)
	}

	tempName := temp.Name()
	cleanup := func() {
		_ = temp.Close()
		_ = os.Remove(tempName)
	}

	defer func() {
		if err != nil {
			cleanup()
		}
	}()

	if _, err = io.Copy(temp, src); err != nil {
		return false, fmt.Errorf("write temporary file: %w", err)
	}

	if err = temp.Chmod(perm); err != nil {
		return false, fmt.Errorf("set file permissions: %w", err)
	}
	if err = temp.Close(); err != nil {
		return false, fmt.Errorf("close temporary file: %w", err)
	}

	if policy == ConflictOverwrite && exists {
		if removeErr := os.Remove(dst); removeErr != nil {
			return false, fmt.Errorf("remove existing destination: %w", removeErr)
		}
	}

	if err = os.Rename(tempName, dst); err != nil {
		return false, fmt.Errorf("finalize download: %w", err)
	}

	return true, nil
}
