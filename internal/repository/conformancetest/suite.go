// Package conformancetest provides reusable repository contract assertions.
package conformancetest

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/edithatogo/osf-cli-go/internal/repository"
)

// Run verifies the invariants every concrete provider descriptor must satisfy.
func Run(t *testing.T, contract repository.Contract) {
	t.Helper()
	if err := contract.Validate(); err != nil {
		t.Fatalf("contract validation: %v", err)
	}
	got := make([]repository.Capability, 0, len(contract.Capabilities))
	for _, detail := range contract.Capabilities {
		got = append(got, detail.Capability)
		if resolved := contract.Support(detail.Capability); resolved.Level != detail.Level {
			t.Errorf("Support(%q) = %q, want %q", detail.Capability, resolved.Level, detail.Level)
		}
	}
	if !slices.Equal(got, repository.AllCapabilities()) {
		t.Errorf("capabilities = %v, want complete vocabulary %v", got, repository.AllCapabilities())
	}
	encoded, err := json.Marshal(contract)
	if err != nil {
		t.Fatalf("marshal contract: %v", err)
	}
	var roundTrip repository.Contract
	if err := json.Unmarshal(encoded, &roundTrip); err != nil {
		t.Fatalf("unmarshal contract: %v", err)
	}
	if err := roundTrip.Validate(); err != nil {
		t.Fatalf("round-trip contract validation: %v", err)
	}
}
