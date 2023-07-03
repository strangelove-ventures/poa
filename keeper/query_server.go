package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/strangelove-ventures/poa"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ poa.QueryServer = queryServer{}

// NewQueryServerImpl returns an implementation of the module QueryServer.
func NewQueryServerImpl(k Keeper) poa.QueryServer {
	return queryServer{k}
}

type queryServer struct {
	k Keeper
}

func (qs queryServer) QueryValidator(goCtx context.Context, req *poa.QueryValidatorRequest) (*poa.QueryValidatorResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	addr, err := sdk.AccAddressFromBech32(req.ValidatorAddress)
	if err != nil {
		return nil, err
	}

	val, found := qs.k.GetValidator(ctx, addr)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &poa.QueryValidatorResponse{
		ValidatorAddress: req.ValidatorAddress,
		IsAccepted:       val.IsAccepted,
	}, nil
}

// TODO: Fix pagination
func (qs queryServer) QueryValidators(goCtx context.Context, req *poa.QueryValidatorsRequest) (*poa.QueryValidatorsResponse, error) {
	// if req == nil {
	// 	return nil, status.Error(codes.InvalidArgument, "invalid request")
	// }

	// var validators []*types.QueryValidatorResponse
	// ctx := sdk.UnwrapSDKContext(goCtx)

	// store := ctx.KVStore(qs.k.storeKey)
	// validatorStore := prefix.NewStore(store, types.ValidatorsKey)

	// pageRes, err := query.Paginate(validatorStore, req.Pagination, func(key []byte, value []byte) error {
	// 	var validator types.QueryValidatorResponse
	// 	if err := k.cdc.Unmarshal(value, &validator); err != nil {
	// 		return err
	// 	}

	// 	validators = append(validators, &validator)
	// 	return nil
	// })

	// if err != nil {
	// 	return nil, status.Error(codes.Internal, err.Error())
	// }

	// return &poa.QueryValidatorsResponse{Validators: validators, Pagination: pageRes}, nil
	return nil, nil
}

func (qs queryServer) QueryVouch(goCtx context.Context, req *poa.QueryVouchRequest) (*poa.QueryVouchResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	candidateAddr, err := sdk.AccAddressFromBech32(req.CandidateAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to decode candidate address as bech32: %w", err)
	}

	vouchrAddr, err := sdk.AccAddressFromBech32(req.CandidateAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to decode vouchr address as bech32: %w", err)
	}

	val, found := qs.k.GetVouch(ctx, append(candidateAddr, vouchrAddr...))
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &poa.QueryVouchResponse{
		CandidateAddress: req.CandidateAddress,
		VoucherAddress:   req.VoucherAddress,
		InFavor:          val.InFavor,
	}, nil
}

// TODO: fix pagination
func (qs queryServer) QueryVouches(goCtx context.Context, req *poa.QueryVouchesRequest) (*poa.QueryVouchesResponse, error) {
	// if req == nil {
	// 	return nil, status.Error(codes.InvalidArgument, "invalid request")
	// }

	// var vouches []*types.QueryVouchResponse
	// ctx := sdk.UnwrapSDKContext(c)

	// var candidateBz []byte
	// if req.CandidateAddress != "" {
	// 	candidateAddr, err := sdk.AccAddressFromBech32(req.CandidateAddress)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("failed to decode candidate address as bech32: %w", err)
	// 	}
	// 	candidateBz = candidateAddr
	// }

	// store := ctx.KVStore(k.storeKey)
	// vouchStore := prefix.NewStore(store, append(types.VouchesKey, candidateBz...))

	// pageRes, err := query.Paginate(vouchStore, req.Pagination, func(key []byte, value []byte) error {
	// 	var vouch types.QueryVouchResponse
	// 	if err := k.cdc.Unmarshal(value, &vouch); err != nil {
	// 		return err
	// 	}

	// 	vouches = append(vouches, &vouch)
	// 	return nil
	// })

	// if err != nil {
	// 	return nil, status.Error(codes.Internal, err.Error())
	// }

	// return &types.QueryVouchesResponse{Vouches: vouches, Pagination: pageRes}, nil
	return nil, nil
}
