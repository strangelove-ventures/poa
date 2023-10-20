package poa

import (
	sdkerrors "cosmossdk.io/errors"
)

var (
	ErrStakingActionNotAllowed = sdkerrors.Register(ModuleName, 1, "staking actions are now allowed on this chain")
)
