package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/strangelove-ventures/poa"
)

const (
	admins           = "admins"
	allowValSelfExit = "allow_validator_self_exit"
)

func RandomizedGenState(simState *module.SimulationState) {
	var (
		adm           []string
		allowSelfExit bool
	)

	simState.AppParams.GetOrGenerate(admins, &adm, simState.Rand, func(r *rand.Rand) {
		// Select a random number of admins from the simState accounts
		adminSet := make(map[string]bool)
		numAdmins := simulation.RandIntBetween(r, 1, len(simState.Accounts))
		for i := 0; i < numAdmins; i++ {
			acc, _ := simulation.RandomAcc(r, simState.Accounts)
			adminSet[acc.Address.String()] = true
		}

		for k := range adminSet {
			adm = append(adm, k)
		}
	})

	// Allow validator self exit is enabled 50% of the time
	simState.AppParams.GetOrGenerate(allowValSelfExit, &allowSelfExit, simState.Rand, func(r *rand.Rand) { allowSelfExit = r.Intn(2) == 1 })

	poaGenesis := poa.GenesisState{
		Params: poa.Params{
			Admins:                 adm,
			AllowValidatorSelfExit: allowSelfExit,
		},
	}

	bz, err := json.MarshalIndent(&poaGenesis.Params, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated poa parameters:\n%s\n", bz) // nolint: forbidigo
	simState.GenState[poa.ModuleName] = simState.Cdc.MustMarshalJSON(&poaGenesis)
}
