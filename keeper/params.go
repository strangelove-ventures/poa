package keeper

import (
	"context"

	"github.com/strangelove-ventures/poa"
)

// SetParams sets the module parameters.
func (k Keeper) SetParams(ctx context.Context, p poa.Params) error {
	if err := p.Validate(); err != nil {
		return err
	}

	return k.Params.Set(ctx, p)
}

// GetParams returns the current module parameters.
func (k Keeper) GetParams(ctx context.Context) (poa.Params, error) {
	p, err := k.Params.Get(ctx)
	if err != nil {
		return poa.DefaultParams(), err
	}

	return p, nil
}
