package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/strangelove-ventures/poa"
)

func TestAddPending(t *testing.T) {
	f := SetupTest(t, 2_000_000)
	require := require.New(t)

	val := GenAcc()
	valAddr := sdk.ValAddress(val.addr)
	v := poa.ConvertPOAToStaking(CreateNewValidator(
		"myval",
		valAddr.String(),
		val.valKey.PubKey(),
		int64(1_000_000),
	))

	// successful add
	err := f.k.AddPendingValidator(f.ctx, v, val.valKey.PubKey())
	require.NoError(err)

	// duplicate (fails)
	err = f.k.AddPendingValidator(f.ctx, v, val.valKey.PubKey())
	require.Error(err)
	require.Equal(poa.ErrValidatorAlreadyPending, err)

	// remove pending
	err = f.k.RemovePendingValidator(f.ctx, v.OperatorAddress)
	require.NoError(err)

	pending, err := f.k.GetPendingValidators(f.ctx)
	require.NoError(err)
	require.Empty(pending.Validators)
}
