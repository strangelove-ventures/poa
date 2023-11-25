package poa

import (
	"testing"

	"github.com/stretchr/testify/suite"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	prefix = "/strangelove_ventures.poa.v1."
)

type CodecTestSuite struct {
	suite.Suite
}

func TestCodecSuite(t *testing.T) {
	suite.Run(t, new(CodecTestSuite))
}

func (suite *CodecTestSuite) TestRegisterInterfaces() {
	registry := codectypes.NewInterfaceRegistry()
	registry.RegisterInterface(sdk.MsgInterfaceProtoName, (*sdk.Msg)(nil))
	RegisterInterfaces(registry)

	impls := registry.ListImplementations(sdk.MsgInterfaceProtoName)

	suite.Require().Len(impls, 5)
	suite.Require().ElementsMatch([]string{
		prefix + "MsgSetPower",
		prefix + "MsgCreateValidator",
		prefix + "MsgUpdateParams",
		prefix + "MsgRemoveValidator",
		prefix + "MsgUpdateStakingParams",
	}, impls)
}
