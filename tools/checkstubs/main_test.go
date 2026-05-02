package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScanFileReportsProductionMarkers(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "production.go")
	if err := os.WriteFile(path, []byte("package main\n\n// TODO: replace this\n"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	findings, err := scanFile(path)
	if err != nil {
		t.Fatalf("scanFile: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d (%v)", len(findings), findings)
	}
	if got := findings[0]; !strings.Contains(got, "production.go:3") || !strings.Contains(got, `contains "TODO:"`) {
		t.Fatalf("unexpected finding: %q", got)
	}
}

func TestIgnoredTestAndFixturePaths(t *testing.T) {
	cases := []struct {
		name string
		path string
		want bool
	}{
		{name: "test file", path: "internal/cli/command_test.go", want: false},
		{name: "root testdata", path: "testdata", want: true},
		{name: "nested testdata", path: filepath.Join("internal", "cli", "testdata"), want: true},
		{name: "root fixtures", path: "fixtures", want: true},
		{name: "nested fixtures", path: filepath.Join("internal", "cli", "fixtures"), want: true},
		{name: "normal dir", path: "internal/cli", want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := shouldSkipDir(tc.path); got != tc.want {
				t.Fatalf("shouldSkipDir(%q) = %v, want %v", tc.path, got, tc.want)
			}
		})
	}
}

func TestProductionGoExcludesToolAndTests(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{path: "internal/cli/root.go", want: true},
		{path: "internal/cli/root_test.go", want: false},
		{path: "tools/checkstubs/main.go", want: false},
		{path: "tools/checkstubs/helper.go", want: false},
	}

	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			if got := isProductionGo(tc.path); got != tc.want {
				t.Fatalf("isProductionGo(%q) = %v, want %v", tc.path, got, tc.want)
			}
		})
	}
}
