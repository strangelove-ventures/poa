package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"cosmossdk.io/collections"
	addresscodec "cosmossdk.io/core/address"
	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"

	"github.com/strangelove-ventures/poa"
)

type Keeper struct {
	cdc                   codec.BinaryCodec
	validatorAddressCodec addresscodec.Codec

	stakingKeeper *stakingkeeper.Keeper
	slashKeeper   SlashingKeeper
	bankKeeper    BankKeeper

	logger log.Logger

	// state management
	Schema            collections.Schema
	Params            collections.Item[poa.Params]
	PendingValidators collections.Item[poa.Validators]

	CachedBlockPower            collections.Item[poa.PowerCache]
	AbsoluteChangedInBlockPower collections.Item[poa.PowerCache]

	// if set, anyone with this token is a PoA admin
	AdminTokenDenom string
}

// NewKeeper creates a new poa Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeService storetypes.KVStoreService,
	sk *stakingkeeper.Keeper,
	slk SlashingKeeper,
	bk BankKeeper,
	validatorAddressCodec addresscodec.Codec,
	logger log.Logger,
) Keeper {
	logger = logger.With(log.ModuleKey, "x/"+poa.ModuleName)

	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		cdc:                   cdc,
		stakingKeeper:         sk,
		validatorAddressCodec: validatorAddressCodec,
		slashKeeper:           slk,
		bankKeeper:            bk,
		logger:                logger,
		AdminTokenDenom:       "", // disabled by default

		// Stores
		Params:            collections.NewItem(sb, poa.ParamsKey, "params", codec.CollValue[poa.Params](cdc)),
		PendingValidators: collections.NewItem(sb, poa.PendingValidatorsKey, "pending", codec.CollValue[poa.Validators](cdc)),

		CachedBlockPower:            collections.NewItem(sb, poa.CachedPreviousBlockPowerKey, "cached_block", codec.CollValue[poa.PowerCache](cdc)),
		AbsoluteChangedInBlockPower: collections.NewItem(sb, poa.AbsoluteChangedInBlockPowerKey, "absolute_changed_power", codec.CollValue[poa.PowerCache](cdc)),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}

	k.Schema = schema

	return k
}

// GetStakingKeeper returns the staking keeper.
func (k Keeper) GetStakingKeeper() *stakingkeeper.Keeper {
	return k.stakingKeeper
}

// GetSlashingKeeper returns the slashing keeper.
func (k Keeper) GetSlashingKeeper() SlashingKeeper {
	return k.slashKeeper
}

func (k Keeper) GetBankKeeper() BankKeeper {
	return k.bankKeeper
}

// SetAdminTokenDenom sets a token denomination that when held by an account, makes them a PoA admin.
// This allows for key rotation like behaviors using standard bank denoms for PoA.
// Pairs nicely with tokenfactory.
func (k *Keeper) SetAdminTokenDenom(denom string) {
	k.AdminTokenDenom = denom
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
	if k.AdminTokenDenom != "" {
		accAddr, err := sdk.AccAddressFromBech32(fromAddr)
		if err != nil {
			return false
		}

		bal := k.bankKeeper.GetBalance(ctx, accAddr, k.AdminTokenDenom)
		return !bal.IsZero()
	}

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

// updateBondedPoolPower updates the bonded pool to the correct power for the network.
func (k Keeper) UpdateBondedPoolPower(ctx context.Context) error {
	newTotal := sdkmath.ZeroInt()

	del, err := k.stakingKeeper.GetAllDelegations(ctx)
	if err != nil {
		return err
	}

	for _, d := range del {
		newTotal = newTotal.Add(d.Shares.RoundInt())
	}

	bondDenom, err := k.stakingKeeper.BondDenom(ctx)
	if err != nil {
		return err
	}

	prevBal := k.bankKeeper.GetBalance(ctx, authtypes.NewModuleAddress(stakingtypes.BondedPoolName), bondDenom).Amount

	if newTotal.Equal(prevBal) {
		return nil
	}

	if newTotal.GT(prevBal) {
		diff := newTotal.Sub(prevBal)
		coins := sdk.NewCoins(sdk.NewCoin(bondDenom, diff))

		if err := k.bankKeeper.MintCoins(ctx, minttypes.ModuleName, coins); err != nil {
			return err
		}

		if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, stakingtypes.BondedPoolName, coins); err != nil {
			return err
		}
	}

	// no need to check if it goes down. When it does, it's automatic from the staking module as tokens are moved from
	// bonded -> ToNotBonded pool. As PoA, we do not want any tokens in the ToNotBonded pool, so when a validator is removed
	// they are slashed 100% (since it is PoA this is fine) which decreases the BondedPool balance, and leave NotBonded at 0.

	return nil
}
