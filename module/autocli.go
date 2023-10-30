package module

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	poav1 "github.com/strangelove-ventures/poa/api/v1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service:           poav1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				// {
				// 	RpcMethod: "Params",
				// 	Use:       "params",
				// 	Short:     "Get the current module parameters",
				// },
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: poav1.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "SetPower",
					Use:       "set-power [validator] [power]",
					Short:     "Sets a validators network power",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "sender"},
						{ProtoField: "validator"},
						{ProtoField: "power"},
					},
				},
				{
					RpcMethod: "RemoveValidator",
					Use:       "remove [validator]",
					Short:     "Remove an active validator from the set",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "validator"},
					},
				},
				{
					RpcMethod: "CreateValidator",
					Use:       "create-validator [validator]",
					Short:     "Remove a pending validator for the set",
				},
				{
					RpcMethod: "UpdateParams",
					Use:       "update-params [admin1,admin2,admin3,...]",
					Short:     "Update the params for the module",
				},
			},
		},
	}
}
