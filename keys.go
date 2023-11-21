package poa

import (
	"cosmossdk.io/collections"
)

var (
	ParamsKey            = collections.NewPrefix(0)
	PendingValidatorsKey = collections.NewPrefix(1)

	// CachedPreviousBlockPowerKey saves the previous blocks total delegated power amount.
	CachedPreviousBlockPowerKey = collections.NewPrefix(2)

	// AbsoluteChangedInBlockPowerKey tracks the current blocks total delegated power amount.
	// If this becomes >30% of CachedPreviousBlockPowerKey, messages will fail to limit IBC issues.
	AbsoluteChangedInBlockPowerKey = collections.NewPrefix(3)
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
