package conformancetest

import (
	"testing"

	"github.com/edithatogo/osf-cli-go/internal/repository"
)

func TestRunConcreteContracts(t *testing.T) {
	for _, contract := range []repository.Contract{repository.OSFContract(), repository.ZenodoContract()} {
		t.Run(string(contract.Provider), func(t *testing.T) { Run(t, contract) })
	}
}

func TestValidateRejectsIncompleteContracts(t *testing.T) {
	contract := repository.OSFContract()
	contract.Capabilities = contract.Capabilities[:1]
	if err := validate(contract); err == nil {
		t.Fatal("incomplete contract validated successfully")
	}
}

func TestValidateRejectsMismatchedSupportLevel(t *testing.T) {
	contract := repository.OSFContract()
	contract.Capabilities[0].Level = "unsupported-level"
	if err := validate(contract); err == nil {
		t.Fatal("mismatched support level validated successfully")
	}
}
