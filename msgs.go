package poa

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"

	"cosmossdk.io/math"
)

var (
	_ codectypes.UnpackInterfacesMessage = (*MsgCreateValidator)(nil)
	_ codectypes.UnpackInterfacesMessage = (*Validator)(nil)
)

// NewMsgCreateValidator creates a new MsgCreateValidator instance.
func NewMsgCreateValidator(
	valAddr string, pubKey cryptotypes.PubKey, description Description, commission CommissionRates, minSelfDelegation math.Int,
) (*MsgCreateValidator, error) {
	var pkAny *codectypes.Any
	if pubKey != nil {
		var err error
		if pkAny, err = codectypes.NewAnyWithValue(pubKey); err != nil {
			return nil, err
		}
	}

	return &MsgCreateValidator{
		Description:       description,
		ValidatorAddress:  valAddr,
		Pubkey:            pkAny,
		Commission:        commission,
		MinSelfDelegation: minSelfDelegation,
	}, nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgCreateValidator) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var pubKey cryptotypes.PubKey

	return unpacker.UnpackAny(msg.Pubkey, &pubKey)
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (v Validator) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var pk cryptotypes.PubKey

	return unpacker.UnpackAny(v.ConsensusPubkey, &pk)
}
