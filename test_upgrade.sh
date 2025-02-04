#!/bin/bash
# Run this script to quickly install, setup, and run the current version, upgrade, and run the new version after migration
# run manually rather than doing a direct script run, it is pretty hacky from my Juno days
# sh ./test_upgrade.sh
# then: sh ./test_upgrade.sh doUpgrade

# run this in the root of your dir
export KEY="juno1"
export CHAINID="local-1"
export MONIKER="localjuno"
export KEYALGO="secp256k1"
export KEYRING="test"

export OLDBINARY="poad"
export NEWBINARY="poad-removed"

poad config set client keyring-backend $KEYRING
poad config set client chain-id $CHAINID

command -v poad > /dev/null 2>&1 || { echo >&2 "poad command not found. Ensure this is setup / properly installed in your GOPATH."; exit 1; }
command -v jq > /dev/null 2>&1 || { echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"; exit 1; }

doUpgrade() {
  # Run this block manually in your terminal
  $OLDBINARY tx gov submit-proposal ./draft_proposal.json --from $KEY --keyring-backend $KEYRING --chain-id $CHAINID --yes

  sleep 5

  ID="1"
  $OLDBINARY tx gov vote $ID yes --from $KEY --keyring-backend $KEYRING --chain-id $CHAINID --yes
  $OLDBINARY q gov proposal $ID

  # Wait then we will run these after it halts at height
    #   $NEWBINARY start --pruning=nothing --minimum-gas-prices=0stake
    #   $NEWBINARY tx staking delegate `$NEWBINARY q staking validators --output=json | jq .validators[0].operator_address -r` 1stake --from=$KEY --keyring-backend $KEYRING --chain-id $CHAINID --yes
    #   $NEWBINARY q staking validator `$NEWBINARY q staking validators --output=json | jq .validators[0].operator_address -r`
}

if [ "$1" = "doUpgrade" ]; then
  doUpgrade
  exit 0
fi

from_scratch () {
  # installs latest in current branch
  # make install

  # remove existing daemon.
  rm -rf ~/.simapp/*

  # juno1efd63aw40lxf3n4mhf7dzhjkr453axurv2zdzk
  echo "decorate bright ozone fork gallery riot bus exhaust worth way bone indoor calm squirrel merry zero scheme cotton until shop any excess stage laundry" | $OLDBINARY keys add $KEY --keyring-backend $KEYRING --algo $KEYALGO --recover
  # juno1hj5fveer5cjtn4wd6wstzugjfdxzl0xps73ftl
  echo "wealth flavor believe regret funny network recall kiss grape useless pepper cram hint member few certain unveil rather brick bargain curious require crowd raise" | $OLDBINARY keys add feeacc --keyring-backend $KEYRING --algo $KEYALGO --recover

  $OLDBINARY init $MONIKER --chain-id $CHAINID

  # Function updates the config based on a jq argument as a string
  update_test_genesis () {
    # update_test_genesis '.consensus_params["block"]["max_gas"]="100000000"'
    cat $HOME/.simapp/config/genesis.json | jq "$1" > $HOME/.simapp/config/tmp_genesis.json && mv $HOME/.simapp/config/tmp_genesis.json $HOME/.simapp/config/genesis.json
  }

  # Set gas limit in genesis
  update_test_genesis '.consensus_params["block"]["max_gas"]="100000000"'
  update_test_genesis '.app_state["gov"]["params"]["voting_period"]="20s"'
  update_test_genesis '.app_state["gov"]["params"]["expedited_voting_period"]="15s"'

  update_test_genesis '.app_state["staking"]["params"]["bond_denom"]="stake"'
  # update_test_genesis '.app_state["bank"]["params"]["send_enabled"]=[{"denom": "stake","enabled": true}]'
  # update_test_genesis '.app_state["staking"]["params"]["min_commission_rate"]="0.100000000000000000"' # sdk 46 only

  update_test_genesis '.app_state["mint"]["params"]["mint_denom"]="stake"'
  update_test_genesis '.app_state["gov"]["deposit_params"]["min_deposit"]=[{"denom": "stake","amount": "1"}]'
  update_test_genesis '.app_state["crisis"]["constant_fee"]={"denom": "stake","amount": "1000"}'

  # NOTE: since we test v11 -> 12, no TF setting here
  # update_test_genesis '.app_state["tokenfactory"]["params"]["denom_creation_fee"]=[{"denom":"stake","amount":"100"}]'
  # update_test_genesis '.app_state["feeshare"]["params"]["allowed_denoms"]=["stake"]'

  # Allocate genesis accounts
  $OLDBINARY genesis add-genesis-account $KEY 1000000000stake,1000utest --keyring-backend $KEYRING
  $OLDBINARY genesis add-genesis-account feeacc 1000000stake,1000utest --keyring-backend $KEYRING

  $OLDBINARY genesis gentx $KEY 1000000stake --keyring-backend $KEYRING --chain-id $CHAINID

  # Collect genesis tx
  $OLDBINARY genesis collect-gentxs

  # Run this to ensure junorything worked and that the genesis file is setup correctly
  $OLDBINARY genesis validate-genesis
}

from_scratch

echo "Starting node..."

# Opens the RPC endpoint to outside connections
sed -i '/laddr = "tcp:\/\/127.0.0.1:26657"/c\laddr = "tcp:\/\/0.0.0.0:26657"' ~/.simapp/config/config.toml
sed -i 's/cors_allowed_origins = \[\]/cors_allowed_origins = \["\*"\]/g' ~/.simapp/config/config.toml
sed -i 's/enable = false/enable = true/g' ~/.simapp/config/app.toml

$OLDBINARY start --pruning=nothing --minimum-gas-prices=0stake
