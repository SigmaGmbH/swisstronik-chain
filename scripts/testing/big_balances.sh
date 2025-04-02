#!/bin/bash

set -e

CHAIN_ID="swisstronik_1291-1"
MONIKER="myvalidator"
STAKE_AMOUNT="310000swtr"
KEY_NAME="validator"
HOMEDIR="$HOME/.swisstronik"
BINARY="$(dirname "$0")/../../build/swisstronikd"
KEYRING_BACKEND="test"
APP_TOML=$HOMEDIR/config/app.toml
GENESIS=$HOMEDIR/config/genesis.json
TMP_GENESIS=$HOMEDIR/config/tmp_genesis.json

TREASURY_KEY_NAME="treasury"
TREASURY_AMOUNT="3100000000swtr"

rm -rf $HOMEDIR

# Initializing chain...
$BINARY init $MONIKER --chain-id $CHAIN_ID --overwrite
$BINARY config keyring-backend $KEYRING_BACKEND --home "$HOMEDIR"
$BINARY config chain-id $CHAINID --home "$HOMEDIR"

# Adding keys...
$BINARY keys add $KEY_NAME --keyring-backend $KEYRING_BACKEND
$BINARY keys add $TREASURY_KEY_NAME --keyring-backend $KEYRING_BACKEND

# Adding genesis account...
$BINARY add-genesis-account $($BINARY keys show $KEY_NAME -a) $STAKE_AMOUNT
$BINARY add-genesis-account $($BINARY keys show $TREASURY_KEY_NAME -a) $TREASURY_AMOUNT

# Creating gentx...
$BINARY gentx $KEY_NAME $STAKE_AMOUNT --chain-id $CHAIN_ID

# Collecting gentx...
$BINARY collect-gentxs

# Validating genesis file...
$BINARY validate-genesis

# Set min gas price
sed -i 's/minimum-gas-prices = ""/minimum-gas-prices = "0aswtr"/' "$APP_TOML"

jq '.app_state["feemarket"]["params"]["base_fee"]="7"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["staking"]["params"]["bond_denom"]="aswtr"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["staking"]["params"]["unbonding_time"]="1s"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["crisis"]["constant_fee"]["denom"]="aswtr"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="aswtr"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["gov"]["params"]["min_deposit"][0]["denom"]="aswtr"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["evm"]["params"]["evm_denom"]="aswtr"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["inflation"]["params"]["mint_denom"]="aswtr"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["mint"]["params"]["mint_denom"]="aswtr"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.consensus_params["block"]["max_gas"]="10000000"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["compliance"]["operators"]=[{"operator":"swtr1ml2knanpk8sv94f8h9g8vaf9k3yyfva4fykyn9", "operator_type": 1}]' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# Starting the chain...
$BINARY start
