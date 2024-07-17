package poa

import (
	"encoding/json"
	fmt "fmt"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// DefaultParams returns default module parameters.
func DefaultParams() Params {
	return Params{
		Admins:                 []string(nil), // uses the authority as admin only now
		AllowValidatorSelfExit: true,
	}
}

// NewParams returns a new POA Params.
func NewParams(allowValSelfExit bool) (Params, error) {
	p := Params{
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
	if len(p.Admins) != 0 {
		return fmt.Errorf("DEPERECATED: admins must be empty as only the keeper authority is used")
	}

	return nil
}
