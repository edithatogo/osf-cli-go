// Package conformancetest provides reusable repository contract assertions.
package conformancetest

import (
	"encoding/json"
	"fmt"
	"slices"
	"testing"

	"github.com/edithatogo/osf-cli-go/internal/repository"
)

// Run verifies the invariants every concrete provider descriptor must satisfy.
func Run(t *testing.T, contract repository.Contract) {
	t.Helper()
	if err := validate(contract); err != nil {
		t.Fatalf("contract validation: %v", err)
	}
}

func validate(contract repository.Contract) error {
	if err := contract.Validate(); err != nil {
		return fmt.Errorf("contract validation: %w", err)
	}
	got := make([]repository.Capability, 0, len(contract.Capabilities))
	for _, detail := range contract.Capabilities {
		got = append(got, detail.Capability)
		if resolved := contract.Support(detail.Capability); resolved.Level != detail.Level {
			return fmt.Errorf("Support(%q) = %q, want %q", detail.Capability, resolved.Level, detail.Level)
		}
	}
	if !slices.Equal(got, repository.AllCapabilities()) {
		return fmt.Errorf("capabilities = %v, want complete vocabulary %v", got, repository.AllCapabilities())
	}
	encoded, err := json.Marshal(contract)
	if err != nil {
		return fmt.Errorf("marshal contract: %w", err)
	}
	var roundTrip repository.Contract
	if err := json.Unmarshal(encoded, &roundTrip); err != nil {
		return fmt.Errorf("unmarshal contract: %w", err)
	}
	if err := roundTrip.Validate(); err != nil {
		return fmt.Errorf("round-trip contract validation: %w", err)
	}
	return nil
}
