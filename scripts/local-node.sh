#!/bin/bash

CHAINID="swisstronik_1291-1"
MONIKER="localtestnet"
KEYRING="test"
KEYALGO="eth_secp256k1"
HOMEDIR="$HOME/.swisstronik"

# Path variables
CONFIG=$HOMEDIR/config/config.toml
APP_TOML=$HOMEDIR/config/app.toml
GENESIS=$HOMEDIR/config/genesis.json
TMP_GENESIS=$HOMEDIR/config/tmp_genesis.json

# validate dependencies are installed
command -v jq >/dev/null 2>&1 || {
	echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"
	exit 1
}

# used to exit on first error (any non-zero exit code)
set -e

rm -rf "$HOMEDIR"

swisstronikd config keyring-backend $KEYRING --home "$HOMEDIR"
swisstronikd config chain-id $CHAINID --home "$HOMEDIR"

echo "betray theory cargo way left cricket doll room donkey wire reunion fall left surprise hamster corn village happy bulb token artist twelve whisper expire" | swisstronikd keys add alice --keyring-backend $KEYRING --home $HOMEDIR --recover
echo "toss sense candy point cost rookie jealous snow ankle electric sauce forward oblige tourist stairs horror grunt tenant afford master violin final genre reason" | swisstronikd keys add bob --keyring-backend $KEYRING --home $HOMEDIR --recover

swisstronikd init $MONIKER -o --chain-id $CHAINID --home "$HOMEDIR"

jq '.app_state["staking"]["params"]["bond_denom"]="uswtr"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["crisis"]["constant_fee"]["denom"]="uswtr"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="uswtr"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["evm"]["params"]["evm_denom"]="uswtr"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["inflation"]["params"]["mint_denom"]="uswtr"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.consensus_params["block"]["max_gas"]="10000000"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# expose ports
sed -i 's/127.0.0.1:26657/0.0.0.0:26657/g' "$CONFIG"
sed -i 's/127.0.0.1:8545/0.0.0.0:8545/g' "$APP_TOML"
sed -i 's/127.0.0.1:8546/0.0.0.0:8546/g' "$APP_TOML"

# enable prometheus metrics
sed -i 's/prometheus = false/prometheus = true/' "$CONFIG"
sed -i 's/prometheus-retention-time  = "0"/prometheus-retention-time  = "1000000000000"/g' "$APP_TOML"
sed -i 's/enabled = false/enabled = true/g' "$APP_TOML"

# set min gas price
sed -i 's/minimum-gas-prices = ""/minimum-gas-prices = "0uswtr"/' "$APP_TOML"

# Change proposal periods to pass within a reasonable time for local testing
sed -i.bak 's/"max_deposit_period": "172800s"/"max_deposit_period": "30s"/g' "$HOMEDIR"/config/genesis.json
sed -i.bak 's/"voting_period": "172800s"/"voting_period": "30s"/g' "$HOMEDIR"/config/genesis.json

# set custom pruning settings
sed -i.bak 's/pruning = "default"/pruning = "custom"/g' "$APP_TOML"
sed -i.bak 's/pruning-keep-recent = "0"/pruning-keep-recent = "2"/g' "$APP_TOML"
sed -i.bak 's/pruning-interval = "0"/pruning-interval = "10"/g' "$APP_TOML"

# Allocate genesis accounts
swisstronikd add-genesis-account alice 10000000000000000000000000uswtr --keyring-backend $KEYRING --home "$HOMEDIR"
swisstronikd add-genesis-account bob 10000000000000000000000000uswtr --keyring-backend $KEYRING --home "$HOMEDIR"

# Sign genesis transaction
swisstronikd gentx alice 1000000000000000000000uswtr --keyring-backend $KEYRING --chain-id $CHAINID --home "$HOMEDIR"

# Collect genesis tx
swisstronikd collect-gentxs --home "$HOMEDIR"

# Run this to ensure everything worked and that the genesis file is setup correctly
swisstronikd validate-genesis --home "$HOMEDIR"

# Initialize master key
swisstronikd enclave create-master-key