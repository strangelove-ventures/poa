package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
)

type BankKeeper interface {
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error
}

type SlashingKeeper interface {
	DeleteMissedBlockBitmap(ctx context.Context, addr sdk.ConsAddress) error
	SetValidatorSigningInfo(ctx context.Context, address sdk.ConsAddress, info slashingtypes.ValidatorSigningInfo) error

	GetValidatorSigningInfo(ctx context.Context, address sdk.ConsAddress) (slashingtypes.ValidatorSigningInfo, error)
}
