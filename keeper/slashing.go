package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// setSlashingInfo sets validator slashing defaults (useful for downtime jailing)
func (k Keeper) setSlashingInfo(sdkCtx sdk.Context, val stakingtypes.Validator) error {
	cons, err := val.GetConsAddr()
	if err != nil {
		return err
	}

	ctx := sdk.UnwrapSDKContext(sdkCtx)
	return k.slashKeeper.SetValidatorSigningInfo(ctx, sdk.ConsAddress(cons), slashingtypes.ValidatorSigningInfo{
		Address:             sdk.ConsAddress(cons).String(),
		StartHeight:         sdkCtx.BlockHeight(),
		IndexOffset:         0,
		JailedUntil:         sdkCtx.BlockHeader().Time,
		Tombstoned:          false,
		MissedBlocksCounter: 0,
	})
}

func (k Keeper) clearSlashingInfo(ctx context.Context, val stakingtypes.Validator) error {
	cons, err := val.GetConsAddr()
	if err != nil {
		return err
	}

	if err := k.slashKeeper.DeleteMissedBlockBitmap(ctx, sdk.ConsAddress(cons)); err != nil {
		return err
	}

	return k.slashKeeper.SetValidatorSigningInfo(ctx, sdk.ConsAddress(cons), slashingtypes.ValidatorSigningInfo{})
}
