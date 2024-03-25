# `PoA` Module

The Proof of Authority (PoA) module allows for permissioned networks to be controlled by a predefined set of validators to verify transactions. This implementation extends the Cosmos-SDK's x/staking module to a a set of administrators over the chain. These administrators gate keep the chain by whitelisting validators, updating cosnensus power, and removing validators from the network.

## Integration

Since this module depends on x/staking, carefully read through the [Integration Guide](./INTEGRATION.md) to add it to your network. This design choice was made to allow for the PoA module to have backwards compatibility with:
- Website UIs
- Uptime bots
- Validator scripts
- in-process multi-validator testnets

## Configuration

After integrating the PoA module into your chain, read the [network considerations](./INTEGRATION.md#network-considerations) before launching the network.

This includes: parameters, full module control, migrating from PoA->PoS, and other useful information.

## Concepts

The PoA flow is divided into a few key steps:
- **GenTx**: Validators submit standard genesis transactions with the app binary CLI wrapper
- **Start**: The chain controller merges these genesis transactions into the genesis file and starts the chain
- **Updates**: The chain admin(s) can update the validator set by adding validators or modifying their consensus power.
- **Removal**: The chain admin(s) can remove validators from the network.

All delegation actions are disabled with the [disabled-staking ante](./INTEGRATION.md#ante-handler-integration). While the validator is the set delegator of its own account, only the admin(s) can modify this delegation of power via x/poa. Validators can only go down the following ways:

| Module	    | Action 	      |
|---	        |---	          |
| x/slashing  | downtime 	    |
| x/slashing  | double sign   |
| x/poa       | admin removal |
| x/poa       | self removal  |

## State

### Genesis
`Params` is found in the genesis state. This object contains the `admins` field, which is a list of bech32 addresses that are allowed to modify the validator set, PoA params, and validator params (via x/staking). If no admin is set, then the chain should set the governance account as the admin (default). By allowing for an array of admins, the chain can be controlled by multiple different parties according to your use case. For example, governance can be an admin while also allowing a multisig, DAO, and/or single account to control the validator set. There must be at least one admin in the list at all times.

`AllowValidatorSelfExit` is a boolean option *(default true)* that toggles validators ability to remove themselves from the active set. This is useful for validators that are no longer able or decide this PoA network is no longer for them. If set to false, only admins can remove validators from the set. **NOTE** This does not stop validators from going down and being jailed for downtime. It just provides a graceful way to self remove power.

### Pending Validators
`PendingValidators` stores an array of PoA validator objects pending approval (from the admins) into the active set. This only is required after the chain has started.

For better UX, this is accomplished by wrapping the x/staking module's `create-validator` command with our own logic. Validators only have to modify the namespace of their create command (from `tx staking create-validator` -> `tx poa create-validator`) with all else being equal.

### Previous Block Power
`CachedPreviousBlockPower` saves the previous blocks total consensus power amount for queries at Height + 1. It allows for safety checks onm updating too much of the sets power resulting in broken IBC connections. It's protection can be passed by using the `--unsafe` flag in the `set-power` CLI command.

### Absolute Changed Block Power
`AbsoluteChangedInBlockPower` tracks the per block modification in block power. It follows the absolute power difference a single block can change against the previous block power. Attempting to increase more than 30% of the validator set power (relative to last) will error.

**Flow**:
- Validator previous block set power is 9 (3 validators @ 3 power)
- The admin increases validator[0] to 4 power (+11%)
- The admin increases validator[1] to 4 power (+22%)
- The admin increases validator[2] to 4 power (+33%, error)

The `AbsoluteChangedPower` of +1 to each validator is 3, which is 33% of the previous block power (3/9). It can be bypassed with the use of the `--unsafe` flag in the CLI command.


## Messages

### CreateValidator
```json
{
  "@type": "/strangelove_ventures.poa.v1.MsgCreateValidator",
  "description": {
    "moniker": "Validator Name",
    "identity": "",
    "website": "https://website.com",
    "security_contact": "security@cosmos.xyz",
    "details": "description"
  },
  "commission": {
    "rate": "0.100000000000000000",
    "max_rate": "0.200000000000000000",
    "max_change_rate": "0.010000000000000000"
  },
  "min_self_delegation": "1",
  "delegator_address": "",
  "validator_address": "cosmosvaloper1addr",
  "pubkey": {
    "@type": "/cosmos.crypto.ed25519.PubKey",
    "key": "pl3Q8OQwtC7G2dSqRqsUrO5VZul7l40I+MKUcejqRsg="
  }
}
```

### SetPower (admin only)

- `power` is a micro unit of power (1,000,000 = 1 power) to derive a validators consensus power.
- `unsafe` allows an admin to bypass the 30% of consensus power per block limitation.

```json
{
  "@type": "/strangelove_ventures.poa.v1.MsgSetPower",
  "sender": "cosmos1addr",
  "validator_address": "cosmosvaloper1addr",
  "power": "12356789",
  "unsafe": true
}
```

### RemoveValidator (admin only)
```json
{
  "@type": "/strangelove_ventures.poa.v1.MsgRemoveValidator",
  "sender": "cosmos1addr",
  "validator_address": "cosmosvaloper1addr"
}
```

### RemovePending (admin only)
```json
{
  "@type": "/strangelove_ventures.poa.v1.MsgRemovePending",
  "sender": "cosmos1hj5fveer5cjtn4wd6wstzugjfdxzl0xpxvjjvr",
  "validator_address": "cosmosvaloper1efd63aw40lxf3n4mhf7dzhjkr453axurlv4rfe"
}
```

### UpdateParams (admin only)

- `admins` is a list of bech32 addresses that control the validator set and parameters of the chain.
- `allow_validator_self_exit` allows validators to force remove themselves from the active set at any time.

```json
{
  "@type": "/strangelove_ventures.poa.v1.MsgUpdateParams",
  "sender": "cosmos1addr",
  "params": {
    "admins": [
      "cosmos1addr",
      "cosmos1addr1",
      "cosmos1addr2",
    ],
    "allow_validator_self_exit": true
  }
}
```

### UpdateStakingParams (admin only)
```json
{
  "@type": "/strangelove_ventures.poa.v1.MsgUpdateStakingParams",
  "sender": "cosmos1hj5fveer5cjtn4wd6wstzugjfdxzl0xpxvjjvr",
  "params": {
    "unbonding_time": "1209600s",
    "max_validators": 100,
    "max_entries": 7,
    "historical_entries": 7,
    "bond_denom": "stake",
    "min_commission_rate": "0.000000000000000000"
  }
}
```

## [Begin Block](./module/abci.go)

As the PoA logic is dependent on the x/staking module, the PoA module must be run before the x/staking modules `BeginBlock` logic. This is described in the [integration guide](./INTEGRATION.md).

When removing validators, the validator can not be instantly removed from the set and it required a few intermediate blocks.

**Flow**


A validator is removed by an admin at height H; it increases the minimum self delegation to current+1. This puts the bonded validator into `Unbonding` status when checked by the x/staking module in it's BeginBlock (at H). The next iteration (H+1) the PoA module force updates the `Unbonding` validator to the status of `Unbonded`. The x/staking module then performs it's BeginBlock logic and sets it as `Unbonded`. The validator is now deleted from consensus at H+2 in the PoA BeginBlock.

## Client

| LCD / API | gRPC |
|------------------------------|-----------------------------------------------------|
| `/poa/v1/params`             |`strangelove_ventures.poa.v1.Query/Params`           |
| `/poa/v1/pending_validators` |`strangelove_ventures.poa.v1.Query/PendingValidators`|
| `/poa/v1/{consensus_power}`  |`strangelove_ventures.poa.v1.Query/ConsensusPower`   |


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

# (admin) Update the PoA module params
# - admins is a comma separated list of bech32 addresses
# - any admin can modify the list at any time
# - there must be at least one admin in the list at all times
# - allow_validator_self_exit is a bool to allow validators to force remove themselves from the set.
poad tx poa update-params [admin1,admin2,admin3,...] [allow_validator_self_exit]

# (admin) Update the staking module params
# - unbondingTime is the time that a validator must wait to unbond (ex: 336h)
# - maxVals is the maximum number of validators for the active set.
# - maxEntries is the maximum number of unbonding delegations per validator (does not apply for PoA)
# - historicalEntries is the maximum number of historical entries stored per validator (does not apply for PoA)
# - bondDenom is the denom of the bond token (e.g. uatom)
# - minCommissionRate is the minimum commission rate a validator can set
poad tx poa update-staking-params [unbondingTime] [maxVals] [maxEntries] [historicalEntries] [bondDenom] [minCommissionRate]
```

## Appendix

- [Cosmos-SDK Staking Module](https://github.com/cosmos/cosmos-sdk/tree/main/x/staking)