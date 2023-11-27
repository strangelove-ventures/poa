# Module Integration

## Table of Contents
* [Introduction](#introduction)
* [Example integration of the PoA Module](#example-integration-of-the-poa-module)
    * [Ante Handler Setup](#ante-handlers)
* [Design Details](#design-details)

# Introduction

This document provides the instructions on integration and configuring the Proof-of-Authority (poa) module within your Cosmos SDK chain implementation. This document makes the assumption that you have some existing codebase for your chain. If you do not, you can grab a template simapp from the [Cosmos SDK repo](https://github.com/cosmos/cosmos-sdk/tree/main/simapp). Validate your simapp version is on the same tagged version as this module is (eg. use v0.50.1 simapp for the v0.50.1 POA module).

As of the time of writing (Nov 2023) migrating a PoS (Proof of Stake) chains to PoA is not supported. This is possible, but the upgrade code has not yet been written to support this. If you are interested, please [create a PR](https://github.com/strangelove-ventures/poa/pulls).

The integration steps include the following:
1. Importing POA, setting the Module + Keeper, initialize the store keys + module params, and initialize the Begin/End Block logic and InitGenesis order.
2. Setup the Ante handler(s) to enforce the POA logic and give more control.


## Example integration of the PoA Module

```go
// app.go

// Import the poa module
import (
    ...

	"github.com/strangelove-ventures/poa"
	poatypes "github.com/strangelove-ventures/poa"
	poakeeper "github.com/strangelove-ventures/poa/keeper"
	poamodule "github.com/strangelove-ventures/poa/module"
)

...

// Set the PoA module Account permissions
var (
    maccPerms = map[string][]string{
		...
		poa.ModuleName:                 {authtypes.Minter},
	}
)


...

// Add PoA Keeper
type App struct {
	...
	POAKeeper    poakeeper.Keeper
	...
}


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

// Create store keys
keys := sdk.NewKVStoreKeys(
    ...
    poa.StoreKey,
    ...
)

// See the section below for configuring an application stack with the PoA

...

// Register PoA AppModule
app.moduleManager = module.NewManager(
    ...
    packetforward.NewAppModule(app.PacketForwardKeeper),
)

...

// Add PoA to begin blocker logic
// NOTE: This must be before the staking module begin blocker
app.moduleManager.SetOrderBeginBlockers(
    ...
    poa.ModuleName,
    stakingtypes.ModuleName,
    ...
)

// Add PoA to end blocker logic
app.moduleManager.SetOrderEndBlockers(
    ...
    poa.ModuleName,
    stakingtypes.ModuleName,
    ...
)

// Add PoA to init genesis logic
// NOTE: This must be after the staking module init genesis
app.moduleManager.SetOrderInitGenesis(
    ...
    stakingtypes.ModuleName,
    poa.ModuleName,
    ...
)
```

## Ante Handlers
- TODO:



----


## Design Details
- Wraps x/staking for XYZ
- delegation is forced from the validators account.
- validators can not unbond or redelegate. Can only go down 3 ways: (x/slashing module): downtime, doubleslash or (x/poa) admin removal.

- Ante: All staking commands except `MsgUpdateValidator`
- Ante: Commission limits (forced range, or specific value)

- If you want a module's control not to be based on governance, update your `app.go` authorities to use your set account instead of the gov account by default. This could be useful for the Upgrade module to not require governance but still allow the chain to get upgrades.