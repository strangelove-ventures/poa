package poa

import (
	"cosmossdk.io/collections"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// TODO: migrate to the new collections API

var (
	ValidatorsKey         = collections.NewPrefix(0)                                      // prefix for each key to a validator account key
	ValidatorsByConsKey   = collections.NewPrefix(1)                                      // prefix for each key to a validator consensus key
	VouchesKey            = collections.NewPrefix(2)                                      // prefix for each key to a vouch
	VouchesByValidatorKey = collections.NewPrefix(3)                                      // prefix for each key to a validator
	ParamsKey             = collections.NewPrefix(4)                                      // prefix for the param store
	HistoricalInfoKey     = collections.PairPrefix[int64, stakingtypes.HistoricalInfo](5) // prefix for historical entries
)

const (
	// ModuleName is the name of the module
	ModuleName = "poa"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey to be used for routing msgs
	RouterKey = ModuleName

	// QuerierRoute to be used for querier msgs
	QuerierRoute = ModuleName

	// TransientStoreKey defines the transient store key
	TransientStoreKey = "transient_" + ModuleName
)
