package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var markers = []string{
	"panic(\"TODO\")",
	"panic(\"todo\")",
	"panic(\"not implemented\")",
	"not implemented",
	"TODO:",
	"FIXME:",
	"STUB:",
	"PLACEHOLDER:",
}

func main() {
	var findings []string

	err := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if shouldSkipDir(path) {
				return filepath.SkipDir
			}
			return nil
		}
		if !isProductionGo(path) {
			return nil
		}
		fileFindings, scanErr := scanFile(path)
		if scanErr != nil {
			return scanErr
		}
		findings = append(findings, fileFindings...)
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "checkstubs: %v\n", err)
		os.Exit(1)
	}

	if len(findings) > 0 {
		fmt.Fprintln(os.Stderr, "production code contains incomplete-work markers:")
		for _, finding := range findings {
			fmt.Fprintln(os.Stderr, "  "+finding)
		}
		os.Exit(1)
	}
}

func shouldSkipDir(path string) bool {
	clean := filepath.ToSlash(path)
	switch clean {
	case ".git", ".gocache", ".gomodcache":
		return true
	}
	for _, part := range strings.Split(clean, "/") {
		switch part {
		case "testdata", "fixtures":
			return true
		}
	}
	return false
}

func isProductionGo(path string) bool {
	clean := filepath.ToSlash(path)
	return strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") && !strings.HasPrefix(clean, "tools/checkstubs/")
}

func scanFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var findings []string
	scanner := bufio.NewScanner(f)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		for _, marker := range markers {
			if strings.Contains(line, marker) {
				findings = append(findings, fmt.Sprintf("%s:%d contains %q", filepath.ToSlash(path), lineNumber, marker))
			}
		}
	}
	return findings, scanner.Err()
}
