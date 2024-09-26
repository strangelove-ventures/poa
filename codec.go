package poa

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	types "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers the necessary interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgSetPower{}, "poa/MsgSetPower")
	legacy.RegisterAminoMsg(cdc, &MsgCreateValidator{}, "poa/MsgCreateValidator")
	legacy.RegisterAminoMsg(cdc, &MsgRemoveValidator{}, "poa/MsgRemoveValidator")
	legacy.RegisterAminoMsg(cdc, &MsgRemovePending{}, "poa/MsgRemovePending")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateStakingParams{}, "poa/MsgUpdateStakingParams")
}

// RegisterInterfaces registers the interfaces types with the interface registry.
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgSetPower{},
		&MsgCreateValidator{},
		&MsgRemoveValidator{},
		&MsgRemovePending{},
		&MsgUpdateStakingParams{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
