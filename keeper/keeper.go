package keeper

import (
	"context"
	"os"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"cosmossdk.io/collections"
	addresscodec "cosmossdk.io/core/address"
	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"

	"github.com/strangelove-ventures/poa"
)

type Keeper struct {
	cdc codec.BinaryCodec

	stakingKeeper StakingKeeper
	accountKeeper AccountKeeper // for testing
	slashKeeper   SlashingKeeper
	bankKeeper    BankKeeper

	logger log.Logger

	// state management
	Schema                 collections.Schema
	PendingValidators      collections.Item[poa.Validators]
	UpdatedValidatorsCache collections.KeySet[string]

	CachedBlockPower            collections.Item[poa.PowerCache]
	AbsoluteChangedInBlockPower collections.Item[poa.PowerCache]

	authority string
}

// NewKeeper creates a new poa Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeService storetypes.KVStoreService,
	sk StakingKeeper,
	slk SlashingKeeper,
	bk BankKeeper,
	logger log.Logger,
	adminAuthority string,
) Keeper {
	logger = logger.With(log.ModuleKey, "x/"+poa.ModuleName)

	sb := collections.NewSchemaBuilder(storeService)

	if address := os.Getenv("POA_ADMIN_ADDRESS"); address != "" {
		adminAuthority = address
		logger.Info("admin authority override from environment variable `POA_ADMIN_ADDRESS`", "address", adminAuthority)
	}

	k := Keeper{
		cdc:           cdc,
		stakingKeeper: sk,
		slashKeeper:   slk,
		bankKeeper:    bk,
		logger:        logger,

		// Stores
		PendingValidators:      collections.NewItem(sb, poa.PendingValidatorsKey, "pending", codec.CollValue[poa.Validators](cdc)),
		UpdatedValidatorsCache: collections.NewKeySet(sb, poa.UpdatedValidatorsCacheKey, "updated_validators", collections.StringKey),

		CachedBlockPower:            collections.NewItem(sb, poa.CachedPreviousBlockPowerKey, "cached_block", codec.CollValue[poa.PowerCache](cdc)),
		AbsoluteChangedInBlockPower: collections.NewItem(sb, poa.AbsoluteChangedInBlockPowerKey, "absolute_changed_power", codec.CollValue[poa.PowerCache](cdc)),

		authority: adminAuthority,
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}

	k.Schema = schema

	return k
}

// GetStakingKeeper returns the staking keeper.
func (k Keeper) GetStakingKeeper() StakingKeeper {
	return k.stakingKeeper
}

func (k Keeper) GetValidatorAddressCodec() addresscodec.Codec {
	return k.stakingKeeper.ValidatorAddressCodec()
}

// GetSlashingKeeper returns the slashing keeper.
func (k Keeper) GetSlashingKeeper() SlashingKeeper {
	return k.slashKeeper
}

func (k Keeper) GetBankKeeper() BankKeeper {
	return k.bankKeeper
}

func (k *Keeper) GetTestAccountKeeper() AccountKeeper {
	return k.accountKeeper
}

func (k *Keeper) SetTestAccountKeeper(ak AccountKeeper) {
	k.accountKeeper = ak
}

func (k *Keeper) SetTestAuthority(addr string) {
	k.authority = addr
}

func (k Keeper) GetAdmin(ctx context.Context) string {
	return k.authority
}

// IsAdmin checks if the given address is an admin.
func (k Keeper) IsAdmin(ctx context.Context, fromAddr string) bool {
	if os.Getenv("POA_BYPASS_ADMIN_CHECK_FOR_SIMULATION_TESTING_ONLY") == "not_for-production" {
		return true
	}

	return k.authority == fromAddr
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

// ResetCachedTotalPower resets the block power index to the current total power.
func (k Keeper) ResetCachedTotalPower(ctx context.Context) error {
	currValPower, err := k.GetStakingKeeper().GetLastTotalPower(ctx)
	if err != nil {
		return err
	}

	prev, err := k.GetCachedBlockPower(ctx)
	if err != nil {
		return err
	}

	if currValPower.Uint64() != prev {
		return k.SetCachedBlockPower(ctx, currValPower.Uint64())
	}

	return nil
}

// resetAbsoluteBlockPower resets the absolute block power to 0 since updates per block have been executed upon.
func (k Keeper) ResetAbsoluteBlockPower(ctx context.Context) error {
	var err error

	val, err := k.GetAbsoluteChangedInBlockPower(ctx)
	if err != nil {
		return err
	} else if val != 0 {
		return k.SetAbsoluteChangedInBlockPower(ctx, 0)
	}

	return err
}
