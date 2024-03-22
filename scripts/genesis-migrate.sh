#!/bin/sh
set -euo pipefail

echo "Migrating genesis file"

GENESIS_FILE="$HOME/.swisstronik/config/genesis.json"
TMP_GENESIS_FILE="./genesis.json"
UPDATES_FILE="./gov_params_change.json"

gov_params=$(jq '.params' $UPDATES_FILE)

# Update the fields in the JSON object
updated_json=$(jq ".app_state.gov.params = ${gov_params}" $GENESIS_FILE)

# Write the updated JSON object to the temp file
echo "$updated_json" > $TMP_GENESIS_FILE
mv $TMP_GENESIS_FILE $GENESIS_FILE