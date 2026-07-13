package main

import (
	"strings"
	"testing"
)

func TestCheckPackageManifestRequiresSensitiveCredentials(t *testing.T) {
	manifest := packageManifest{Version: "0.3.2"}
	for _, name := range expectedTools {
		manifest.Tools = append(manifest.Tools, struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}{Name: name, Description: "tool"})
	}
	manifest.UserConfig = map[string]struct {
		Sensitive bool `json:"sensitive"`
	}{
		"osf_token": {Sensitive: true}, "osf_username": {Sensitive: true}, "osf_password": {Sensitive: false},
	}
	if err := checkPackageManifest(manifest, "0.3.2"); err == nil || !strings.Contains(err.Error(), "osf_password") {
		t.Fatalf("checkPackageManifest() error = %v, want sensitive password failure", err)
	}
}

func TestSameNamesIgnoresOrderButNotInventory(t *testing.T) {
	if !sameNames([]string{"b", "a"}, []string{"a", "b"}) {
		t.Fatal("sameNames returned false for equivalent inventories")
	}
	if sameNames([]string{"a"}, []string{"a", "b"}) {
		t.Fatal("sameNames returned true for incomplete inventory")
	}
}
