package poa_test

import (
	"testing"
	time "time"

	"github.com/stretchr/testify/suite"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"cosmossdk.io/math"

	"github.com/strangelove-ventures/poa"
)

type ConversionTestSuite struct {
	suite.Suite
}

func TestConversionSuite(t *testing.T) {
	suite.Run(t, new(ConversionTestSuite))
}

func (suite *ConversionTestSuite) TestPOAStakingConversions() {
	// description
	description := poa.NewDescription("moniker", "identity", "website", "securityContact", "details")
	expectedDesc := poa.Description{
		Moniker:         "moniker",
		Identity:        "identity",
		Website:         "website",
		SecurityContact: "securityContact",
		Details:         "details",
	}
	suite.Require().Equal(expectedDesc, description)

	// commissions
	commissions := poa.NewCommissionRates(
		math.LegacyMustNewDecFromStr("0.1"),
		math.LegacyMustNewDecFromStr("0.2"),
		math.LegacyMustNewDecFromStr("0.3"),
	)
	expectedComm := poa.CommissionRates{
		Rate:          math.LegacyMustNewDecFromStr("0.1"),
		MaxRate:       math.LegacyMustNewDecFromStr("0.2"),
		MaxChangeRate: math.LegacyMustNewDecFromStr("0.3"),
	}
	suite.Require().Equal(expectedComm, commissions)

	// poa -> staking
	poaVal := poa.Validator{
		OperatorAddress: "operatorAddress",
		ConsensusPubkey: nil,
		Jailed:          false,
		Status:          poa.Bonded,
		Tokens:          math.NewInt(1000000),
		DelegatorShares: math.LegacyMustNewDecFromStr("1000000"),
		Description:     description,
		UnbondingHeight: 1,
		UnbondingTime:   time.Time{},
		Commission: poa.Commission{
			CommissionRates: commissions,
		},
		MinSelfDelegation:       math.OneInt(),
		UnbondingOnHoldRefCount: 1,
		UnbondingIds:            []uint64{1},
	}

	stakingVal := poa.ConvertPOAToStaking(poaVal)

	expectedStakingVal := stakingtypes.Validator{
		OperatorAddress: "operatorAddress",
		ConsensusPubkey: nil,
		Jailed:          false,
		Status:          stakingtypes.Bonded,
		Tokens:          math.NewInt(1000000),
		DelegatorShares: math.LegacyMustNewDecFromStr("1000000"),
		Description: stakingtypes.NewDescription(
			"moniker",
			"identity",
			"website",
			"securityContact",
			"details",
		),
		UnbondingHeight: 1,
		UnbondingTime:   time.Time{},
		Commission: stakingtypes.NewCommission(
			math.LegacyMustNewDecFromStr("0.1"),
			math.LegacyMustNewDecFromStr("0.2"),
			math.LegacyMustNewDecFromStr("0.3"),
		),
		MinSelfDelegation:       math.OneInt(),
		UnbondingOnHoldRefCount: 1,
		UnbondingIds:            []uint64{1},
	}
	suite.Require().Equal(expectedStakingVal, stakingVal)

	// staking -> poa
	newPoaVal := poa.ConvertStakingToPOA(stakingVal)
	suite.Require().Equal(poaVal, newPoaVal)
}
