package module

import (
	"os"

	"github.com/cosmos/cosmos-sdk/codec"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/store"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"

	modulev1 "github.com/strangelove-ventures/poa/api/module/v1"
	"github.com/strangelove-ventures/poa/keeper"
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

	StakingKeeper  stakingkeeper.Keeper
	SlashingKeeper slashingkeeper.Keeper
	BankKeeper     bankkeeper.Keeper

	Authority string
}

type ModuleOutputs struct {
	depinject.Out

	Module appmodule.AppModule
	Keeper keeper.Keeper
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	k := keeper.NewKeeper(in.Cdc, in.StoreService, &in.StakingKeeper, in.SlashingKeeper, in.BankKeeper, log.NewLogger(os.Stderr), in.Authority)
	m := NewAppModule(in.Cdc, k)

	return ModuleOutputs{Module: m, Keeper: k, Out: depinject.Out{}}
}
