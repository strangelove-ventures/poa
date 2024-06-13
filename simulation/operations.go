package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	poatypes "github.com/strangelove-ventures/poa"
	"github.com/strangelove-ventures/poa/keeper"
)

const (
	OpWeightMsgSetPower               = "op_weight_msg_set_power"
	OpWeightMsgRemoveValidator        = "op_weight_msg_remove_validator"
	OpWeightMsgRemovePendingValidator = "op_weight_msg_remove_pending_validator"
	OpWeightMsgUpdateParams           = "op_weight_msg_update_params"
	OpWeightMsgCreateValidator        = "op_weight_msg_create_validator"
	OpWeightMsgUpdateStakingParams    = "op_weight_msg_update_staking_params"

	DefaultWeightMsgSetPower               = 100
	DefaultWeightMsgRemoveValidator        = 5
	DefaultWeightMsgRemovePendingValidator = 5
	DefaultWeightMsgUpdateParams           = 85
	DefaultWeightMsgCreateValidator        = 50
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

	operations = append(operations, simulation.NewWeightedOperation(weightMsgSetPower, SimulateMsgSetPower(k)))
	operations = append(operations, simulation.NewWeightedOperation(weightMsgSetPower, SimulateMsgRemoveValidator(k)))
	operations = append(operations, simulation.NewWeightedOperation(weightMsgSetPower, SimulateMsgRemovePendingValidator(k)))
	operations = append(operations, simulation.NewWeightedOperation(weightMsgSetPower, SimulateMsgUpdateParams(k)))
	operations = append(operations, simulation.NewWeightedOperation(weightMsgSetPower, SimulateMsgCreateValidator(k)))
	operations = append(operations, simulation.NewWeightedOperation(weightMsgSetPower, SimulateMsgUpdateStakingParams(k)))

	return operations
}

func SimulateMsgUpdateStakingParams(k keeper.Keeper) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		return simtypes.NoOpMsg(poatypes.ModuleName, OpWeightMsgUpdateStakingParams, "placeholder"), nil, nil
	}
}

func SimulateMsgCreateValidator(k keeper.Keeper) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		return simtypes.NoOpMsg(poatypes.ModuleName, OpWeightMsgCreateValidator, "placeholder"), nil, nil
	}
}

func SimulateMsgUpdateParams(k keeper.Keeper) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		return simtypes.NoOpMsg(poatypes.ModuleName, OpWeightMsgUpdateParams, "placeholder"), nil, nil
	}
}

func SimulateMsgRemovePendingValidator(k keeper.Keeper) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		return simtypes.NoOpMsg(poatypes.ModuleName, OpWeightMsgRemovePendingValidator, "placeholder"), nil, nil
	}
}

func SimulateMsgRemoveValidator(k keeper.Keeper) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		return simtypes.NoOpMsg(poatypes.ModuleName, OpWeightMsgRemoveValidator, "placeholder"), nil, nil
	}
}

func SimulateMsgSetPower(k keeper.Keeper) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		return simtypes.NoOpMsg(poatypes.ModuleName, OpWeightMsgSetPower, "placeholder"), nil, nil
	}

}
