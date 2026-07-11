// Command checksums writes SHA-256 checksums for files in a directory.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: checksums <directory> <output>")
		os.Exit(2)
	}
	entries, err := os.ReadDir(os.Args[1])
	if err != nil {
		fail(err)
	}
	var names []string
	for _, entry := range entries {
		if !entry.IsDir() && entry.Name() != filepath.Base(os.Args[2]) {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	out, err := os.Create(os.Args[2])
	if err != nil {
		fail(err)
	}
	defer func() {
		if err := out.Close(); err != nil {
			fail(err)
		}
	}()
	for _, name := range names {
		file, err := os.Open(filepath.Join(os.Args[1], name))
		if err != nil {
			fail(err)
		}
		hash := sha256.New()
		if _, err := io.Copy(hash, file); err != nil {
			_ = file.Close()
			fail(err)
		}
		if err := file.Close(); err != nil {
			fail(err)
		}
		if _, err := fmt.Fprintf(out, "%s  %s\n", hex.EncodeToString(hash.Sum(nil)), name); err != nil {
			fail(err)
		}
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
