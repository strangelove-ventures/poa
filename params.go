package poa

import (
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// DefaultParams returns the default x/staking parameters.
func DefaultStakingParams() StakingParams {
	sp := stakingtypes.DefaultParams()

	return StakingParams{
		UnbondingTime:     sp.UnbondingTime,
		MaxValidators:     sp.MaxValidators,
		MaxEntries:        sp.MaxEntries,
		HistoricalEntries: sp.HistoricalEntries,
		BondDenom:         sp.BondDenom,
		MinCommissionRate: sp.MinCommissionRate,
	}
}
