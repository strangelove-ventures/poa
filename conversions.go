package poa

import (
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func NewDescription(moniker, identity, website, securityContact, details string) Description {
	return Description{
		Moniker:         moniker,
		Identity:        identity,
		Website:         website,
		SecurityContact: securityContact,
		Details:         details,
	}
}

func NewCommissionRates(rate, maxRate, maxChangeRate math.LegacyDec) CommissionRates {
	return CommissionRates{
		Rate:          rate,
		MaxRate:       maxRate,
		MaxChangeRate: maxChangeRate,
	}
}

// TODO: ideally we should remove this need?
func ConvertPOAToStaking(poa Validator) types.Validator {
	return types.Validator{
		OperatorAddress: poa.OperatorAddress,
		ConsensusPubkey: poa.ConsensusPubkey,
		Jailed:          poa.Jailed,
		Status:          types.BondStatus(poa.Status),
		Tokens:          poa.Tokens,
		DelegatorShares: poa.DelegatorShares,
		Description: types.NewDescription(
			poa.Description.Moniker,
			poa.Description.Identity,
			poa.Description.Website,
			poa.Description.SecurityContact,
			poa.Description.Details,
		),
		UnbondingHeight: poa.UnbondingHeight,
		UnbondingTime:   poa.UnbondingTime,
		Commission: types.NewCommission(
			poa.Commission.CommissionRates.Rate,
			poa.Commission.CommissionRates.MaxRate,
			poa.Commission.CommissionRates.MaxChangeRate,
		),
		MinSelfDelegation:       poa.MinSelfDelegation,
		UnbondingOnHoldRefCount: poa.UnbondingOnHoldRefCount,
		UnbondingIds:            poa.UnbondingIds,
	}
}

func ConvertStakingToPOA(val types.Validator) Validator {
	return Validator{
		OperatorAddress: val.OperatorAddress,
		ConsensusPubkey: val.ConsensusPubkey,
		Jailed:          val.Jailed,
		Status:          BondStatus(val.Status),
		Tokens:          val.Tokens,
		DelegatorShares: val.DelegatorShares,
		Description: Description{
			Moniker:         val.Description.Moniker,
			Identity:        val.Description.Identity,
			Website:         val.Description.Website,
			SecurityContact: val.Description.SecurityContact,
			Details:         val.Description.Details,
		},
		UnbondingHeight: val.UnbondingHeight,
		UnbondingTime:   val.UnbondingTime,
		Commission: Commission{
			CommissionRates: CommissionRates{
				Rate:          val.Commission.Rate,
				MaxRate:       val.Commission.MaxRate,
				MaxChangeRate: val.Commission.MaxChangeRate,
			},
		},
		MinSelfDelegation:       val.MinSelfDelegation,
		UnbondingOnHoldRefCount: val.UnbondingOnHoldRefCount,
		UnbondingIds:            val.UnbondingIds,
	}
}
