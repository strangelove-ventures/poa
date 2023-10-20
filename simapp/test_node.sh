#!/bin/bash
#
# Example:
# cd simapp
# BINARY="poad" CHAIN_ID="poa-1" HOME_DIR="$HOME/.poad" TIMEOUT_COMMIT="500ms" CLEAN=true sh test_node.sh
#
# poad tx poa set-power $(poad q staking validators --output=json | jq .validators[0].operator_address -r) 1230000 --home=$HOME_DIR --yes --from=acc1
# poad q staking validators
# poad tx poa remove $(poad q staking validators --output=json | jq .validators[0].operator_address -r) --home=$HOME_DIR --yes --from=acc1
#
# validate staking msg err
# poad tx staking delegate $(poad q staking validators --output=json | jq .validators[0].operator_address -r) 1stake --home=$HOME_DIR --yes --from=acc1
#
# Create a validator
# poad tx poa create-validator simapp/validator_file.json --from acc3 --yes --home=$HOME_DIR # no genesis amount
# poad q poa pending-validators --output json
# poad tx poa set-power $(poad q poa pending-validators --output=json | jq .pending[0].operator_address -r) 123 --home=$HOME_DIR --yes --from=acc1
# poad tx poa remove $(poad q staking validators --output=json | jq .validators[1].operator_address -r) --home=$HOME_DIR --yes --from=acc1

export KEY="acc1" # validator
export KEY2="acc2"
export KEY2="acc3"

export CHAIN_ID=${CHAIN_ID:-"local-1"}
export MONIKER="moniker"
export KEYALGO="secp256k1"
export KEYRING=${KEYRING:-"test"}
export HOME_DIR=$(eval echo "${HOME_DIR:-"~/.simapp"}")
export BINARY=${BINARY:-simd}

export CLEAN=${CLEAN:-"false"}
export RPC=${RPC:-"26657"}
export REST=${REST:-"1317"}
export PROFF=${PROFF:-"6060"}
export P2P=${P2P:-"26656"}
export GRPC=${GRPC:-"9090"}
export GRPC_WEB=${GRPC_WEB:-"9091"}
export ROSETTA=${ROSETTA:-"8080"}
export TIMEOUT_COMMIT=${TIMEOUT_COMMIT:-"5s"}

alias BINARY="$BINARY --home=$HOME_DIR"

command -v $BINARY > /dev/null 2>&1 || { echo >&2 "$BINARY command not found. Ensure this is setup / properly installed in your GOPATH (make install)."; exit 1; }
command -v jq > /dev/null 2>&1 || { echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"; exit 1; }

BINARY config set client keyring-backend $KEYRING
BINARY config set client chain-id $CHAIN_ID

# poad tx poa set-power acc1 $(poad q staking validators --output=json | jq .validators[0].operator_address -r) 5000000 --home=$HOME_DIR

from_scratch () {
  # Fresh install on current branch
  go install ./...
  status=$?
  if [ $status -ne 0 ]; then
    echo "Failed to install binary"
    exit $status
  fi


  # remove existing daemon.
  rm -rf $HOME_DIR && echo "Removed $HOME_DIR"  

  # cosmos1hj5fveer5cjtn4wd6wstzugjfdxzl0xpxvjjvr
  echo "decorate bright ozone fork gallery riot bus exhaust worth way bone indoor calm squirrel merry zero scheme cotton until shop any excess stage laundry" | BINARY keys add $KEY --keyring-backend $KEYRING --algo $KEYALGO --recover
  # cosmos1efd63aw40lxf3n4mhf7dzhjkr453axur6cpk92
  echo "wealth flavor believe regret funny network recall kiss grape useless pepper cram hint member few certain unveil rather brick bargain curious require crowd raise" | BINARY keys add $KEY2 --keyring-backend $KEYRING --algo $KEYALGO --recover
  # cosmos1evcfka7s3200ypj0k2449cujlq3u4xe850hc4w
  echo "year action hospital impulse repeat town caught glue palace guilt diet about melt outdoor orbit field income left visit client route float wife media" | BINARY keys add $KEY3 --keyring-backend $KEYRING --algo $KEYALGO --recover
  
  # TODO: move from stake to another denom, this works for now though.
  BINARY init $MONIKER --chain-id $CHAIN_ID --default-denom stake

  # Function updates the config based on a jq argument as a string
  update_test_genesis () {    
    cat $HOME_DIR/config/genesis.json | jq "$1" > $HOME_DIR/config/tmp_genesis.json && mv $HOME_DIR/config/tmp_genesis.json $HOME_DIR/config/genesis.json
  }

  # Block
  update_test_genesis '.consensus_params["block"]["max_gas"]="100000000"'
  # Gov
  update_test_genesis '.app_state["gov"]["params"]["min_deposit"]=[{"denom": "stake","amount": "1000000"}]'
  update_test_genesis '.app_state["gov"]["voting_params"]["voting_period"]="15s"'
  # staking
  update_test_genesis '.app_state["staking"]["params"]["bond_denom"]="stake"'  
  update_test_genesis '.app_state["staking"]["params"]["min_commission_rate"]="0.050000000000000000"'  
  # mint
  update_test_genesis '.app_state["mint"]["params"]["mint_denom"]="stake"'  
  # crisis
  update_test_genesis '.app_state["crisis"]["constant_fee"]={"denom": "stake","amount": "1000"}'  

  # x/POA
  # allows gov & acc1 to perform actions on POA.
  update_test_genesis '.app_state["poa"]["params"]["admins"]=["cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn","cosmos1hj5fveer5cjtn4wd6wstzugjfdxzl0xpxvjjvr"]'  

  # Allocate genesis accounts 
  # stake should ONLY be as much as they gentx with. No more.
  BINARY genesis add-genesis-account $KEY 1000000stake,1000utest --keyring-backend $KEYRING
  BINARY genesis add-genesis-account $KEY2 1000000stake,1000utest --keyring-backend $KEYRING  

  # 1 power (these rates will be overwriten)
  BINARY genesis gentx $KEY 1000000stake --keyring-backend $KEYRING --chain-id $CHAIN_ID --commission-rate="0.05" --commission-max-rate="0.50"

  # Collect genesis tx
  BINARY genesis collect-gentxs

  # Run this to ensure junorything worked and that the genesis file is setup correctly
  BINARY genesis validate-genesis
}

# check if CLEAN is not set to false
if [ "$CLEAN" != "false" ]; then
  echo "Starting from a clean state"
  from_scratch
fi

echo "Starting node..."

# Opens the RPC endpoint to outside connections
sed -i 's/laddr = "tcp:\/\/127.0.0.1:26657"/c\laddr = "tcp:\/\/0.0.0.0:'$RPC'"/g' $HOME_DIR/config/config.toml
sed -i 's/cors_allowed_origins = \[\]/cors_allowed_origins = \["\*"\]/g' $HOME_DIR/config/config.toml

# REST endpoint
sed -i 's/address = "tcp:\/\/localhost:1317"/address = "tcp:\/\/0.0.0.0:'$REST'"/g' $HOME_DIR/config/app.toml
sed -i 's/enable = false/enable = true/g' $HOME_DIR/config/app.toml

# replace pprof_laddr = "localhost:6060" binding
sed -i 's/pprof_laddr = "localhost:6060"/pprof_laddr = "localhost:'$PROFF_LADDER'"/g' $HOME_DIR/config/config.toml

# change p2p addr laddr = "tcp://0.0.0.0:26656"
sed -i 's/laddr = "tcp:\/\/0.0.0.0:26656"/laddr = "tcp:\/\/0.0.0.0:'$P2P'"/g' $HOME_DIR/config/config.toml

# GRPC
sed -i 's/address = "localhost:9090"/address = "0.0.0.0:'$GRPC'"/g' $HOME_DIR/config/app.toml
sed -i 's/address = "localhost:9091"/address = "0.0.0.0:'$GRPC_WEB'"/g' $HOME_DIR/config/app.toml

# Rosetta Api
sed -i 's/address = ":8080"/address = "0.0.0.0:'$ROSETTA'"/g' $HOME_DIR/config/app.toml

# faster blocks
sed -i 's/timeout_commit = "5s"/timeout_commit = "'$TIMEOUT_COMMIT'"/g' $HOME_DIR/config/config.toml

# Start the node with 0 gas fees
BINARY start --pruning=nothing  --minimum-gas-prices=0stake,0utest --rpc.laddr="tcp://0.0.0.0:$RPC"