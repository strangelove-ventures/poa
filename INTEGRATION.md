# Module Integration

## Table of Contents
* [Introduction](#introduction)
* [Example integration of the PoA Module](#example-integration-of-the-poa-module)
    * [Ante Handler Setup](#ante-handler-integration)
* [Design Details](#design-details)

# Introduction

This document provides the instructions on integration and configuring the Proof-of-Authority (PoA) module within your Cosmos SDK chain implementation. This document makes the assumption that you have some existing codebase for your chain. If you do not, you can grab a template simapp from the [Cosmos SDK repo](https://github.com/cosmos/cosmos-sdk/tree/main/simapp). Validate your simapp version is on the same tagged version as this module is (eg. use v0.50.1 simapp for the v0.50.1 POA module).

As of the time of writing (Nov 2023) migrating a PoS (Proof of Stake) chains to PoA is not supported. This is possible, but the upgrade code has not yet been written to support this. If you are interested, please [create a PR](https://github.com/strangelove-ventures/poa/pulls).

The integration steps include the following:
1. Importing POA, setting the Module + Keeper, initialize the store keys + module params, and initialize the Begin/End Block logic and InitGenesis order.
2. Setup the Ante handler(s) to enforce the POA logic and give more control.


## Example integration of the PoA Module

```go
// app.go

// Import the PoA module
import (
    ...
    "github.com/strangelove-ventures/poa"
	poatypes "github.com/strangelove-ventures/poa"
	poakeeper "github.com/strangelove-ventures/poa/keeper"
    poamodule "github.com/strangelove-ventures/poa/module"
)

...

// Add PoA Keeper
type App struct {
	...
	POAKeeper    poakeeper.Keeper
	...
}

...

// Create POA store key
keys := storetypes.NewKVStoreKeys(
    ...
    poa.StoreKey,
)

...

// Initialize the PoA Keeper and and AppModule
app.POAKeeper = poakeeper.NewKeeper(
    appCodec,
    runtime.NewKVStoreService(keys[poatypes.StoreKey]),
    app.StakingKeeper,
    app.SlashingKeeper,
    authcodec.NewBech32Codec(sdk.Bech32PrefixValAddr),
    logger,
)

...

// Register PoA AppModule
app.ModuleManager = module.NewManager(
    ...
    poamodule.NewAppModule(appCodec, app.POAKeeper),
)

...

// Add PoA to BeginBlock logic
// NOTE: This must be before the staking module begin blocker
app.ModuleManager.SetOrderBeginBlockers(
    ...
    poa.ModuleName,
    stakingtypes.ModuleName,
    ...
)

// Add PoA to end blocker logic
app.ModuleManager.SetOrderEndBlockers(
    ...
    poa.ModuleName,
    stakingtypes.ModuleName,
    ...
)

// Add PoA to init genesis logic
// NOTE: This must be after the staking module init genesis
app.ModuleManager.SetOrderInitGenesis(
    ...
    stakingtypes.ModuleName,
    poa.ModuleName,
    ...
)

// go get github.com/strangelove-ventures/poa

```

## Ante Handler Integration

### [Disable Staking](./ante/disable_staking.go)
A core feature of the POA module is to disable staking to all wallets. Make sure to add this decorator to your ante handler. An example can be found in the [simapp mock ante](./simapp/ante.go).

This blocks the following staking commands: Redelegate, Cancel Unbonding, Delegate, and Undelegate. MsgCreateValidator and UpdateParams are also blocked however the logic is wrapped in the PoA implementation & CLI. This also includes recursive authz ExecMsgs.

```go
import (
    ...
    poaante "github.com/strangelove-ventures/poa/ante"
)

...

func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error) {
    ...
    anteDecorators := []sdk.AnteDecorator{
        ...
        poaante.NewPOADisableStakingDecorator(),
        ...
    }
    ...
}
```

### [Commission Limits](./ante/commission_limit.go)
Depending on the chain use case, it may be desired to limit the commission rate range for min, max, or set value.

- `doGenTxRateValidation`: if true, genesis transactions also are required to be within the commission limit for the network.
- `rateFloor`: The minimum commission rate allowed. *(note: this must be higher than the StakingParams MinCommissionRate)*
- `rateCeil`: The maximum commission rate allowed.

if both rateFloor and rateCiel are set to the same value, then the commission rate is forced to that value.

```go
import (
    ...
    sdkmath "cosmossdk.io/math"
    poaante "github.com/strangelove-ventures/poa/ante"
)

...

func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error) {
    ...

    doGenTxRateValidation := false
    rateFloor := sdkmath.LegacyMustNewDecFromStr("0.10")
	rateCeil := sdkmath.LegacyMustNewDecFromStr("0.50")

    anteDecorators := []sdk.AnteDecorator{
        ...
        poaante.CommissionLimitDecorator(doGenTxRateValidation, rateFloor, rateCeil),
        ...
    }
    ...
}
```


----


## Design Details
- Wraps x/staking for XYZ
- delegation is forced from the validators account.
- validators can not unbond or redelegate. Can only go down 3 ways: (x/slashing module): downtime, doubleslash or (x/poa) admin removal.

- Ante: All staking commands except `MsgUpdateValidator`
- Ante: Commission limits (forced range, or specific value)

- If you want a module's control not to be based on governance, update your `app.go` authorities to use your set account instead of the gov account by default. This could be useful for the Upgrade module to not require governance but still allow the chain to get upgrades.