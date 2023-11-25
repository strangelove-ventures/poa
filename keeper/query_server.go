package keeper

import (
	"context"

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

// PendingValidators returns the pending validators.
func (qs queryServer) PendingValidators(ctx context.Context, _ *poa.QueryPendingValidatorsRequest) (*poa.PendingValidatorsResponse, error) {
	pending, err := qs.k.GetPendingValidators(ctx)
	if err != nil {
		return nil, err
	}

	return &poa.PendingValidatorsResponse{Pending: pending.Validators}, nil
}

// Params returns the current module params.
func (qs queryServer) Params(ctx context.Context, msg *poa.QueryParamsRequest) (*poa.ParamsResponse, error) {
	params, err := qs.k.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	return &poa.ParamsResponse{Params: params}, nil
}
