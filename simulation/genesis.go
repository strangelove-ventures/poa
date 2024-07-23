package simulation

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/strangelove-ventures/poa"
)

func RandomizedGenState(simState *module.SimulationState) {
	poaGenesis := poa.GenesisState{
		Vals: []poa.Validator{},
	}

	bz, err := json.MarshalIndent(&poaGenesis, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated poa genesis :\n%s\n", bz) // nolint: forbidigo
	simState.GenState[poa.ModuleName] = simState.Cdc.MustMarshalJSON(&poaGenesis)
}
