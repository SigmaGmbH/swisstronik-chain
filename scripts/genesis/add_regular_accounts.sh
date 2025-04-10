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

DENOM="swtr"

if [[ -z "$SWTR_BINARY" ]]; then
  BINARY="./build/swisstronikd"
else
  BINARY="${SWTR_BINARY}"
fi

# CSV File format looks like as following:
# -----------
# address,balance
# swtr1qa2h6a27waactkrc6dyxrn2jzfjjfg24dgxzu8,3000
# swtr1qg9e0d8y9w0z5h7v4x5lq2k3m8n0p6s9y3k5t2,3000
# swtr1q8n0p6s9y3k5t2a1b2c3d4e5f6g7h8i9j0k1l2,6000
# swtr1q7k5t2a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6,120000
# -----------

header_added="\033[32m[ADD]\033[0m"
header_removed="\033[31m[DEL]\033[0m"

# Read the CSV file and process each line
while IFS=, read -r address balance; do
    # Skip the header line
    if [ "$address" == "address" ]; then
        continue
    fi

    $BINARY add-genesis-account "$address" "${balance}${DENOM}" --home "$HOMEDIR" --keyring-backend "$KEYRING"

done < "$CSV_FILE"