# `PoA` Module

The Proof of Authority (PoA) module allows for permissioned networks to be controlled by a predefined set of validators to verify transactions. This implementation extends the Cosmos-SDK's x/staking module to a a set of administrators over the chain. These administrators gate keep the chain by whitelisting validators, updating cosnensus power, and removing validators from the network.

Since this module depends on x/staking, carefully read through the [Integration Guide](./INTEGRATION.md) for a full understanding of how the two modules interact. This design choice was made to allow for the PoA module backwards compatible with website UIs, bots, and validator experience. It also has the added security benefit of a tested core module in Cosmos.

## Concepts

The PoA flow is divided into a few key steps:
- **GenTx**: Validators submit standard genesis transactions with the app binary CLI wrapper
- **Start**: The chain controller merges these genesis transactions into the genesis file and starts the chain
- **Updates**: The chain admin(s) can update the validator set by adding validators or modifying their consensus power.
- **Removal**: The chain admin(s) can remove validators from the network.

## State

### Genesis
`Params` is found in the genesis state. This object contains the `admins` field, which is a list of bech32 addresses that are allowed to modify the validator set. If none are set, then the chain is controlled by governance itself. By allowing for an array of admins, the chain can be controlled by multiple different parties according to your use case. For example, governance can be an admin while also allowing a multisig, DAO, and/or single account to control the validator set.

Only the admins can update the params. This includes the admins themselves. There must be at least one admin in the list at all times. If you do not want a single account to control the set, set it to the chain's governance address.

### Pending Validators
`PendingValidators` stores an array of PoA validator objects pending approval (from the admins) into the active set. This only is required after the chain has started.

For better UX, this is accomplished by wrapping the x/staking module's `create-validator` command with our own logic. Validators only have to modify the namespace of their create command (from `tx staking create-validator` -> `tx poa create-validator`) with all else being equal.

### Previous Block Power
`CachedPreviousBlockPower` saves the previous blocks total consensus power amount for quieries at Height + 1. It allows for safety checks onm updating too much of the sets power resulting in broken IBC connections. It's protection can be passed by using the `--unsafe` flag in the `set-power` CLI command.

### Absolute Changed Block Power
`AbsoluteChangedInBlockPower` tracks the per block modification in block power. It follows the absolute power difference a single block can change against the previous block power. Attempting to increase more than 30% of the validator set power (relative to last) will error.

**Flow**:
- Validator previous block set power is 9 (3 validators @ 3 power)
- The admin increases validator[0] to 4 power (+11%)
- The admin increases validator[1] to 4 power (+22%)
- The admin increases validator[2] to 4 power (+33%, error)

The `AbsoluteChangedPower` of +1 to each validator is 3, which is 33% of the previous block power (3/9). It can be bypassed with the use of the `--unsafe` flag in the CLI command.


## Messages

TODO: ...

The following messages can only be submitted from the chain admin(s).

### CreateValidator

### SetPower

### RemoveValidator

### UpdateParams

### UpdateStakingParams

## [Begin Block](./module/abci.go)

As the PoA logic is dependent on the x/staking module, the PoA module must be run before the x/staking modules `BeginBlock` logic. This is described in the [integration guide](./INTEGRATION.md).

When removing validators, the validator can not be instantly removed from the set and it required a few intermediate blocks.

**Flow**
A validator is removed by an admin at height H; it increases the minimum self delegation to current+1. This puts the bonded validator into `Unbonding` status when checked by the x/staking module in it's BeginBlock (at H). The next iteration (H+1) the PoA module force updates the `Unbonding` validator to the status of `Unbonded`. The x/staking module then performs it's BeginBlock logic and sets it as `Unbonded`. The validator is now deleted from consensus at H+2 in the PoA BeginBlock.


## Hooks

Hooks are called as would be expected from the x/staking hooks.

e.g. `AfterValidatorCreated` is called after a validator is accepted into the set, not just created into `PendingValidators`.

## Events

Events are broadcast as expected from the x/staking events.

## Client

TODO: REST / gRPC Endpoints.

### CLI

A user can query and interact with the `poa` module using the CLI.

### Query
The `query` commands allow users to query the `poa` state.

```bash
# Get module params (admins list)
poad q poa params

# Get validators waiting to be added to the set
poad q poa pending-validators

# Get the current consensus power of a specific validator
poad q poa power [validator]
```

To get validator specific information such as commission rates, details, etc. use the x/staking module's query commands.

```bash
# e.g. validator, validators, params
poad q staking --help
```

### Transactions
The `tx` commands allow users to interact and update the `poa` state.

```bash
# Create a validator and add it to the pending set
poad tx poa create-validator path/to/validator.json --from keyname

# (admin) Remove a validator from the set and delete them
poad tx poa remove [validator]

# (admin) Modify the consensus power of a validator
# - validator is the bech32 address of the validator operator
# - amount uses 10^6 precision (1,000,000 = 1 power)
# - --unsafe flag allows for bypassing the 30% max change per block
poad tx poa set-power [validator] [amount] [--unsafe]

# (admin) Update the module params (admins list)
# - admins is a comma separated list of bech32 addresses
# - any admin can modify the list at any time
# - there must be at least one admin in the list at all times
poad tx poa update-params admin1,admin2,admin3,...

# (admin) Update the staking module params
# - unbondingTime is the time in seconds that a validator must wait to unbond
# - maxVals is the maximum number of validators allowed in the set
# - maxEntries is the maximum number of unbonding delegations per validator (this does not apply for PoA)
# - historicalEntries is the maximum number of historical entries stored per validator (this does not apply for PoA)
# - bondDenom is the denom of the bond (e.g. uatom)
# - minCommissionRate is the minimum commission rate a validator can set
poad tx poa update-staking-params [unbondingTime] [maxVals] [maxEntries] [historicalEntries] [bondDenom] [minCommissionRate]
```

## Appendix

- [Cosmos-SDK Staking Module](https://github.com/cosmos/cosmos-sdk/tree/main/x/staking)