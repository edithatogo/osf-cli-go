package crossprovider

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// LocalSource confines validation inputs to regular, non-symlink files below Root.
type LocalSource struct {
	Root string
}

// Open implements SourceReader without allowing absolute paths or traversal.
func (source LocalSource) Open(_ context.Context, file File) (io.ReadCloser, error) {
	root, err := filepath.Abs(strings.TrimSpace(source.Root))
	if err != nil || strings.TrimSpace(source.Root) == "" {
		return nil, errors.New("cross-provider local source root is required")
	}
	clean := filepath.Clean(filepath.FromSlash(file.Path))
	if clean == "." || filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return nil, errors.New("cross-provider source path must remain below its root")
	}
	resolved := filepath.Join(root, clean)
	relative, err := filepath.Rel(root, resolved)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return nil, errors.New("cross-provider source path escapes its root")
	}
	info, err := os.Lstat(resolved)
	if err != nil {
		return nil, fmt.Errorf("inspect cross-provider source file: %w", err)
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return nil, errors.New("cross-provider source must be a regular non-symlink file")
	}
	return os.Open(resolved)
}
