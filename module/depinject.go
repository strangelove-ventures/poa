package module

import (
	"os"

	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/store"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/strangelove-ventures/poa"
	"github.com/strangelove-ventures/poa/keeper"
	"github.com/strangelove-ventures/poa/simulation"

	modulev1 "github.com/strangelove-ventures/poa/api/module/v1"
)

var _ appmodule.AppModule = AppModule{}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

func init() {
	appmodule.Register(
		&modulev1.Module{},
		appmodule.Provide(ProvideModule),
	)
}

type ModuleInputs struct {
	depinject.In

	Cdc          codec.Codec
	StoreService store.KVStoreService
	AddressCodec address.Codec

	// TODO: Breaks depinject
	//StakingKeeper  stakingkeeper.Keeper
	SlashingKeeper keeper.SlashingKeeper
	BankKeeper     keeper.BankKeeper
}

type ModuleOutputs struct {
	depinject.Out

	Module appmodule.AppModule
	Keeper keeper.Keeper
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	// TODO: Breaks depinject
	k := keeper.NewKeeper(in.Cdc, in.StoreService /*&in.StakingKeeper,*/, in.SlashingKeeper, in.BankKeeper, log.NewLogger(os.Stderr))
	m := NewAppModule(in.Cdc, k)

	return ModuleOutputs{Module: m, Keeper: k, Out: depinject.Out{}}
}

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the slashing module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState)
}

// RegisterStoreDecoder registers a decoder for supply module's types
func (am AppModule) RegisterStoreDecoder(sdr simtypes.StoreDecoderRegistry) {
	sdr[poa.StoreKey] = simtypes.NewStoreDecoderFuncFromCollectionsSchema(am.keeper.Schema)
}

//// ProposalMsgs returns msgs used for governance proposals for simulations.
// func (AppModule) ProposalMsgs(simState module.SimulationState) []simtypes.WeightedProposalMsg {
//	return simulation.ProposalMsgs()
//}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return simulation.WeightedOperations(simState.AppParams, simState.Cdc, simState.TxConfig, am.keeper)
}
