package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/strangelove-ventures/poa"
)

var _ poa.QueryServer = queryServer{}

// NewQueryServerImpl returns an implementation of the module QueryServer.
func NewQueryServerImpl(k Keeper) poa.QueryServer {
	return queryServer{k}
}

type queryServer struct {
	k Keeper
}

// ConsensusPower returns the consensus power of the given validator.
func (qs queryServer) ConsensusPower(ctx context.Context, msg *poa.QueryConsensusPowerRequest) (*poa.QueryConsensusPowerResponse, error) {
	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return nil, err
	}

	// verify that the validator exists
	if _, err := qs.k.stakingKeeper.GetValidator(ctx, valAddr); err != nil {
		return nil, err
	}

	lastPower, err := qs.k.stakingKeeper.GetLastValidatorPower(ctx, valAddr)
	if err != nil {
		return nil, err
	}

	return &poa.QueryConsensusPowerResponse{ConsensusPower: lastPower}, nil
}

// PendingValidators returns the pending validators.
func (qs queryServer) PendingValidators(ctx context.Context, _ *poa.QueryPendingValidatorsRequest) (*poa.PendingValidatorsResponse, error) {
	pending, err := qs.k.GetPendingValidators(ctx)
	if err != nil {
		return nil, err
	}

	return &poa.PendingValidatorsResponse{Pending: pending.Validators}, nil
}
