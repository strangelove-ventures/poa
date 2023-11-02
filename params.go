package poa

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// DefaultParams returns default module parameters.
func DefaultParams() Params {
	govModuleAddress := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	return Params{
		Admins: []string{govModuleAddress},
	}
}

// NewParams returns a new POA Params.
func NewParams(addresses []string) (Params, error) {
	if len(addresses) == 0 {
		return Params{}, ErrMustProvideAtLeastOneAddress
	}

	for _, address := range addresses {
		if _, err := sdk.AccAddressFromBech32(address); err != nil {
			return Params{}, err
		}
	}

	return Params{
		Admins: addresses,
	}, nil
}

// DefaultParams returns default x/staking parameters.
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

// add the stringer method for Params
func (p Params) String() string {
	bz, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}

	return string(bz)
}

// Validate does the sanity check on the params.
func (p Params) Validate() error {
	for _, auth := range p.Admins {
		if _, err := sdk.AccAddressFromBech32(auth); err != nil {
			return err
		}
	}

	return nil
}
