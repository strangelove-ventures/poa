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
}

// We use almost all the methods from StakingKeeper. Just using it directly
/*
type StakingKeeper interface {
	Hooks() stakingtypes.StakingHooks

	SetParams(ctx context.Context, params stakingtypes.Params) error
	SetLastValidatorPower(ctx context.Context, operator sdk.ValAddress, power int64) error
	SetDelegation(ctx context.Context, delegation stakingtypes.Delegation) error
	SetValidator(ctx context.Context, validator stakingtypes.Validator) error
	SetValidatorByConsAddr(ctx context.Context, validator stakingtypes.Validator) error
	SetNewValidatorByPowerIndex(ctx context.Context, validator stakingtypes.Validator) error
	SetLastTotalPower(ctx context.Context, power math.Int) error
	DeleteLastValidatorPower(ctx context.Context, operator sdk.ValAddress) error
	GetLastTotalPower(ctx context.Context) (math.Int, error)

	Slash(ctx context.Context, consAddr sdk.ConsAddress, infractionHeight, power int64, slashFactor math.LegacyDec) (math.Int, error)

	MinCommissionRate(ctx context.Context) (math.LegacyDec, error)
	BondDenom(ctx context.Context) (string, error)

	GetValidator(ctx context.Context, addr sdk.ValAddress) (validator stakingtypes.Validator, err error)
	GetAllValidators(ctx context.Context) (validators []stakingtypes.Validator, err error)
	GetValidatorByConsAddr(ctx context.Context, consAddr sdk.ConsAddress) (validator stakingtypes.Validator, err error)
	GetLastValidatorPower(ctx context.Context, operator sdk.ValAddress) (power int64, err error)

	GetAllDelegations(ctx context.Context) (delegations []stakingtypes.Delegation, err error)

	TokensFromConsensusPower(ctx context.Context, power int64) math.Int
	TokensToConsensusPower(ctx context.Context, tokens math.Int) int64

	PowerReduction(ctx context.Context) math.Int
}
*/
