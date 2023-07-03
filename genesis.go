package poa

import (
	"encoding/base64"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewGenesisState creates a new genesis state with default values.
func NewGenesisState() *GenesisState {
	return &GenesisState{
		Validators: []*Validator{},
		Params:     DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
func (gs *GenesisState) Validate() error {
	valAddressMap := make(map[string]struct{})
	valPubKeyMap := make(map[string]struct{})

	for _, val := range gs.Validators {
		// Check for duplicated validator address
		address := sdk.ValAddress(val.Address).String()
		if _, ok := valAddressMap[address]; ok {
			return fmt.Errorf("duplicated validator address: %s", address)
		}
		valAddressMap[address] = struct{}{}

		// Check for duplicated pub key
		pubKey := base64.StdEncoding.EncodeToString(val.Pubkey.Value)
		if _, ok := valPubKeyMap[pubKey]; ok {
			return fmt.Errorf("duplicated validator pub key: %s", pubKey)
		}
		valPubKeyMap[pubKey] = struct{}{}
	}

	vouchesMap := make(map[string]struct{})

	for _, vouch := range gs.Vouches {
		// Check for duplicated vouches
		vouchr := sdk.ValAddress(vouch.VoucherAddress).String()
		candidate := sdk.ValAddress(vouch.CandidateAddress).String()
		vouchrCandidateKey := vouchr + candidate
		if _, ok := vouchesMap[vouchrCandidateKey]; ok {
			return fmt.Errorf("duplicated vouch from vouchr: %s for candidate: %s", vouchr, candidate)
		}
		valAddressMap[vouchrCandidateKey] = struct{}{}
	}

	return gs.Params.Validate()
}
