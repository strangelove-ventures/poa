# Module Integration

## Table of Contents
* [Introduction](#introduction)
* [Example integration of the PoA Module](#example-integration-of-the-poa-module)
    * [Ante Handler Setup](#ante-handler-integration)
* [Network Considerations](#network-considerations)

# Introduction

This document provides the instructions on integration and configuring the Proof-of-Authority (PoA) module within your Cosmos SDK chain implementation. This document makes the assumption that you have some existing codebase for your chain. If you do not, you can grab a template simapp from the [Cosmos SDK repo](https://github.com/cosmos/cosmos-sdk/tree/main/simapp). Validate your app version is on the same tagged version as this module (eg. use v0.50.1 simapp for the v0.50.1 PoA module).

As of the time of writing (Nov 2023) migrating a PoS (Proof of Stake) chains to PoA is not yet supported. This is possible, but the upgrade code has not yet been written yet. If you are interested, please [create a PR](https://github.com/strangelove-ventures/poa/pulls).

The integration steps include the following:
1. Importing POA, setting the Module + Keeper, initialize the store keys, and initialize the Begin/End Block logic and InitGenesis order.
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

// Create PoA store key
keys := storetypes.NewKVStoreKeys(
    ...
    poa.StoreKey,
)

...

// Initialize the PoA Keeper and AppModule
app.POAKeeper = poakeeper.NewKeeper(
    appCodec,
    runtime.NewKVStoreService(keys[poatypes.StoreKey]),
    app.StakingKeeper,
    app.SlashingKeeper,
    app.BankKeeper,
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
    ...
)

// Add PoA to end blocker logic
app.ModuleManager.SetOrderEndBlockers(
    ...
    poa.ModuleName,
    ...
)

// Add PoA to init genesis logic
// NOTE: This must be after the staking module init genesis
app.ModuleManager.SetOrderInitGenesis(
    ...
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

## Network Considerations

### Slashing - Genesis Params

When setting up your network genesis, it is important to consider setting slash infractions to 0%. Setting downtime is more reasonable to diminish their weight in the network.

```json
app_state.slashing.params.slash_fraction_double_sign
0.000000000000000000

app_state.slashing.params.slash_fraction_downtime
0.000000000000000000
```

It is up to you on setting the slashing window requirements. An Example:

```json
app_state.slashing.params.signed_blocks_window
10000

app_state.slashing.params.min_signed_per_window
0.100000000000000000

app_state.slashing.params.downtime_jail_duration
600s
```

## PoA - Genesis params

The PoA module can be configured to pre-define the admin controls. One admin must be set for a valid network. If you do not yet have an admin, set the chain's governance address. (CLI: `appd q auth module-account gov`)

```json
// base + module accounts, DAOs, and multisigs are supported.
app_state.poa.params.admins
["admin1...", "admin2..."]
```

## Staking - Genesis Params

You must modify the genesis staking parameters for some other PoA configuration options.

```json
// The maximum size of the active set.
app_state.staking.params.max_validators
100

// Any unique token can be used here.
app_state.staking.params.bond_denom
"upoa"

// 0 is recommended here if you use the `CommissionLimitDecorator` ante.
app_state.staking.params.min_commission_rate
0.000000000000000000
```

The bond_denom is discussed in the next section.

### Tokens

Because all network staking is disabled, it grants you the ability to use any token as the power token. This also includes using the PoA power token as a standard user token in the network. *(e.g. validator power can be `token` while the network mints and uses `token` also to distribute to user).*

For best practice, use a dedicated power token (e.g. upoa) for validators and another token(s) for the day to day network and gas fee operations.

### Validator Commission Rates

Force your validator set commission rate range with the [Forced Commission Rate Ante Handler](./INTEGRATION.md#ante-handler-integration).

### Full Module Control

If you want a module's control not to be based on governance (e.g. x/upgrade for software upgrades), update that module's app.go authority string to use your own account instead of the gov address `authtypes.NewModuleAddress(govtypes.ModuleName).String()`. This can be one of the accounts in the PoA admin set, or any other valid account on chain (e.g. a multisig, DAO, Base or Module account).

## Migrating to PoS from PoA

A future optional upgrade will grant PoA networks the ability to migrate to PoS (Proof-of-Stake).

Reasons this may be desired:
- The chain product has been successful and the network is ready to be decentralized.
- There is a new token use case that requires a PoS network for user delegations (ex: sharing platform rewards with stakers).
