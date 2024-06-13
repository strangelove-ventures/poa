package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/strangelove-ventures/poa"
)

const (
	admins           = "admins"
	allowValSelfExit = "allow_validator_self_exit"
)

// TODO: Fix randomization
func genAdmins(r *rand.Rand) []string {
	govModuleAddress := authtypes.NewModuleAddress(govtypes.ModuleName).String()
	return []string{govModuleAddress}
}

func RandomizedGenState(simState *module.SimulationState) {
	var (
		adm           []string
		allowSelfExit bool
	)

	// The POA admin is the governance module address N% of the time
	// Allow validator self exit is enabled 50% of the time

	simState.AppParams.GetOrGenerate(admins, &adm, simState.Rand, func(r *rand.Rand) { adm = genAdmins(r) })
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
	fmt.Printf("Selected randomly generated poa parameters:\n%s\n", bz)
	simState.GenState[poa.ModuleName] = simState.Cdc.MustMarshalJSON(&poaGenesis)
}
