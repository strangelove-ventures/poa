# Module Integration

## Design Details
- Wraps x/staking for XYZ
- delegation is forced from the validators account.
- validators can not unbond or redelegate. Can only go down 3 ways: (x/slashing module): downtime, doubleslash or (x/poa) admin removal.

- Ante: All staking commands except `MsgUpdateValidator`
- Ante: Commission limits (forced range, or specific value)

- If you want a module's control not to be based on governance, update your `app.go` authorities to use your set account instead of the gov account by default. This could be useful for the Upgrade module to not require governance but still allow the chain to get upgrades.

## Integration

- Add module in
- Add module to `app/app.go` (ensure it is before the stakingtypes since that runs the `BeginBlocker` at current Height)


