#!/bin/bash

set -e

CHAINID="swisstronik_1848-1"
MONIKER="localtestnet"
KEYRING="test"
KEYALGO="eth_secp256k1"
HOMEDIR="$HOME/.swisstronik"
BINARY="./build/swisstronikd"

CSV_FILE="$(dirname "$0")/misc/balances.csv"
GENESIS=$HOMEDIR/config/genesis.json
TMP_GENESIS=$HOMEDIR/config/tmp_genesis.json

if [[ -z "$SWTR_BINARY" ]]; then
  BINARY="./build/swisstronikd"
else
  BINARY="${SWTR_BINARY}"
fi

# CSV File format looks like as following:
# -----------
# address,balance
# swtr1qa2h6a27waactkrc6dyxrn2jzfjjfg24dgxzu8,30000000000000000000000aswtr
# swtr1qg9e0d8y9w0z5h7v4x5lq2k3m8n0p6s9y3k5t2,30000000000000000000000aswtr
# swtr1q8n0p6s9y3k5t2a1b2c3d4e5f6g7h8i9j0k1l2,60000000000000000000000aswtr
# swtr1q7k5t2a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6,120000000000000000000000aswtr
# -----------

header_added="\033[32m[ADD]\033[0m"
header_removed="\033[31m[DEL]\033[0m"

# Read the CSV file and process each line
while IFS=, read -r address balance; do
    # Skip the header line
    if [ "$address" == "address" ]; then
        continue
    fi

    # Get the current max account number from genesis
    START_ACCOUNT_NUM=$(jq '[.app_state.auth.accounts[]?.account_number | select(type == "string") | select(test("^[0-9]+$")) | tonumber] | max // 0' "$GENESIS")
    echo "START ACCOUNT NUM: $START_ACCOUNT_NUM"

    if [[ "$START_ACCOUNT_NUM" == "null" ]]; then
      START_ACCOUNT_NUM=0
    fi

    ACCOUNT_NUM=$((START_ACCOUNT_NUM + 1))

    # Extract amount and denom from balance
    amount=$(echo "$balance" | sed -E 's/^([0-9]+).*/\1/')
    denom=$(echo "$balance" | sed -E 's/^[0-9]+([a-zA-Z]+)$/\1/')

    jq --arg addr "$address" --arg amt "$amount" --arg denom "$denom" \
      '.app_state.bank.balances += [{
        "address": $addr,
        "coins": [{"denom": $denom, "amount": $amt}]
    }]' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"


    jq --arg addr "$address" --argjson acctnum "$ACCOUNT_NUM" \
      '.app_state.auth.accounts += [{
        "@type": "/cosmos.auth.v1beta1.BaseAccount",
        "address": $addr,
        "pub_key": null,
        "account_number": ($acctnum | tostring),
        "sequence": "0"
    }]' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

    echo -e "$header_added Added regular account for address $address"

done < "$CSV_FILE"