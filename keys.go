package poa

import (
	"cosmossdk.io/collections"
)

var (
	ParamsKey            = collections.NewPrefix(0)
	PendingValidatorsKey = collections.NewPrefix(1)
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
