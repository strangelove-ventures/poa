package simulation

import (
	"math"
	"math/rand"
	"sort"

	"github.com/pkg/errors"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	sdkmath "cosmossdk.io/math"

	poatypes "github.com/strangelove-ventures/poa"
	"github.com/strangelove-ventures/poa/keeper"
)

const (
	OpWeightMsgPOASetPower               = "op_weight_msg_poa_set_power"                // nolint: gosec
	OpWeightMsgPOARemoveValidator        = "op_weight_msg_poa_remove_validator"         // nolint: gosec
	OpWeightMsgPOARemovePendingValidator = "op_weight_msg_poa_remove_pending_validator" // nolint: gosec
	OpWeightMsgPOAUpdateParams           = "op_weight_msg_poa_update_params"            // nolint: gosec
	OpWeightMsgPOACreateValidator        = "op_weight_msg_poa_create_validator"         // nolint: gosec

)

var (
	weights = map[string]int{
		OpWeightMsgPOASetPower:               100,
		OpWeightMsgPOARemoveValidator:        20,
		OpWeightMsgPOARemovePendingValidator: 100,
		OpWeightMsgPOAUpdateParams:           85,
		OpWeightMsgPOACreateValidator:        100,
	}
)

// WeightedOperations returns all the poa module operations with their respective weights.
func WeightedOperations(appParams simtypes.AppParams,
	cdc codec.JSONCodec,
	txGen client.TxConfig,
	k keeper.Keeper) []simtypes.WeightedOperation {
	operations := make([]simtypes.WeightedOperation, 0)

	// Iterating over a map is non-deterministic in Go. To simulate determinism, we sort the keys in a slice and iterate over them.
	keys := make([]string, 0, len(weights))
	for k := range weights {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, op := range keys {
		defaultWeight := weights[op]
		var weight int
		appParams.GetOrGenerate(op, &weight, nil, func(r *rand.Rand) { weight = defaultWeight })
		operations = append(operations, getWeightedOperation(op, weight, txGen, k))
	}

	return operations
}

func getWeightedOperation(op string, weight int, txGen client.TxConfig, k keeper.Keeper) simtypes.WeightedOperation {
	switch op {
	case OpWeightMsgPOASetPower:
		return simulation.NewWeightedOperation(weight, SimulateMsgSetPower(txGen, k))
	case OpWeightMsgPOARemoveValidator:
		return simulation.NewWeightedOperation(weight, SimulateMsgRemoveValidator(txGen, k))
	case OpWeightMsgPOARemovePendingValidator:
		return simulation.NewWeightedOperation(weight, SimulateMsgRemovePendingValidator(txGen, k))
	case OpWeightMsgPOAUpdateParams:
		return simulation.NewWeightedOperation(weight, SimulateMsgUpdateParams(txGen, k))
	case OpWeightMsgPOACreateValidator:
		return simulation.NewWeightedOperation(weight, SimulateMsgCreateValidator(txGen, k))
	default:
		panic("invalid poa operation type")
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
		if balance.Amount.LT(sdkmath.NewInt(2)) {
			return simtypes.NoOpMsg(poatypes.ModuleName, msgType, "balance is less than two"), nil, nil
		}

		// Generate a random self-delegation amount between 1 and balance
		amount := sdkmath.NewInt(r.Int63n(balance.Amount.Int64()-1) + 1)
		if err != nil {
			return simtypes.NoOpMsg(poatypes.ModuleName, msgType, "unable to generate positive amount"), nil, err
		}

		selfDelegation := sdk.NewCoin(denom, amount)
		description := generateRandomDescription(r)
		commission := generateRandomCommission(r)

		msg, err := poatypes.NewMsgCreateValidator(address.String(), simAccount.ConsKey.PubKey(), description, commission, selfDelegation.Amount)
		if err != nil {
			return simtypes.NoOpMsg(poatypes.ModuleName, msgType, errors.WithMessage(err, "unable to create MsgCreateValidator").Error()), nil, err
		}

		return genAndDeliverTxWithRandFees(r, app, ctx, txGen, simAccount, msg, k)
	}
}

func SimulateMsgUpdateParams(txGen client.TxConfig, k keeper.Keeper) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msgType := sdk.MsgTypeURL(&poatypes.MsgUpdateParams{})

		params, err := k.GetParams(ctx)
		if err != nil {
			return simtypes.NoOpMsg(poatypes.ModuleName, msgType, "unable to get params"), nil, err
		}

		admins := params.GetAdmins()
		if len(admins) < 2 {
			return simtypes.NoOpMsg(poatypes.ModuleName, msgType, "not enough admins found"), nil, nil
		}

		// Remove a random admin from the list
		idx := r.Intn(len(admins))
		admins[idx] = admins[len(admins)-1]
		admins = admins[:len(admins)-1]

		adminAcc, err := selectRandomPOAAccount(r, admins, accs)
		if err != nil {
			return simtypes.NoOpMsg(poatypes.ModuleName, msgType, err.Error()), nil, err
		}

		msg := poatypes.MsgUpdateParams{
			Sender: adminAcc.Address.String(),
			Params: poatypes.Params{
				Admins:                 admins,
				AllowValidatorSelfExit: r.Intn(2) == 1,
			},
		}

		return genAndDeliverTxWithRandFees(r, app, ctx, txGen, adminAcc, &msg, k)
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

		// Pick a random pending validator address
		valAddr := pending.Validators[r.Intn(len(pending.Validators))].OperatorAddress

		adminAcc, err := selectRandomPOAAccount(r, k.GetAdmins(ctx), accs)
		if err != nil {
			return simtypes.NoOpMsg(poatypes.ModuleName, msgType, err.Error()), nil, err
		}

		msg := poatypes.MsgRemovePending{
			Sender:           adminAcc.Address.String(),
			ValidatorAddress: valAddr,
		}

		return genAndDeliverTxWithRandFees(r, app, ctx, txGen, adminAcc, &msg, k)
	}
}

func SimulateMsgRemoveValidator(txGen client.TxConfig, k keeper.Keeper) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msgType := sdk.MsgTypeURL(&poatypes.MsgRemoveValidator{})

		validators, err := k.GetStakingKeeper().GetBondedValidatorsByPower(ctx)
		if err != nil {
			return simtypes.NoOpMsg(poatypes.ModuleName, msgType, "unable to get bonded validators"), nil, err
		}
		if len(validators) == 0 {
			return simtypes.NoOpMsg(poatypes.ModuleName, msgType, "no bonded validators found"), nil, nil
		}
		if len(validators) == 1 {
			return simtypes.NoOpMsg(poatypes.ModuleName, msgType, "cannot remove the last validator in the set"), nil, nil
		}

		// Select a random bonded validator to remove
		validator := validators[r.Intn(len(validators))]

		adminAcc, err := selectRandomPOAAccount(r, k.GetAdmins(ctx), accs)
		if err != nil {
			return simtypes.NoOpMsg(poatypes.ModuleName, msgType, err.Error()), nil, err
		}

		power := validator.GetConsensusPower(k.GetStakingKeeper().PowerReduction(ctx))
		if power == 0 {
			return simtypes.NoOpMsg(poatypes.ModuleName, msgType, "validator has no power"), nil, nil
		}

		msg := poatypes.MsgRemoveValidator{
			Sender:           adminAcc.Address.String(),
			ValidatorAddress: validator.OperatorAddress,
		}

		return genAndDeliverTxWithRandFees(r, app, ctx, txGen, adminAcc, &msg, k)
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

		// Select a random validator
		validator := validators[r.Intn(len(validators))]

		// Get the new power for the validator
		newPower, errStr, isError := getNewPower(r, k, ctx, validator)
		if errStr != "" {
			if isError {
				return simtypes.NoOpMsg(poatypes.ModuleName, msgType, errStr), nil, errors.New(errStr)
			}
			return simtypes.NoOpMsg(poatypes.ModuleName, msgType, errStr), nil, nil
		}

		adminAcc, err := selectRandomPOAAccount(r, k.GetAdmins(ctx), accs)
		if err != nil {
			return simtypes.NoOpMsg(poatypes.ModuleName, msgType, err.Error()), nil, err
		}

		// Generate random transaction fees
		msg := poatypes.MsgSetPower{
			Sender:           adminAcc.Address.String(),
			ValidatorAddress: validator.OperatorAddress,
			Power:            newPower,
			Unsafe:           false, // We only cover the case where the power is <= 30% of the total power
		}

		return genAndDeliverTxWithRandFees(r, app, ctx, txGen, adminAcc, &msg, k)
	}
}

// getNewPower returns a random new power value for a validator.
func getNewPower(r *rand.Rand, k keeper.Keeper, ctx sdk.Context, validator stakingtypes.Validator) (uint64, string, bool) {
	// Get the total power of the *previous* block
	cachedPower, err := k.GetCachedBlockPower(ctx)
	if err != nil {
		return 0, "unable to get cached block power", true
	}

	// Get the total power changed in the *current* block
	totalChangedPower, err := k.GetAbsoluteChangedInBlockPower(ctx)
	if err != nil {
		return 0, "unable to get absolute changed in block power", true
	}

	valAddr, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
	if err != nil {
		return 0, "unable to get validator address", true
	}

	// Get the current power of the validator
	currentPower, err := k.GetStakingKeeper().GetLastValidatorPower(ctx, valAddr)
	if err != nil {
		return 0, "unable to get last validator power", true
	}

	// Get the power reduction value used to convert tokens to consensus power
	powerReduction := k.GetStakingKeeper().PowerReduction(ctx)

	// Compute the new power of the validator
	const minPower = 1 // 1 Consensus Power = 1_000_000 shares by default

	// The new power is between 1 and 30% of the total power of the *previous* block
	// Changes over 30% of the total power are considered unsafe
	// See the `unsafe` flag of `MsgSetPower`
	maxPower := int(float64(cachedPower) * 0.3) // Decimal places are truncated
	if maxPower < minPower {
		return 0, "total power too low", false
	}

	// Generate a random new power value
	newPower := simtypes.RandIntBetween(r, minPower, maxPower)

	// No change in power
	if currentPower == int64(newPower) {
		return 0, "same power", false
	}

	// Compute the absolute change for this operation, i.e., how much the power of the validator will change
	absPowerDiff := uint64(math.Abs(float64(newPower) - float64(currentPower)))

	// The power change for the *entire block* needs to be below 30%
	percentChange := ((absPowerDiff + totalChangedPower) * 100) / cachedPower
	if percentChange >= 30 {
		return 0, "unsafe power", false
	}

	// Convert the new power to tokens
	newPowerTokens := sdk.TokensFromConsensusPower(int64(newPower), powerReduction)

	return newPowerTokens.Uint64(), "", false
}

func selectRandomPOAAccount(r *rand.Rand, admins []string, accs []simtypes.Account) (simtypes.Account, error) {
	randomAdminAddr, err := getRandomPOAAdmin(r, admins)
	if err != nil {
		return simtypes.Account{}, err
	}

	acc, found := simtypes.FindAccount(accs, sdk.MustAccAddressFromBech32(randomAdminAddr))
	if !found {
		return simtypes.Account{}, errors.New("admin not found in simulator accounts")
	}

	return acc, nil
}

func getRandomPOAAdmin(r *rand.Rand, admins []string) (string, error) {
	if len(admins) == 0 {
		return "", errors.New("no admins found")
	}

	idx := r.Intn(len(admins))
	admin := admins[idx]

	return admin, nil
}

func generateRandomDescription(r *rand.Rand) poatypes.Description {
	return poatypes.Description{
		Moniker:         simtypes.RandStringOfLength(r, 10),
		Identity:        simtypes.RandStringOfLength(r, 10),
		Website:         simtypes.RandStringOfLength(r, 10),
		SecurityContact: simtypes.RandStringOfLength(r, 10),
		Details:         simtypes.RandStringOfLength(r, 10),
	}
}

func generateRandomCommission(r *rand.Rand) poatypes.CommissionRates {
	minCommissionInt := 10
	maxCommissionInt := simtypes.RandIntBetween(r, minCommissionInt, 50)
	for maxCommissionInt == minCommissionInt {
		maxCommissionInt = simtypes.RandIntBetween(r, minCommissionInt, 50)
	}
	rateInt := simtypes.RandIntBetween(r, minCommissionInt, maxCommissionInt)
	maxChangeInt := simtypes.RandIntBetween(r, minCommissionInt, maxCommissionInt)

	return poatypes.CommissionRates{
		Rate:          sdkmath.LegacyNewDecWithPrec(int64(rateInt), 2),
		MaxRate:       sdkmath.LegacyNewDecWithPrec(int64(maxCommissionInt), 2),
		MaxChangeRate: sdkmath.LegacyNewDecWithPrec(int64(maxChangeInt), 2),
	}
}

func newOperationInput(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, txGen client.TxConfig, simAccount simtypes.Account, msg sdk.Msg, k keeper.Keeper) simulation.OperationInput {
	return simulation.OperationInput{
		R:             r,
		App:           app,
		TxGen:         txGen,
		Cdc:           nil,
		Msg:           msg,
		Context:       ctx,
		SimAccount:    simAccount,
		AccountKeeper: k.GetTestAccountKeeper(),
		Bankkeeper:    k.GetBankKeeper(),
		ModuleName:    poatypes.ModuleName,
	}
}

func genAndDeliverTxWithRandFees(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, txGen client.TxConfig, simAccount simtypes.Account, msg sdk.Msg, k keeper.Keeper) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
	return simulation.GenAndDeliverTxWithRandFees(newOperationInput(r, app, ctx, txGen, simAccount, msg, k))
}
