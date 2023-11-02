package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/abci/types"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/strangelove-ventures/poa"
	"github.com/strangelove-ventures/poa/keeper"

	"cosmossdk.io/log"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

var (
	maccPerms = map[string][]string{
		authtypes.FeeCollectorName:     nil,
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:            {authtypes.Burner},
		poa.ModuleName:                 {authtypes.Minter},
	}
)

type testFixture struct {
	suite.Suite

	ctx         sdk.Context
	k           keeper.Keeper
	msgServer   poa.MsgServer
	queryServer poa.QueryServer

	accountkeeper  authkeeper.AccountKeeper
	stakingKeeper  *stakingkeeper.Keeper
	slashingKeeper slashingkeeper.Keeper
	bankkeeper     bankkeeper.BaseKeeper

	addrs      []sdk.AccAddress
	govModAddr string
}

func SetupTest(t *testing.T) *testFixture {
	s := new(testFixture)

	logger := log.NewTestLogger(t)
	s.govModAddr = authtypes.NewModuleAddress(govtypes.ModuleName).String()

	encCfg := moduletestutil.MakeTestEncodingConfig()
	key := storetypes.NewKVStoreKey(poa.ModuleName)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	s.ctx = testCtx.Ctx

	storeService := runtime.NewKVStoreService(key)
	s.addrs = simtestutil.CreateIncrementalAccounts(3)

	s.accountkeeper = authkeeper.NewAccountKeeper(encCfg.Codec, storeService, authtypes.ProtoBaseAccount, maccPerms, authcodec.NewBech32Codec(sdk.Bech32MainPrefix), sdk.Bech32MainPrefix, s.govModAddr)
	s.bankkeeper = bankkeeper.NewBaseKeeper(encCfg.Codec, storeService, s.accountkeeper, nil, s.govModAddr, logger)

	bankkeeper.NewMsgServerImpl(s.bankkeeper)

	s.stakingKeeper = stakingkeeper.NewKeeper(encCfg.Codec, storeService, s.accountkeeper, s.bankkeeper, s.govModAddr, authcodec.NewBech32Codec(sdk.Bech32PrefixValAddr), authcodec.NewBech32Codec(sdk.Bech32PrefixConsAddr))
	err := s.stakingKeeper.SetParams(s.ctx, stakingtypes.DefaultParams())
	require.NoError(t, err)

	s.slashingKeeper = slashingkeeper.NewKeeper(encCfg.Codec, encCfg.Amino, storeService, s.stakingKeeper, authtypes.NewModuleAddress(govtypes.ModuleName).String())
	err = s.slashingKeeper.SetParams(s.ctx, slashingtypes.DefaultParams())
	require.NoError(t, err)

	s.k = keeper.NewKeeper(encCfg.Codec, storeService, s.stakingKeeper, s.slashingKeeper, addresscodec.NewBech32Codec("cosmosvaloper"))
	s.msgServer = keeper.NewMsgServerImpl(s.k)
	s.queryServer = keeper.NewQueryServerImpl(s.k)

	// register interfaces
	authtypes.RegisterInterfaces(encCfg.InterfaceRegistry)
	stakingtypes.RegisterInterfaces(encCfg.InterfaceRegistry)
	poa.RegisterInterfaces(encCfg.InterfaceRegistry)

	genState := poa.NewGenesisState()
	genState.Params.Admins = []string{s.addrs[0].String(), s.govModAddr}
	err = s.k.InitGenesis(s.ctx, genState)
	require.NoError(t, err)

	s.createBaseStakingValidators(t)
	return s
}

type valSetup struct {
	priv   *secp256k1.PrivKey
	addr   sdk.AccAddress
	valKey *ed25519.PrivKey
}

func GenAcc() valSetup {
	priv1 := secp256k1.GenPrivKey()
	addr1 := sdk.AccAddress(priv1.PubKey().Address())
	valKey1 := ed25519.GenPrivKey()

	return valSetup{
		priv:   priv1,
		addr:   addr1,
		valKey: valKey1,
	}
}

func (f *testFixture) createBaseStakingValidators(t *testing.T) {
	bondCoin := sdk.NewCoin("stake", math.NewInt(1_000_000))

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
			Power:            1_000_000,
			Unsafe:           true,
		})
		require.NoError(t, err)

		// increase the block so the new validator is in the validator set
		f.ctx = f.ctx.WithBlockHeight(f.ctx.BlockHeight() + 1)
		_, err = f.stakingKeeper.ApplyAndReturnValidatorSetUpdates(f.ctx)
		require.NoError(t, err)

		valAddrBz, err := sdk.ValAddressFromBech32(val.GetOperator())
		require.NoError(t, err)

		validator, err := f.stakingKeeper.GetValidator(f.ctx, valAddrBz)
		require.NoError(t, err)

		validator.Status = stakingtypes.Bonded
		if err := f.stakingKeeper.SetValidator(f.ctx, validator); err != nil {
			panic(err)
		}

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
		Tokens:          math.NewInt(amt),
		DelegatorShares: math.LegacyNewDecFromInt(math.NewInt(amt)),
		Description:     poa.NewDescription(moniker, "", "", "", ""),
		UnbondingHeight: 0,
		UnbondingTime:   time.Time{},
		Commission: poa.Commission{
			CommissionRates: poa.NewCommissionRates(math.LegacyZeroDec(), math.LegacyZeroDec(), math.LegacyZeroDec()),
		},
		MinSelfDelegation:       math.OneInt(),
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

func (f *testFixture) IncreaseBlock(amt int64) ([]abci.ValidatorUpdate, error) {
	f.ctx = f.ctx.WithBlockHeight(f.ctx.BlockHeight() + amt)
	updates, err := f.stakingKeeper.ApplyAndReturnValidatorSetUpdates(f.ctx)
	return updates, err
}
