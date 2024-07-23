package keeper

import (
	"context"

	abci "github.com/cometbft/cometbft/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	addresscodec "cosmossdk.io/core/address"
	"cosmossdk.io/math"
)

type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
}

type BankKeeper interface {
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
}

type SlashingKeeper interface {
	DeleteMissedBlockBitmap(ctx context.Context, addr sdk.ConsAddress) error
	SetValidatorSigningInfo(ctx context.Context, address sdk.ConsAddress, info slashingtypes.ValidatorSigningInfo) error

	GetValidatorSigningInfo(ctx context.Context, address sdk.ConsAddress) (slashingtypes.ValidatorSigningInfo, error)
}

type StakingKeeper interface {
	GetValidator(ctx context.Context, addr sdk.ValAddress) (validator stakingtypes.Validator, err error)
	GetLastValidatorPower(ctx context.Context, operator sdk.ValAddress) (power int64, err error)
	SetDelegation(ctx context.Context, delegation stakingtypes.Delegation) error
	SetValidator(ctx context.Context, validator stakingtypes.Validator) error
	SetLastValidatorPower(ctx context.Context, operator sdk.ValAddress, power int64) error
	TokensToConsensusPower(ctx context.Context, tokens math.Int) int64
	TokensFromConsensusPower(ctx context.Context, power int64) math.Int
	Slash(ctx context.Context, consAddr sdk.ConsAddress, infractionHeight, power int64, slashFactor math.LegacyDec) (math.Int, error)
	DeleteLastValidatorPower(ctx context.Context, operator sdk.ValAddress) error
	DeleteValidatorByPowerIndex(ctx context.Context, validator stakingtypes.Validator) error
	SetNewValidatorByPowerIndex(ctx context.Context, validator stakingtypes.Validator) error
	SetValidatorByPowerIndex(ctx context.Context, validator stakingtypes.Validator) error
	SetValidatorByConsAddr(ctx context.Context, validator stakingtypes.Validator) error
	GetAllValidators(ctx context.Context) (validators []stakingtypes.Validator, err error)
	PowerReduction(ctx context.Context) math.Int
	SetLastTotalPower(ctx context.Context, power math.Int) error
	MinCommissionRate(ctx context.Context) (math.LegacyDec, error)
	GetValidatorByConsAddr(ctx context.Context, consAddr sdk.ConsAddress) (validator stakingtypes.Validator, err error)
	SetParams(ctx context.Context, params stakingtypes.Params) error
	GetAllDelegations(ctx context.Context) (delegations []stakingtypes.Delegation, err error)
	BondDenom(ctx context.Context) (string, error)
	ValidatorAddressCodec() addresscodec.Codec
	GetLastTotalPower(ctx context.Context) (math.Int, error)
	GetBondedValidatorsByPower(ctx context.Context) ([]stakingtypes.Validator, error)
	GetValidatorUpdates(ctx context.Context) ([]abci.ValidatorUpdate, error)

	BeginBlocker(ctx context.Context) error
	EndBlocker(ctx context.Context) ([]abci.ValidatorUpdate, error)

	Hooks() stakingtypes.StakingHooks
}
