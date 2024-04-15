package poa

import (
	"cosmossdk.io/collections"
)

var (
	// ParamsKey saves the current module params.
	ParamsKey = collections.NewPrefix(0)

	// PendingValidatorsKey saves the current pending validators.
	PendingValidatorsKey = collections.NewPrefix(1)

	// CachedPreviousBlockPowerKey saves the previous blocks total power amount.
	CachedPreviousBlockPowerKey = collections.NewPrefix(2)

	// AbsoluteChangedInBlockPowerKey tracks the current blocks total power amount.
	// If this becomes >30% of CachedPreviousBlockPowerKey, messages will fail to limit IBC issues.
	AbsoluteChangedInBlockPowerKey = collections.NewPrefix(3)

	// UpdatedValidatorsCacheKey tracks recently updated validators from SetPower.
	UpdatedValidatorsCacheKey = collections.NewPrefix(4)

	// TODO: change name
	// BeforeJailedValidatorsKey tracks validators that are about to be jailed (from staking hooks).
	BeforeJailedValidatorsKey = collections.NewPrefix(5)
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
