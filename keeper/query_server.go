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

func (qs queryServer) QueryValidator(context.Context, *poa.QueryValidatorRequest) (*poa.QueryValidatorResponse, error) {
	return nil, nil
}

func (qs queryServer) QueryValidators(context.Context, *poa.QueryValidatorsRequest) (*poa.QueryValidatorsResponse, error) {
	return nil, nil
}

func (qs queryServer) QueryVouch(context.Context, *poa.QueryVouchRequest) (*poa.QueryVouchResponse, error) {
	return nil, nil
}

func (qs queryServer) QueryVouches(context.Context, *poa.QueryVouchesRequest) (*poa.QueryVouchesResponse, error) {
	return nil, nil
}
