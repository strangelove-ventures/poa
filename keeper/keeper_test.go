package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	abci "github.com/cometbft/cometbft/abci/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/strangelove-ventures/poa"
	"github.com/strangelove-ventures/poa/keeper"
	poamodule "github.com/strangelove-ventures/poa/module"
)

var maccPerms = map[string][]string{
	authtypes.FeeCollectorName:     nil,
	stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
	stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
	minttypes.ModuleName:           {authtypes.Minter},
	govtypes.ModuleName:            {authtypes.Burner},
}

type testFixture struct {
	suite.Suite

	ctx         sdk.Context
	k           keeper.Keeper
	msgServer   poa.MsgServer
	queryServer poa.QueryServer
	appModule   poamodule.AppModule

	accountkeeper  authkeeper.AccountKeeper
	stakingKeeper  *stakingkeeper.Keeper
	slashingKeeper slashingkeeper.Keeper
	bankkeeper     bankkeeper.BaseKeeper
	mintkeeper     mintkeeper.Keeper

	addrs      []sdk.AccAddress
	govModAddr string
}

func SetupTest(t *testing.T, baseValShares int64) *testFixture {
	t.Helper()
	f := new(testFixture)
	require := require.New(t)

	// Base setup
	logger := log.NewTestLogger(t)
	encCfg := moduletestutil.MakeTestEncodingConfig()

	f.govModAddr = authtypes.NewModuleAddress(govtypes.ModuleName).String()
	f.addrs = simtestutil.CreateIncrementalAccounts(3)

	key := storetypes.NewKVStoreKey(poa.ModuleName)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))

	f.ctx = testCtx.Ctx

	// Register SDK modules.
	registerBaseSDKModules(f, encCfg, storeService, logger, require)

	// Setup POA Keeper.
	f.k = keeper.NewKeeper(encCfg.Codec, storeService, f.stakingKeeper, f.slashingKeeper, f.bankkeeper, f.accountkeeper, logger)
	f.msgServer = keeper.NewMsgServerImpl(f.k)
	f.queryServer = keeper.NewQueryServerImpl(f.k)
	f.appModule = poamodule.NewAppModule(encCfg.Codec, f.k)

	// register interfaces
	registerModuleInterfaces(encCfg)

	// Setup initial keeper states
	require.NoError(f.accountkeeper.AccountNumber.Set(f.ctx, 1))
	f.accountkeeper.SetModuleAccount(f.ctx, f.stakingKeeper.GetNotBondedPool(f.ctx))
	f.accountkeeper.SetModuleAccount(f.ctx, f.stakingKeeper.GetBondedPool(f.ctx))
	f.accountkeeper.SetModuleAccount(f.ctx, f.accountkeeper.GetModuleAccount(f.ctx, minttypes.ModuleName))
	f.mintkeeper.InitGenesis(f.ctx, f.accountkeeper, minttypes.DefaultGenesisState())

	// Set initial PoA state
	f.InitPoAGenesis(t)

	f.createBaseStakingValidators(t, baseValShares)

	return f
}

func (f *testFixture) InitPoAGenesis(t *testing.T) {
	t.Helper()

	genState := poa.NewGenesisState()
	genState.Params.Admins = []string{f.addrs[0].String(), f.govModAddr}
	require.NoError(t, f.k.InitGenesis(f.ctx, genState))
}

func registerBaseSDKModules(
	f *testFixture,
	encCfg moduletestutil.TestEncodingConfig,
	storeService store.KVStoreService,
	logger log.Logger,
	require *require.Assertions,
) {
	// Auth Keeper.
	f.accountkeeper = authkeeper.NewAccountKeeper(
		encCfg.Codec, storeService,
		authtypes.ProtoBaseAccount,
		maccPerms,
		authcodec.NewBech32Codec(sdk.Bech32MainPrefix), sdk.Bech32MainPrefix,
		f.govModAddr,
	)

	// Bank Keeper.
	f.bankkeeper = bankkeeper.NewBaseKeeper(
		encCfg.Codec, storeService,
		f.accountkeeper,
		nil,
		f.govModAddr, logger,
	)

	// Staking Keeper.
	f.stakingKeeper = stakingkeeper.NewKeeper(
		encCfg.Codec, storeService,
		f.accountkeeper, f.bankkeeper, f.govModAddr,
		authcodec.NewBech32Codec(sdk.Bech32PrefixValAddr),
		authcodec.NewBech32Codec(sdk.Bech32PrefixConsAddr),
	)
	err := f.stakingKeeper.SetParams(f.ctx, stakingtypes.DefaultParams())
	require.NoError(err)

	// Slashing Keeper.
	f.slashingKeeper = slashingkeeper.NewKeeper(
		encCfg.Codec, encCfg.Amino, storeService,
		f.stakingKeeper,
		f.govModAddr,
	)
	err = f.slashingKeeper.SetParams(f.ctx, slashingtypes.DefaultParams())
	require.NoError(err)

	// Mint Keeper.
	// This is required for `MintTokensToBondedPool` to remain happy.
	f.mintkeeper = mintkeeper.NewKeeper(
		encCfg.Codec, storeService,
		f.stakingKeeper, f.accountkeeper, f.bankkeeper,
		authtypes.FeeCollectorName, f.govModAddr,
	)
}

func registerModuleInterfaces(encCfg moduletestutil.TestEncodingConfig) {
	authtypes.RegisterInterfaces(encCfg.InterfaceRegistry)
	stakingtypes.RegisterInterfaces(encCfg.InterfaceRegistry)
	poa.RegisterInterfaces(encCfg.InterfaceRegistry)
}

type valSetup struct {
	priv   *secp256k1.PrivKey
	addr   sdk.AccAddress
	valKey *ed25519.PrivKey
}

func GenAcc() valSetup {
	priv := secp256k1.GenPrivKey()

	return valSetup{
		priv:   priv,
		addr:   sdk.AccAddress(priv.PubKey().Address()),
		valKey: ed25519.GenPrivKey(),
	}
}

func (f *testFixture) createBaseStakingValidators(t *testing.T, baseValShares int64) {
	t.Helper()
	require := require.New(t)
	bondCoin := sdk.NewCoin("stake", sdkmath.NewInt(baseValShares))

	vals := []valSetup{
		GenAcc(),
		GenAcc(),
		GenAcc(),
	}

	for idx, val := range vals {
		valAddr := sdk.ValAddress(val.addr).String()

		pubKey := val.valKey.PubKey()

		val := poa.ConvertPOAToStaking(CreateNewValidator(
			fmt.Sprintf("val-%d", idx),
			valAddr,
			pubKey,
			bondCoin.Amount.Int64(),
		))

		if err := f.k.AddPendingValidator(f.ctx, val, pubKey); err != nil {
			panic(err)
		}

		_, err := f.msgServer.SetPower(f.ctx, &poa.MsgSetPower{
			Sender:           f.addrs[0].String(),
			ValidatorAddress: valAddr,
			Power:            bondCoin.Amount.Uint64(),
			Unsafe:           true,
		})
		require.NoError(err)

		// increase the block so the new validator is in the validator set
		_, err = f.IncreaseBlock(1)
		require.NoError(err)

		valAddrBz, err := sdk.ValAddressFromBech32(val.GetOperator())
		require.NoError(err)

		validator, err := f.stakingKeeper.GetValidator(f.ctx, valAddrBz)
		require.NoError(err)

		validator.Status = stakingtypes.Bonded
		if err := f.stakingKeeper.SetValidator(f.ctx, validator); err != nil {
			panic(err)
		}

		if _, err := f.IncreaseBlock(1, true); err != nil {
			panic(err)
		}
	}

	totalBondToken := bondCoin.Amount.MulRaw(int64(len(vals)))
	total := sdkmath.NewInt(f.stakingKeeper.TokensToConsensusPower(f.ctx, totalBondToken))
	if err := f.stakingKeeper.SetLastTotalPower(f.ctx, total); err != nil {
		panic(err)
	}

	if err := f.k.InitCacheStores(f.ctx); err != nil {
		panic(err)
	}

	// override proper consensus power for testing (from InitCacheStores)
	if err := f.k.SetCachedBlockPower(f.ctx, total.Uint64()); err != nil {
		panic(err)
	}

	// inc block
	if _, err := f.IncreaseBlock(1, true); err != nil {
		panic(err)
	}
}

func CreateNewValidator(moniker string, opAddr string, pubKey cryptotypes.PubKey, amt int64) poa.Validator {
	var pkAny *codectypes.Any
	if pubKey != nil {
		var err error
		if pkAny, err = codectypes.NewAnyWithValue(pubKey); err != nil {
			panic(err)
		}
	}

	return poa.Validator{
		OperatorAddress: opAddr,
		ConsensusPubkey: pkAny,
		Jailed:          false,
		Status:          poa.Bonded,
		Tokens:          sdkmath.NewInt(amt),
		DelegatorShares: sdkmath.LegacyNewDecFromInt(sdkmath.NewInt(amt)),
		Description:     poa.NewDescription(moniker, "", "", "", ""),
		UnbondingHeight: 0,
		UnbondingTime:   time.Time{},
		Commission: poa.Commission{
			CommissionRates: poa.NewCommissionRates(sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec()),
		},
		MinSelfDelegation:       sdkmath.OneInt(),
		UnbondingOnHoldRefCount: 0,
		UnbondingIds:            nil,
	}
}

func (f *testFixture) CreatePendingValidator(name string, power uint64) sdk.ValAddress {
	val := GenAcc()
	valAddr := sdk.ValAddress(val.addr)

	v := poa.ConvertPOAToStaking(CreateNewValidator(
		name,
		valAddr.String(),
		val.valKey.PubKey(),
		int64(power),
	))

	if err := f.k.AddPendingValidator(f.ctx, v, val.valKey.PubKey()); err != nil {
		panic(err)
	}

	if _, err := f.IncreaseBlock(1); err != nil {
		panic(err)
	}

	return valAddr
}

func (f *testFixture) IncreaseBlock(amt int64, debug ...bool) ([]abci.ValidatorUpdate, error) {
	f.ctx = f.ctx.WithBlockHeight(f.ctx.BlockHeight() + amt)

	allUpdates := make([]abci.ValidatorUpdate, 0)
	for i := int64(0); i < amt; i++ {
		if err := f.k.GetStakingKeeper().BeginBlocker(f.ctx); err != nil {
			return nil, err
		}

		updates, err := f.k.GetStakingKeeper().EndBlocker(f.ctx)
		if err != nil {
			return nil, err
		}

		allUpdates = append(allUpdates, updates...)
		if len(debug) > 0 && debug[0] && len(updates) > 0 {
			f.k.Logger().Debug("IncreaseBlock(...) updates", "updates", updates)
		}

		if err := f.appModule.BeginBlock(f.ctx); err != nil {
			return nil, err
		}
	}

	return allUpdates, nil
}
