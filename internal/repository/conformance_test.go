package repository_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/edithatogo/osf-cli-go/internal/repository"
	"github.com/edithatogo/osf-cli-go/internal/repository/conformancetest"
)

func TestConcreteProviderConformance(t *testing.T) {
	for _, contract := range []repository.Contract{repository.OSFContract(), repository.ZenodoContract()} {
		t.Run(string(contract.Provider), func(t *testing.T) {
			conformancetest.Run(t, contract)
		})
	}
}

func TestCapabilityFixture(t *testing.T) {
	data, err := os.ReadFile("testdata/capability-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		SchemaVersion int `json:"schema_version"`
		Cases         []struct {
			Provider   repository.Provider     `json:"provider"`
			Capability repository.Capability   `json:"capability"`
			Level      repository.SupportLevel `json:"level"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SchemaVersion != 1 || len(fixture.Cases) == 0 {
		t.Fatalf("invalid fixture schema or empty cases")
	}
	contracts := map[repository.Provider]repository.Contract{
		repository.ProviderOSF:    repository.OSFContract(),
		repository.ProviderZenodo: repository.ZenodoContract(),
	}
	seenLevels := map[repository.SupportLevel]bool{}
	for _, testCase := range fixture.Cases {
		contract, ok := contracts[testCase.Provider]
		if !ok {
			t.Errorf("unknown fixture provider %q", testCase.Provider)
			continue
		}
		if got := contract.Support(testCase.Capability).Level; got != testCase.Level {
			t.Errorf("%s %s = %q, want %q", testCase.Provider, testCase.Capability, got, testCase.Level)
		}
		seenLevels[testCase.Level] = true
	}
	for _, level := range []repository.SupportLevel{repository.SupportSupported, repository.SupportPartial, repository.SupportUnsupported} {
		if !seenLevels[level] {
			t.Errorf("fixture does not exercise %q support", level)
		}
	}
}
