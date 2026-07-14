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
