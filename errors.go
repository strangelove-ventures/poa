package poa

import (
	sdkerrors "cosmossdk.io/errors"
)

var (
	ErrEmptyValidatorAddr       = sdkerrors.Register(ModuleName, 1, "empty validator address")
	ErrBadValidatorAddr         = sdkerrors.Register(ModuleName, 2, "validator address is invalid")
	ErrNoValidatorFound         = sdkerrors.Register(ModuleName, 3, "validator does not exist")
	ErrNoAcceptedValidatorFound = sdkerrors.Register(ModuleName, 4, "accepted validator does not exist")
	ErrBadValidatorPubKey       = sdkerrors.Register(ModuleName, 5, "validator pub key is invalid")
)
