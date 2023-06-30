package module

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	examplev1 "github.com/cosmosregistry/example/api/v1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: examplev1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Counter",
					Use:       "counter [address]",
					Short:     "Get the current value of the counter for an address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "address"},
					},
				},
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Get the current module parameters",
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: examplev1.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "IncrementCounter",
					Use:       "counter [sender]",
					Short:     "Increments the counter by 1 for the sender",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "sender"},
					},
				},
				// The UpdateParams tx is purposely left empty, the MsgUpdateParams is gov gated.
			},
		},
	}
}
