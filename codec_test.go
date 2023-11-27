package poa

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const prefix = "/strangelove_ventures.poa.v1."

func TestCodecRegisterInterfaces(t *testing.T) {
	registry := codectypes.NewInterfaceRegistry()
	registry.RegisterInterface(sdk.MsgInterfaceProtoName, (*sdk.Msg)(nil))
	RegisterInterfaces(registry)

	impls := registry.ListImplementations(sdk.MsgInterfaceProtoName)

	require.Len(t, impls, 6)
	require.ElementsMatch(t, []string{
		prefix + "MsgSetPower",
		prefix + "MsgCreateValidator",
		prefix + "MsgUpdateParams",
		prefix + "MsgRemoveValidator",
		prefix + "MsgRemovePending",
		prefix + "MsgUpdateStakingParams",
	}, impls)
}

func TestRegisterLegacyAminoCodec(t *testing.T) {
	cdc := codec.NewLegacyAmino()
	RegisterLegacyAminoCodec(cdc)

	bz, err := cdc.MarshalJSON(MsgSetPower{
		Sender:           "sender",
		ValidatorAddress: "validator",
		Power:            1_234_567,
		Unsafe:           true,
	})
	require.NoError(t, err)
	require.Equal(t, `{"type":"poa/MsgSetPower","value":{"sender":"sender","validator_address":"validator","power":"1234567","unsafe":true}}`, string(bz))
}
