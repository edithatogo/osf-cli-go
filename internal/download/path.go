package download

import (
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

var errPathTraversal = errors.New("remote path escapes destination root")

// NormalizeDestination normalizes the user-selected destination directory.
func NormalizeDestination(dest string) (string, error) {
	dest = strings.TrimSpace(dest)
	if dest == "" {
		return "", fmt.Errorf("destination directory is required")
	}

	cleaned := filepath.Clean(dest)
	if cleaned == "." {
		return "", fmt.Errorf("destination directory is required")
	}

	abs, err := filepath.Abs(cleaned)
	if err != nil {
		return "", fmt.Errorf("normalize destination %q: %w", dest, err)
	}
	return abs, nil
}

// NormalizeRemotePath cleans an OSF remote path into a relative forward-slash path.
func NormalizeRemotePath(remote string) (string, error) {
	remote = strings.TrimSpace(remote)
	if remote == "" {
		return "", fmt.Errorf("remote path is required")
	}

	remote = strings.ReplaceAll(remote, "\\", "/")
	remote = strings.TrimPrefix(remote, "/")
	if remote == "" {
		return "", fmt.Errorf("remote path is required")
	}

	parts := strings.Split(remote, "/")
	cleaned := make([]string, 0, len(parts))
	for _, part := range parts {
		switch part {
		case "", ".":
			continue
		case "..":
			return "", errPathTraversal
		}
		cleaned = append(cleaned, part)
	}

	if len(cleaned) == 0 {
		return "", fmt.Errorf("remote path is required")
	}

	return path.Clean(strings.Join(cleaned, "/")), nil
}

// ResolveDestination joins the normalized destination root and remote path safely.
func ResolveDestination(destRoot, remote string) (string, error) {
	normalizedRoot, err := NormalizeDestination(destRoot)
	if err != nil {
		return "", err
	}

	normalizedRemote, err := NormalizeRemotePath(remote)
	if err != nil {
		return "", err
	}

	resolved := filepath.Join(normalizedRoot, filepath.FromSlash(normalizedRemote))
	ok, err := withinBase(normalizedRoot, resolved)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", errPathTraversal
	}

	return resolved, nil
}

func withinBase(base, target string) (bool, error) {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return false, fmt.Errorf("check destination path: %w", err)
	}
	if rel == "." {
		return true, nil
	}

	rel = filepath.Clean(rel)
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false, nil
	}
	return true, nil
}
