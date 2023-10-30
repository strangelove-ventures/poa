package poaante

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/strangelove-ventures/poa"
	"github.com/stretchr/testify/require"
)

func TestCommissionRanges(t *testing.T) {
	testCases := []struct {
		name       string
		rateFloor  math.LegacyDec
		rateCeil   math.LegacyDec
		commission poa.CommissionRates
		expPass    bool
	}{
		{
			name:      "fail: rate < floor",
			rateFloor: math.LegacyMustNewDecFromStr("0.10"),
			rateCeil:  math.LegacyMustNewDecFromStr("0.50"),
			commission: poa.CommissionRates{
				Rate: math.LegacyMustNewDecFromStr("0.01"),
			},
			expPass: false,
		},
		{
			name:      "fail: rate > ceil",
			rateFloor: math.LegacyMustNewDecFromStr("0.10"),
			rateCeil:  math.LegacyMustNewDecFromStr("0.50"),
			commission: poa.CommissionRates{
				Rate: math.LegacyMustNewDecFromStr("0.51"),
			},
			expPass: false,
		},
		{
			name:      "fail: rate != ceil and floor",
			rateFloor: math.LegacyMustNewDecFromStr("0.15"),
			rateCeil:  math.LegacyMustNewDecFromStr("0.15"),
			commission: poa.CommissionRates{
				Rate: math.LegacyMustNewDecFromStr("0.14"),
			},
			expPass: false,
		},
		{
			name:      "pass: rate == ceil and floor",
			rateFloor: math.LegacyMustNewDecFromStr("0.15"),
			rateCeil:  math.LegacyMustNewDecFromStr("0.15"),
			commission: poa.CommissionRates{
				Rate: math.LegacyMustNewDecFromStr("0.15"),
			},
			expPass: true,
		},
		{
			name:      "pass: rate == floor",
			rateFloor: math.LegacyMustNewDecFromStr("0.15"),
			rateCeil:  math.LegacyMustNewDecFromStr("0.50"),
			commission: poa.CommissionRates{
				Rate: math.LegacyMustNewDecFromStr("0.15"),
			},
			expPass: true,
		},
		{
			name:      "pass: rate == ceil",
			rateFloor: math.LegacyMustNewDecFromStr("0.15"),
			rateCeil:  math.LegacyMustNewDecFromStr("0.50"),
			commission: poa.CommissionRates{
				Rate: math.LegacyMustNewDecFromStr("0.50"),
			},
			expPass: true,
		},
	}

	for _, tc := range testCases {
		tc := tc

		m := NewMsgCommissionLimiterDecorator(true, tc.rateFloor, tc.rateCeil)

		// Creating a Validator
		err := m.hasInvalidCommissionRange([]sdk.Msg{&poa.MsgCreateValidator{
			Commission: tc.commission,
		}})
		if tc.expPass {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
		}

		// Editing a Validator
		err = m.hasInvalidCommissionRange([]sdk.Msg{&stakingtypes.MsgEditValidator{
			CommissionRate: &tc.commission.Rate,
		}})
		if tc.expPass {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
		}
	}
}
