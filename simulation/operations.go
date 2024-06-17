package simulation

import (
	"math/rand"

	"github.com/pkg/errors"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	sdkmath "cosmossdk.io/math"

	poatypes "github.com/strangelove-ventures/poa"
	"github.com/strangelove-ventures/poa/keeper"
)

const (
	OpWeightMsgSetPower               = "op_weight_msg_set_power"                // nolint: gosec
	OpWeightMsgRemoveValidator        = "op_weight_msg_remove_validator"         // nolint: gosec
	OpWeightMsgRemovePendingValidator = "op_weight_msg_remove_pending_validator" // nolint: gosec
	OpWeightMsgUpdateParams           = "op_weight_msg_update_params"            // nolint: gosec
	OpWeightMsgCreateValidator        = "op_weight_msg_create_validator"         // nolint: gosec
	OpWeightMsgUpdateStakingParams    = "op_weight_msg_update_staking_params"    // nolint: gosec

	DefaultWeightMsgSetPower               = 100
	DefaultWeightMsgRemoveValidator        = 5
	DefaultWeightMsgRemovePendingValidator = 100
	DefaultWeightMsgUpdateParams           = 85
	DefaultWeightMsgCreateValidator        = 100
	DefaultWeightMsgUpdateStakingParams    = 50
)

// WeightedOperations returns the all the gov module operations with their respective weights.
func WeightedOperations(appParams simtypes.AppParams,
	cdc codec.JSONCodec,
	txGen client.TxConfig,
	k keeper.Keeper) []simtypes.WeightedOperation {
	operations := make([]simtypes.WeightedOperation, 0)

	var weightMsgSetPower int
	appParams.GetOrGenerate(OpWeightMsgSetPower, &weightMsgSetPower, nil, func(r *rand.Rand) { weightMsgSetPower = DefaultWeightMsgSetPower })
	appParams.GetOrGenerate(OpWeightMsgRemoveValidator, &weightMsgSetPower, nil, func(r *rand.Rand) { weightMsgSetPower = DefaultWeightMsgRemoveValidator })
	appParams.GetOrGenerate(OpWeightMsgRemovePendingValidator, &weightMsgSetPower, nil, func(r *rand.Rand) { weightMsgSetPower = DefaultWeightMsgRemovePendingValidator })
	appParams.GetOrGenerate(OpWeightMsgUpdateParams, &weightMsgSetPower, nil, func(r *rand.Rand) { weightMsgSetPower = DefaultWeightMsgUpdateParams })
	appParams.GetOrGenerate(OpWeightMsgCreateValidator, &weightMsgSetPower, nil, func(r *rand.Rand) { weightMsgSetPower = DefaultWeightMsgCreateValidator })
	appParams.GetOrGenerate(OpWeightMsgUpdateStakingParams, &weightMsgSetPower, nil, func(r *rand.Rand) { weightMsgSetPower = DefaultWeightMsgUpdateStakingParams })

	operations = append(operations, simulation.NewWeightedOperation(weightMsgSetPower, SimulateMsgSetPower(txGen, k)))
	operations = append(operations, simulation.NewWeightedOperation(weightMsgSetPower, SimulateMsgRemoveValidator(k)))
	operations = append(operations, simulation.NewWeightedOperation(weightMsgSetPower, SimulateMsgRemovePendingValidator(txGen, k)))
	operations = append(operations, simulation.NewWeightedOperation(weightMsgSetPower, SimulateMsgUpdateParams(k)))
	operations = append(operations, simulation.NewWeightedOperation(weightMsgSetPower, SimulateMsgCreateValidator(txGen, k)))
	operations = append(operations, simulation.NewWeightedOperation(weightMsgSetPower, SimulateMsgUpdateStakingParams(k)))

	return operations
}

func SimulateMsgUpdateStakingParams(k keeper.Keeper) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		return simtypes.NoOpMsg(poatypes.ModuleName, OpWeightMsgUpdateStakingParams, "placeholder"), nil, nil
	}
}

func SimulateMsgCreateValidator(txGen client.TxConfig, k keeper.Keeper) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msgType := sdk.MsgTypeURL(&poatypes.MsgCreateValidator{})
		simAccount, _ := simtypes.RandomAcc(r, accs)
		address := sdk.ValAddress(simAccount.Address)

		// ensure the validator doesn't exist already
		_, err := k.GetStakingKeeper().GetValidator(ctx, address)
		if err == nil {
			return simtypes.NoOpMsg(poatypes.ModuleName, msgType, "validator already exists"), nil, nil
		}

		consPubKey := sdk.GetConsAddress(simAccount.ConsKey.PubKey())
		_, err = k.GetStakingKeeper().GetValidatorByConsAddr(ctx, consPubKey)
		if err == nil {
			return simtypes.NoOpMsg(poatypes.ModuleName, msgType, "cons key already used"), nil, nil
		}

		denom, err := k.GetStakingKeeper().BondDenom(ctx)
		if err != nil {
			return simtypes.NoOpMsg(poatypes.ModuleName, msgType, "bond denom not found"), nil, err
		}

		balance := k.GetBankKeeper().GetBalance(ctx, simAccount.Address, denom)
		if !balance.IsPositive() {
			return simtypes.NoOpMsg(poatypes.ModuleName, msgType, "balance is negative"), nil, nil
		}

		amount, err := simtypes.RandPositiveInt(r, balance.Amount)
		if err != nil {
			return simtypes.NoOpMsg(poatypes.ModuleName, msgType, "unable to generate positive amount"), nil, err
		}

		selfDelegation := sdk.NewCoin(denom, amount)

		description := poatypes.Description{
			Moniker:         simtypes.RandStringOfLength(r, 10),
			Identity:        simtypes.RandStringOfLength(r, 10),
			Website:         simtypes.RandStringOfLength(r, 10),
			SecurityContact: simtypes.RandStringOfLength(r, 10),
			Details:         simtypes.RandStringOfLength(r, 10),
		}

		// The commission rate floor and ceil is set to [0.1, 0.5] in the POA simapp AnteHandler
		minCommissionInt := 10
		maxCommissionInt := simtypes.RandIntBetween(r, minCommissionInt, 50)
		// Ensure that maxCommissionInt is different from minCommissionInt
		for maxCommissionInt == minCommissionInt {
			maxCommissionInt = simtypes.RandIntBetween(r, minCommissionInt, 50)
		}
		rateInt := simtypes.RandIntBetween(r, minCommissionInt, maxCommissionInt)
		maxChangeInt := simtypes.RandIntBetween(r, minCommissionInt, maxCommissionInt)
		maxCommission := sdkmath.LegacyNewDecWithPrec(int64(maxCommissionInt), 2) // Random between 0.1 and 0.5
		rate := sdkmath.LegacyNewDecWithPrec(int64(rateInt), 2)                   // Random between 0.1 and maxCommission
		maxChangeRate := sdkmath.LegacyNewDecWithPrec(int64(maxChangeInt), 2)     // Random between 0.1 and maxCommission

		commission := poatypes.CommissionRates{
			Rate:          rate,
			MaxRate:       maxCommission,
			MaxChangeRate: maxChangeRate,
		}

		msg, err := poatypes.NewMsgCreateValidator(address.String(), simAccount.ConsKey.PubKey(), description, commission, selfDelegation.Amount)
		if err != nil {
			return simtypes.NoOpMsg(poatypes.ModuleName, msgType, errors.WithMessage(err, "unable to create MsgCreateValidator").Error()), nil, err
		}

		txCtx := simulation.OperationInput{
			R:             r,
			App:           app,
			TxGen:         txGen,
			Cdc:           nil,
			Msg:           msg,
			Context:       ctx,
			SimAccount:    simAccount,
			AccountKeeper: k.GetAccountKeeper(),
			Bankkeeper:    k.GetBankKeeper(),
			ModuleName:    poatypes.ModuleName,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

func SimulateMsgUpdateParams(k keeper.Keeper) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		return simtypes.NoOpMsg(poatypes.ModuleName, OpWeightMsgUpdateParams, "placeholder"), nil, nil
	}
}

func SimulateMsgRemovePendingValidator(txGen client.TxConfig, k keeper.Keeper) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msgType := sdk.MsgTypeURL(&poatypes.MsgRemovePending{})
		pending, err := k.GetPendingValidators(ctx)
		if err != nil {
			return simtypes.NoOpMsg(poatypes.ModuleName, msgType, "unable to get pending validators"), nil, err
		}

		if len(pending.Validators) == 0 {
			return simtypes.NoOpMsg(poatypes.ModuleName, msgType, "no pending validators"), nil, nil
		}

		// Pick a random pending validator
		pendingIdx := r.Intn(len(pending.Validators))
		pendingVal := pending.Validators[pendingIdx]
		valAddr := pendingVal.OperatorAddress

		admin, err := getRandomPOAAdmin(r, k.GetAdmins(ctx))
		if err != nil {
			return simtypes.NoOpMsg(poatypes.ModuleName, msgType, err.Error()), nil, err
		}

		// Verify that the POA admin is a simulation accounts
		adminAcc, found := simtypes.FindAccount(accs, sdk.MustAccAddressFromBech32(admin))

		if !found {
			return simtypes.NoOpMsg(poatypes.ModuleName, msgType, "admin not found in simulator accounts"), nil, nil
		}

		adminAddr := adminAcc.Address

		// Generate random transaction fees
		spendable := k.GetBankKeeper().SpendableCoins(ctx, adminAddr)
		fees, err := simtypes.RandomFees(r, ctx, spendable)
		if err != nil {
			return simtypes.NoOpMsg(poatypes.ModuleName, msgType, "error generating random fees"), nil, err
		}

		msg := poatypes.MsgRemovePending{
			Sender:           adminAddr.String(),
			ValidatorAddress: valAddr,
		}

		txCtx := simulation.OperationInput{
			R:             r,
			App:           app,
			TxGen:         txGen,
			Cdc:           nil,
			Msg:           &msg,
			Context:       ctx,
			SimAccount:    adminAcc,
			AccountKeeper: k.GetAccountKeeper(),
			Bankkeeper:    k.GetBankKeeper(),
			ModuleName:    poatypes.ModuleName,
		}

		return simulation.GenAndDeliverTx(txCtx, fees)
	}
}

func SimulateMsgRemoveValidator(k keeper.Keeper) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		return simtypes.NoOpMsg(poatypes.ModuleName, OpWeightMsgRemoveValidator, "placeholder"), nil, nil
	}
}

func SimulateMsgSetPower(txGen client.TxConfig, k keeper.Keeper) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msgType := sdk.MsgTypeURL(&poatypes.MsgSetPower{})

		validators, err := k.GetStakingKeeper().GetAllValidators(ctx)
		if err != nil {
			return simtypes.NoOpMsg(poatypes.ModuleName, msgType, "unable to get validators"), nil, err
		}
		if len(validators) == 0 {
			return simtypes.NoOpMsg(poatypes.ModuleName, msgType, "no validators found"), nil, nil
		}

		// Get the power reduction value used to convert tokens to consensus power
		powerReduction := k.GetStakingKeeper().PowerReduction(ctx)

		// Compute the total power of all validators
		totalPower := int64(0)
		for _, val := range validators {
			totalPower += val.GetConsensusPower(powerReduction)
		}

		// Select a random validator to update
		validator := validators[r.Intn(len(validators))]

		// Compute the new power of the validator
		minPower := 1_000_000 // 1 Consensus Power = 1_000_000 shares by default

		// The new power needs to be <= totalPower * 0.3 (30% of the total power)
		maxPower := int(float64(totalPower) * 0.3)
		if maxPower < minPower {
			return simtypes.NoOpMsg(poatypes.ModuleName, msgType, "total power too low"), nil, nil
		}

		// Generate a safe random power
		newPower := uint64(simtypes.RandIntBetween(r, minPower, maxPower))

		// Check if the new power is the same as the current power
		// If it is, return a no-op
		ttcp := sdk.TokensToConsensusPower(sdkmath.NewIntFromUint64(newPower), powerReduction)
		if validator.GetConsensusPower(powerReduction) == ttcp {
			return simtypes.NoOpMsg(poatypes.ModuleName, msgType, "same power"), nil, nil
		}

		// Select a random POA admin
		admin, err := getRandomPOAAdmin(r, k.GetAdmins(ctx))
		if err != nil {
			return simtypes.NoOpMsg(poatypes.ModuleName, msgType, err.Error()), nil, err
		}

		// Verify that the POA admin is a simulation accounts
		adminAcc, found := simtypes.FindAccount(accs, sdk.MustAccAddressFromBech32(admin))
		if !found {
			return simtypes.NoOpMsg(poatypes.ModuleName, msgType, "admin not found in simulator accounts"), nil, nil
		}

		// Generate random transaction fees
		adminAddr := adminAcc.Address
		spendable := k.GetBankKeeper().SpendableCoins(ctx, adminAddr)
		fees, err := simtypes.RandomFees(r, ctx, spendable)
		if err != nil {
			return simtypes.NoOpMsg(poatypes.ModuleName, msgType, "error generating random fees"), nil, err
		}

		msg := poatypes.MsgSetPower{
			Sender:           admin,
			ValidatorAddress: validator.OperatorAddress,
			Power:            newPower,
			Unsafe:           false, // We only cover the case where the power is <= 30% of the total power
		}

		txCtx := simulation.OperationInput{
			R:             r,
			App:           app,
			TxGen:         txGen,
			Cdc:           nil,
			Msg:           &msg,
			Context:       ctx,
			SimAccount:    adminAcc,
			AccountKeeper: k.GetAccountKeeper(),
			Bankkeeper:    k.GetBankKeeper(),
			ModuleName:    poatypes.ModuleName,
		}

		return simulation.GenAndDeliverTx(txCtx, fees)
	}
}

func getRandomPOAAdmin(r *rand.Rand, admins []string) (string, error) {
	if len(admins) == 0 {
		return "", errors.New("no admins found")
	}

	idx := r.Intn(len(admins))
	admin := admins[idx]

	return admin, nil
}
