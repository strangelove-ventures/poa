package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/strangelove-ventures/poa"
)

func TestPendingValidatorsQuery(t *testing.T) {
	f := SetupTest(t, 1_000_000)
	require := require.New(t)

	// Create many pending validators and query them.
	numVals := 10
	for i := 0; i < numVals; i++ {
		f.CreatePendingValidator(fmt.Sprintf("val-%d", i), 1_000_000)
	}

	// Validate pending validators equals the number we created.
	r, err := f.queryServer.PendingValidators(f.ctx, &poa.QueryPendingValidatorsRequest{})
	require.NoError(err)
	require.Len(r.Pending, numVals)

	// Accept one of the validators from pending into the active set.
	valAddr := r.Pending[0].OperatorAddress

	_, err = f.msgServer.SetPower(f.ctx, &poa.MsgSetPower{
		Sender:           f.addrs[0].String(),
		ValidatorAddress: valAddr,
		Power:            1_000_000,
		Unsafe:           true,
	})
	require.NoError(err)

	if _, err := f.IncreaseBlock(2); err != nil {
		panic(err)
	}

	// 1 less pending validator.
	r, err = f.queryServer.PendingValidators(f.ctx, &poa.QueryPendingValidatorsRequest{})
	require.NoError(err)
	require.Len(r.Pending, numVals-1)

	// none of the pending validators should be the one we accepted.
	for _, val := range r.Pending {
		require.NotEqual(valAddr, val.OperatorAddress)
	}
}
