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

// QueryRemoveMe implements poa.QueryServer.
func (queryServer) QueryRemoveMe(context.Context, *poa.QueryRemoveMeRequest) (*poa.QueryRemoveMeResponse, error) {
	panic("unimplemented")
}
