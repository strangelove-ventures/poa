package e2e

import (
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/poa"
)

var (
	POAImage = ibc.DockerImage{
		Repository: "poa",
		Version:    "local",
		UidGid:     "1025:1025",
	}

	POACfg = ibc.ChainConfig{
		Images: []ibc.DockerImage{
			POAImage,
		},
		ModifyGenesis: cosmos.ModifyGenesis([]cosmos.GenesisKV{
			{
				Key: "app_state.poa.params.admins",
				Value: []string{
					"cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn", // gov
					"cosmos1hj5fveer5cjtn4wd6wstzugjfdxzl0xpxvjjvr", // testing account
				},
			},
		}),
		// TODO: modify gentxs / genesis account amounts?
		EncodingConfig: poaEncoding(),
		Type:           "cosmos",
		Name:           "poa",
		ChainID:        "poa-1",
		Bin:            "poad",
		Bech32Prefix:   "cosmos",
		Denom:          "stake", // maybe poa?
		CoinType:       "118",
		GasPrices:      "0stake,0utest",
		TrustingPeriod: "330h",
	}
)

func poaEncoding() *moduletestutil.TestEncodingConfig {
	cfg := cosmos.DefaultEncoding()
	poa.RegisterInterfaces(cfg.InterfaceRegistry)
	return &cfg
}
