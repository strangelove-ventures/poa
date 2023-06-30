# `attestation`

Attestation module processes attestations validators make about the state of rollups. It is responsible for storing attestations about rollup hashes to state and finalizing  them once 2/3+ of validators have seen the same hash. Old attestations, heights after a certain time period, and attestations after a certain time period are pruned.

is the logic for initializing coverage of the rollups another module? I think yes. attestation should just be for storing and finalizing attestations.


## Concepts

Describe specialized concepts and definitions used throughout the spec.

## State

Specify and describe structures expected to marshalled into the store, and their keys

Each block validators submit `[]HeaderAttestation` via vote extensions. When these are processed they are added to the store as follows:

```go
type HeaderAttestation struct {
    Validator sdk.ValAddress
    ChainId   string
    Height    uint64
    Hash      []byte
}
// Store path /{module_name}/unconfirmed/{chain_id}/{height}/{hash}/{validator_addr} -> bool
// Or as a go map:
// map[fmt.Sprintf("%s/%d/%x",chain_id,height,hash)][]sdk.ValAddress
// map[string][]sdk.ValAddress
```

The `attestation` module iterates over each chain's unconfirmed heights and hashes and checks if 2/3+ of validators have seen the same hash. If so, the attestation is finalized and added to the store as follows:

```go
type CoveredHeader struct {
    ChainId string
    Height  uint64
    Hash    []byte
}
// Store path /{module_name}/finalized/{chain_id}/{height} -> {hash}
```

## State Transitions

Each Begin Block we iterate over the unconfirmed store and check if 2/3+ of validators have seen the same hash. If so, we finalize the attestation and add it to the finalized store.

```go
func (ave AttestationVoteExtension) ExtendVote(ctx sdk.Context, req *abci.RequestExtendVote) (*abci.ResponseExtendVote, error) {
    // TODO: do we limit size of votes each block?
    messagemax := ave.MaxVoteSize(ctx)
    // []*CoveredHeader
    observations := ave.Cache.CoveredHeaderObservations(ctx)
    marshalledObservations := [][]byte{}
    for _, observation := range observations {
        bz, err := json.Marshal(observation)
        if err != nil {
            return nil, fmt.Errorf("failed to encode vote extension: %w", err)
        }
        if messagemax - len(bz) > 0 {
            marshalledObservations = append(marshalledObservations, bz)
            messagemax -= len(bz)
        } else {
            break
        }
    }
    // turn the [][]byte into single []byte parsable json (i.e. add commas and [])
    return &abci.ResponseExtendVote{VoteExtension: bz}, nil
 }

// TODO: can we track covered chains and latest heights on the ve struct?
func (ave AttestationVoteExtension) VerifyVoteExtension(ctx sdk.Context, req *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error) {
    var vote CoveredHeader
    if err := json.Unmarshal(req.VoteExtension, &vote); err != nil {
        return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}, nil
    }
    // map[chain_id]height
    chains := ave.CoveredChains(ctx)
    switch {
    case chains[req.ChainId] == nil:
        return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}, nil
    case chains[req.ChainId] >= req.Height:
        return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}, nil
    case len(ve.Data) != 1024:
        return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}, nil
    }
    if err := ave.Keeper.WriteHeaderAttestation(ctx, vote.HeaderAttestation(req.Validator)); err != nil {
        return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}, nil
    }
    return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_ACCEPT}, nil
}
```

## Messages

Specify message structure(s) and expected state machine behaviour(s).

## Begin Block

Specify any begin-block operations.

## End Block

Specify any end-block operations.

## Hooks

Describe available hooks to be called by/from this module.

## Events

List and describe event tags used.

## Client

List and describe CLI commands and gRPC and REST endpoints.

## Params

List all module parameters, their types (in JSON) and examples.

## Future Improvements

Describe future improvements of this module.

## Tests

Acceptance tests.

## Appendix

Supplementary details referenced elsewhere within the spec.