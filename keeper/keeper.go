package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	addresscodec "cosmossdk.io/core/address"
	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/strangelove-ventures/poa"
)

type Keeper struct {
	cdc                   codec.BinaryCodec
	storeService          storetypes.KVStoreService
	validatorAddressCodec addresscodec.Codec

	stakingKeeper *stakingkeeper.Keeper
	slashKeeper   slashingkeeper.Keeper

	logger log.Logger
}

// NewKeeper creates a new poa Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeService storetypes.KVStoreService,
	sk *stakingkeeper.Keeper,
	slk slashingkeeper.Keeper,
	validatorAddressCodec addresscodec.Codec,
	logger log.Logger,
) Keeper {
	logger = logger.With(log.ModuleKey, "x/"+poa.ModuleName)

	k := Keeper{
		cdc:                   cdc,
		storeService:          storeService,
		stakingKeeper:         sk,
		validatorAddressCodec: validatorAddressCodec,
		slashKeeper:           slk,
		logger:                logger,
	}

	return k
}

// GetStakingKeeper returns the staking keeper.
func (k Keeper) GetStakingKeeper() *stakingkeeper.Keeper {
	return k.stakingKeeper
}

// GetSlashingKeeper returns the slashing keeper.
func (k Keeper) GetSlashingKeeper() slashingkeeper.Keeper {
	return k.slashKeeper
}

// GetAdmins returns the module's administrators with delegation of power control.
func (k Keeper) GetAdmins(ctx context.Context) []string {
	p, err := k.GetParams(ctx)
	if err != nil {
		return []string{}
	}

	return p.Admins
}

// IsAdmin checks if the given address is an admin.
func (k Keeper) IsAdmin(ctx context.Context, fromAddr string) bool {
	for _, auth := range k.GetAdmins(ctx) {
		if auth == fromAddr {
			return true
		}
	}

	return false
}

// IsSenderValidator checks if the given sender address is the same address as the validator by bytes.
func (k Keeper) IsSenderValidator(ctx context.Context, sender string, expectedValidator string) (bool, error) {
	from, err := sdk.AccAddressFromBech32(sender)
	if err != nil {
		return false, sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}

	expectedVal, err := sdk.ValAddressFromBech32(expectedValidator)
	if err != nil {
		return false, sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}

	return from.Equals(expectedVal), nil
}

func (k Keeper) Logger() log.Logger {
	return k.logger
}
