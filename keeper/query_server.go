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

func (qs queryServer) Params(ctx context.Context, msg *poa.QueryParamsRequest) (*poa.QueryParamsResponse, error) {
	params, err := qs.k.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	return &poa.QueryParamsResponse{Params: params}, nil
}
