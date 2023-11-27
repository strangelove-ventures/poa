package poa

import (
	"encoding/json"
	fmt "fmt"

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
func NewParams(admins []string, allowValSelfExit bool) (Params, error) {
	p := Params{
		Admins:                 admins,
		AllowValidatorSelfExit: allowValSelfExit,
	}

	return p, p.Validate()
}

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

// Stringer method for Params.
func (p Params) String() string {
	bz, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}

	return string(bz)
}

// Validate does the sanity check on the params.
func (p Params) Validate() error {
	return validateAdmins(p.Admins)
}

func validateAdmins(i interface{}) error {
	admins, ok := i.([]string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if len(admins) == 0 {
		return ErrMustProvideAtLeastOneAddress
	}

	for _, auth := range admins {
		if _, err := sdk.AccAddressFromBech32(auth); err != nil {
			return err
		}
	}

	return nil
}
