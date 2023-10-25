package poa

import (
	sdkerrors "cosmossdk.io/errors"
)

var (
	ErrStakingActionNotAllowed = sdkerrors.Register(ModuleName, 1, "staking actions are now allowed on this chain")
	ErrPowerBelowMinimum       = sdkerrors.Register(ModuleName, 2, "power must be above 1_000_000")
	ErrNotAnAuthority          = sdkerrors.Register(ModuleName, 3, "not an authority")
	ErrUnsafePower             = sdkerrors.Register(ModuleName, 4, "unsafe: msg.Power is >30%% of total power, set unsafe=true to override")	
)
