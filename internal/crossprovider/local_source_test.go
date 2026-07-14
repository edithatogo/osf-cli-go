package crossprovider

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalSourceConfinesRegularFiles(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "data.txt"), []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}
	reader, err := (LocalSource{Root: root}).Open(t.Context(), File{Path: "data.txt"})
	if err != nil {
		t.Fatal(err)
	}
	content, readErr := io.ReadAll(reader)
	closeErr := reader.Close()
	if readErr != nil || closeErr != nil || string(content) != "data" {
		t.Fatalf("content=%q read=%v close=%v", content, readErr, closeErr)
	}
	for _, invalid := range []string{"../outside", "/absolute"} {
		if _, err := (LocalSource{Root: root}).Open(t.Context(), File{Path: invalid}); err == nil {
			t.Fatalf("Open(%q) returned nil error", invalid)
		}
	}
	symlink := filepath.Join(root, "link")
	if err := os.Symlink(filepath.Join(root, "data.txt"), symlink); err == nil {
		if _, err := (LocalSource{Root: root}).Open(t.Context(), File{Path: "link"}); err == nil {
			t.Fatal("Open symlink returned nil error")
		}
	}
}
