package keeper_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
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
	bankkeeper     bankkeeper.Keeper

	addrs      []sdk.AccAddress
	govModAddr string
}

func SetupTest(t *testing.T) *testFixture {
	s := new(testFixture)

	encCfg := moduletestutil.MakeTestEncodingConfig()
	key := storetypes.NewKVStoreKey(poa.ModuleName)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	s.ctx = testCtx.Ctx

	storeService := runtime.NewKVStoreService(key)
	s.addrs = simtestutil.CreateIncrementalAccounts(3)

	// TODO: gomock initializations ?
	// ctrl := gomock.NewController(s.T())
	// s.stakingKeeper = slashingtestutil.NewMockStakingKeeper(ctrl)
	// s.stakingKeeper.EXPECT().ValidatorAddressCodec().Return(address.NewBech32Codec("cosmosvaloper")).AnyTimes()
	// s.stakingKeeper.EXPECT().ConsensusAddressCodec().Return(address.NewBech32Codec("cosmosvalcons")).AnyTimes()

	logger := log.NewTestLogger(t)
	s.govModAddr = authtypes.NewModuleAddress(govtypes.ModuleName).String()

	s.accountkeeper = authkeeper.NewAccountKeeper(encCfg.Codec, storeService, authtypes.ProtoBaseAccount, maccPerms, authcodec.NewBech32Codec(sdk.Bech32MainPrefix), sdk.Bech32MainPrefix, s.govModAddr)
	s.bankkeeper = bankkeeper.NewBaseKeeper(encCfg.Codec, storeService, s.accountkeeper, nil, s.govModAddr, logger)

	s.stakingKeeper = stakingkeeper.NewKeeper(encCfg.Codec, storeService, s.accountkeeper, s.bankkeeper, s.govModAddr, authcodec.NewBech32Codec(sdk.Bech32PrefixValAddr), authcodec.NewBech32Codec(sdk.Bech32PrefixConsAddr))
	s.stakingKeeper.SetParams(s.ctx, stakingtypes.DefaultParams())

	s.slashingKeeper = slashingkeeper.NewKeeper(encCfg.Codec, encCfg.Amino, storeService, s.stakingKeeper, authtypes.NewModuleAddress(govtypes.ModuleName).String())
	s.slashingKeeper.SetParams(s.ctx, slashingtypes.DefaultParams())

	s.k = keeper.NewKeeper(encCfg.Codec, storeService, s.stakingKeeper, s.slashingKeeper, addresscodec.NewBech32Codec("cosmosvaloper"))
	s.msgServer = keeper.NewMsgServerImpl(s.k)
	s.queryServer = keeper.NewQueryServerImpl(s.k)

	// register interfaces
	authtypes.RegisterInterfaces(encCfg.InterfaceRegistry)
	stakingtypes.RegisterInterfaces(encCfg.InterfaceRegistry)
	poa.RegisterInterfaces(encCfg.InterfaceRegistry)

	genState := poa.NewGenesisState()
	genState.Params.Admins = []string{s.addrs[0].String(), s.govModAddr}
	s.k.InitGenesis(s.ctx, genState)

	s.createBaseValidators(t)
	return s
}

type valSetup struct {
	priv   *secp256k1.PrivKey
	addr   sdk.AccAddress
	valKey *ed25519.PrivKey
}

func genAcc() valSetup {
	priv1 := secp256k1.GenPrivKey()
	addr1 := sdk.AccAddress(priv1.PubKey().Address())
	valKey1 := ed25519.GenPrivKey()

	return valSetup{
		priv:   priv1,
		addr:   addr1,
		valKey: valKey1,
	}
}

func (f *testFixture) createBaseValidators(t *testing.T) {
	stakingMsgServer := stakingkeeper.NewMsgServerImpl(f.stakingKeeper)
	bondCoin := sdk.NewCoin("stake", math.NewInt(1_000_000))

	vals := []valSetup{
		genAcc(),
		genAcc(),
		genAcc(),
	}

	for idx, val := range vals {
		description := stakingtypes.NewDescription(fmt.Sprintf("foo-%d", idx), "", "", "", "")
		commissionRates := stakingtypes.NewCommissionRates(math.LegacyZeroDec(), math.LegacyZeroDec(), math.LegacyZeroDec())

		valAddr := sdk.ValAddress(val.addr).String()

		// mint tokens to validator
		f.bankkeeper.MintCoins(f.ctx, poa.ModuleName, sdk.NewCoins(bondCoin))
		f.bankkeeper.SendCoinsFromModuleToAccount(f.ctx, poa.ModuleName, val.addr, sdk.NewCoins(bondCoin))

		// create validator
		msg, err := stakingtypes.NewMsgCreateValidator(
			valAddr, val.valKey.PubKey(), bondCoin, description, commissionRates, math.OneInt(),
		)
		require.NoError(t, err)

		_, err = stakingMsgServer.CreateValidator(f.ctx, msg)
		require.NoError(t, err)

		// increase the block so the new validator is in the validator set
		f.ctx = f.ctx.WithBlockHeight(f.ctx.BlockHeight() + 1)

		_, err = f.stakingKeeper.ApplyAndReturnValidatorSetUpdates(f.ctx)
		require.NoError(t, err)
	}

}
