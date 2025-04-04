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

rm -rf $HOMEDIR

# Initializing chain...
$BINARY init $MONIKER --chain-id $CHAIN_ID --overwrite
$BINARY config keyring-backend $KEYRING_BACKEND --home "$HOMEDIR"
$BINARY config chain-id $CHAINID --home "$HOMEDIR"

# Adding keys...
echo "cup hip eyebrow flock slogan filter gas tent angle purpose rose setup" | $BINARY keys add $KEY_NAME --recover --keyring-backend $KEYRING_BACKEND --home "$HOMEDIR"
echo "offer feel open ancient relax habit field right evoke ball organ beauty" | $BINARY keys add test1 --recover  --keyring-backend $KEYRING_BACKEND --home "$HOMEDIR"
echo "olympic such citizen any bind small neutral hidden prefer pupil trash lemon" | $BINARY keys add test2 --recover  --keyring-backend $KEYRING_BACKEND --home "$HOMEDIR"

# Adding genesis account...
$BINARY add-genesis-account $($BINARY keys show $KEY_NAME -a) $STAKE_AMOUNT

# Creating gentx...
$BINARY gentx $KEY_NAME $STAKE_AMOUNT --chain-id $CHAIN_ID

# Collecting gentx...
$BINARY collect-gentxs

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

# Add vesting accounts
$(dirname "$0")/misc/genesis-vesting.sh $(date +%s)

# Validating genesis file...
$BINARY validate-genesis

# Starting the chain...
$BINARY start
