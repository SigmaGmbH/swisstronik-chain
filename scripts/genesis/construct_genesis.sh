#!/bin/bash

set -e

CHAINID="swisstronik_1848-1"
MONIKER="localtestnet"
KEYRING="test"
KEYALGO="eth_secp256k1"
HOMEDIR="$HOME/.swisstronik"
BINARY="./build/swisstronikd"
GENESIS=$HOMEDIR/config/genesis.json
TMP_GENESIS=$HOMEDIR/config/tmp_genesis.json

if [[ -z "$SWTR_BINARY" ]]; then
  BINARY="./build/swisstronikd"
else
  BINARY="${SWTR_BINARY}"
fi

# Arachnid Deployment
ARACHNID_BYTECODE="7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe03601600081602082378035828234f58015156039578182fd5b8082525050506014600cf3"
ARACHNID_CODEHASH="0x2fa86add0aed31f33a762c9d88e807c475bd51d0f52bd0955754b2608f7e4989"

rm -rf "$HOMEDIR"

# Initial config
$BINARY config keyring-backend $KEYRING --home "$HOMEDIR"
$BINARY config chain-id $CHAINID --home "$HOMEDIR"
$BINARY init $MONIKER -o --chain-id $CHAINID --home "$HOMEDIR"

jq '.app_state["feemarket"]["params"]["base_fee"]="7"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["staking"]["params"]["unbonding_time"]="1209600s"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# Denom
jq '.app_state["staking"]["params"]["bond_denom"]="aswtr"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["crisis"]["constant_fee"]["denom"]="aswtr"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="aswtr"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["gov"]["params"]["min_deposit"][0]["denom"]="aswtr"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["evm"]["params"]["evm_denom"]="aswtr"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["inflation"]["params"]["mint_denom"]="aswtr"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["mint"]["params"]["mint_denom"]="aswtr"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# EVM params
jq '.consensus_params["block"]["max_gas"]="10000000"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# Staking params
jq '.app_state["staking"]["params"]["max_validators"]="21"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# Governance params
jq '.app_state["gov"]["deposit_params"]["min_deposit"][0]["amount"]="2000000000000000000000"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["gov"]["params"]["voting_period"]="432000s"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["gov"]["params"]["max_deposit_period"]="259200s"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# Slashing params
jq '.app_state["slashing"]["params"]["signed_blocks_window"]="90000"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["slashing"]["params"]["downtime_jail_duration"]="1800s"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["slashing"]["params"]["slash_fraction_double_sign"]="0.1"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["slashing"]["params"]["slash_fraction_double_sign"]="0.005"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# Inflation params
jq '.app_state["mint"]["minter"]["inflation"]="0.000010000000000000"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["mint"]["params"]["inflation_rate_change"]="0.000005000000000000"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["mint"]["params"]["inflation_max"]="0.000020000000000000"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["mint"]["params"]["inflation_min"]="0.000005000000000000"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# Arachnid Proxy Deployment
jq --arg BYTECODE $ARACHNID_BYTECODE '.app_state.evm.accounts += [{"address":"0x4e59b44847b379578588920cA78FbF26c0B4956C", "code": $BYTECODE, "storage": []}]' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq --arg CODE_HASH $ARACHNID_CODEHASH '.app_state.auth.accounts += [{"@type": "/ethermint.types.v1.EthAccount", "base_account": {"account_number": "0", "address": "swtr1fevmgjz8kdu40pvgjgx20ralymqtf9tvcggehm", "pub_key": null, "sequence": "1" }, "code_hash": $CODE_HASH}]' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# Add regular accounts
./$(dirname "$0")/add_regular_accounts.sh

# Add vesting accounts
CURRENT_TIMESTAMP=$(date +%s)
./$(dirname "$0")/add_vesting_accounts.sh $CURRENT_TIMESTAMP

# Add issuer and operators
jq '.app_state.compliance.issuerDetails += [{"address": "swtr1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqpe55507", "details": {"creator": "swtr1ml2knanpk8sv94f8h9g8vaf9k3yyfva4fykyn9", "description": "World ID is privacy preserving proof of personhood, which allow for Proof of Humanity verifications", "legalEntity": "Worldcoin Foundation, World Assets Ltd.", "logo": "https://ipfs.io/ipfs/bafkreibi3idudk5wyvnjr7qrfyrpshg3bikpfte4o33wpmbia6o5tovpxe", "name": "Worldcoin Adapter", "url": "https://worldcoin.org"}}]' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state.compliance.addressDetails += [{"address": "swtr1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqpe55507", "details": {"is_revoked": false, "is_verified": true, "verifications": []}}]' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
# TODO: Add operators
# TODO: Set operator as issuer creator
# TODO: Add genesis accounts for operators

ADDRESS_DETAILS_JSON="$(dirname $0)"/misc/address_details.json
jq --slurpfile input "$ADDRESS_DETAILS_JSON" \
   '.app_state.compliance.addressDetails += $input[0]' \
   "$GENESIS" > "$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

VERIFICATION_DETAILS_JSON="$(dirname $0)"/misc/verification_details.json
jq --slurpfile input "$VERIFICATION_DETAILS_JSON" \
   '.app_state.compliance.verificationDetails += $input[0]' \
   "$GENESIS" > "$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"   