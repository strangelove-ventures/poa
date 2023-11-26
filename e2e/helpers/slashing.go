package helpers

import (
	"context"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
)

// poad --node=http://localhost:33835 q slashing signing-infos --output=json
// {
//   "info": [
//     {
//       "address": "cosmosvalcons1qjm654u45hlf72503tuu96zq4vx44sm09p35zl",
//       "index_offset": "45",
//       "jailed_until": "1970-01-01T00:00:00Z"
//     },
//     {
//       "address": "cosmosvalcons1q66ygfwf46ph342j65zf29aayjvrx79rfeg9f0",
//       "index_offset": "45",
//       "jailed_until": "1970-01-01T00:00:00Z"
//     },
//     {
//       "address": "cosmosvalcons1zuaq7dsvauln9wv93umcra5a4z377au7sc2u8r",
//       "jailed_until": "2023-11-26T20:52:11.679151873Z"
//     },
//     {
//       "address": "cosmosvalcons1xq7qkgy6yfr7p4t223tkf6zewn3t0kccj3vzsu",
//       "index_offset": "45",
//       "jailed_until": "1970-01-01T00:00:00Z"
//     },
//     {
//       "address": "cosmosvalcons17ra0e2mzv8g74gemn0vzqm483efmuf7axjndq8",
//       "index_offset": "45",
//       "jailed_until": "1970-01-01T00:00:00Z"
//     }
//   ],
//   "pagination": {
//     "total": "5"
//   }
// }

// func GetSigningInformation(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, valoper string) (TxResponse, error) {
// 	cmd := TxCommandBuilder(ctx, chain, []string{"tx", "distribution", "withdraw-rewards", valoper}, user)
// 	return ExecuteTransaction(ctx, chain, cmd)
// }

type SingingInformation struct {
	Info []struct {
		Address     string `json:"address"`
		IndexOffset string `json:"index_offset,omitempty"`
		JailedUntil string `json:"jailed_until"`
	} `json:"info"`
	Pagination struct {
		Total string `json:"total"`
	} `json:"pagination"`
}

func GetSigningInformation(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain) SingingInformation {
	var res SingingInformation
	ExecuteQuery(ctx, chain, []string{"query", "slashing", "signing-infos"}, &res)
	return res
}
