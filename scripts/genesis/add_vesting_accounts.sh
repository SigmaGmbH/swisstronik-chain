#!/bin/bash

HOMEDIR="$HOME/.swisstronik"

CSV_FILE="$(dirname "$0")/misc/vesting.csv"
GENESIS_FILE=$HOMEDIR/config/genesis.json
TEMP_GENESIS=$HOMEDIR/config/tmp_genesis.json

# validate dependencies are installed
command -v jq >/dev/null 2>&1 || {
	echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"
	exit 1
}

command -v bc &> /dev/null 2>&1 || {
	echo "bc cnot installed. Please install it first."
	exit 1
}

# used to exit on first error (any non-zero exit code)
set -e

# Get start time from command line argument
if [ -z "$1" ]; then
    echo "Please provide the start time(timestamp) as a parameter (e.g., ./genesis-vesting.sh 1720744246)."
    exit 1
fi

START_TIME=$1

# CSV File format looks like as following:
# -----------
# address,original_vesting,cliff_days,months
# swtr1k994w6syjtvcdhdyl9hvmatp6cedwsns7gyw4m,30000000000000000000000aswtr,10,3
# swtr1zpjdwmpumcwajwz6w3ujh0laxq3gage297hff3,30000000000000000000000aswtr,10,3
# swtr140s4tgyc7d9ua47f3sk2dkgqcks235uzhe43lw,60000000000000000000000aswtr,30,6
# swtr1u0l99lamh7rjcpldm4p94xxssxmxtfty9mgvuu,120000000000000000000000aswtr,30,10
# -----------

header_added="\033[32m[ADD]\033[0m"
header_removed="\033[31m[DEL]\033[0m"

# Read the CSV file and process each line
while IFS=, read -r address original_vesting cliff_days months; do
    # Skip the header line
    if [ "$address" == "address" ]; then
        continue
    fi

    # Extract amount and denom from original_vesting
    original_amount=$(echo "$original_vesting" | sed -E 's/^([0-9]+).*/\1/')
    denom=$(echo "$original_vesting" | sed -E 's/^[0-9]+([a-zA-Z]+)$/\1/')

    # Calculate cliff time
    cliff_time=$(($START_TIME + $cliff_days * 60 * 60 * 24))

    # Calculate end time (start time + months * 30 days in seconds)
    end_time=$(($cliff_time + $months * 30 * 24 * 60 * 60))

	# Calculate vesting amount per period using bc for large numbers
	vesting_amount=$(echo "$original_amount / $months" | bc)

    # Create vesting periods (1 month = 30 days in seconds)
    vesting_periods=()
    for (( i=0; i<$months; i++ )); do
        vesting_periods+=("{\"amount\":[{\"amount\":\"$vesting_amount\",\"denom\":\"$denom\"}],\"length\":\"2592000\"}")
    done
    vesting_periods_json=$(IFS=,; echo "[${vesting_periods[*]}]")

    # Create MonthlyVestingAccount entry
    vesting_entry=$(jq -n --arg addr "$address" --arg amount "$original_amount" --arg denom "$denom" --arg start "$START_TIME" --arg cliff "$cliff_time" --arg r_end "$end_time" --argjson periods "$vesting_periods_json" '{
        "@type": "/swisstronik.vesting.MonthlyVestingAccount",
        "base_vesting_account": {
            "base_account": {
                "address": $addr,
                "pub_key": null,
                "sequence": "0"
            },
            "delegated_free": [],
            "delegated_vesting": [],
            "end_time": $r_end,
            "original_vesting": [{
                "amount": $amount,
                "denom": $denom
            }]
        },
        "cliff_time": $cliff,
        "start_time": $start,
        "vesting_periods": $periods
    }')

	# Check if the address already exists in genesis.json
    address_exists=$(jq --arg addr "$address" '.app_state.auth.accounts[] | select(.base_vesting_account.base_account.address == $addr)' "$GENESIS_FILE")

	if [ -n "$address_exists" ]; then
        # Remove existing entry
        jq --arg addr "$address" 'del(.app_state.auth.accounts[] | select(.base_vesting_account.base_account.address == $addr))' "$GENESIS_FILE" > "$TEMP_GENESIS"
        mv "$TEMP_GENESIS" "$GENESIS_FILE"
		echo -e "$header_removed Address $address was removed from genesis.json."
    fi

    # Add the vesting entry to the genesis file
    jq --argjson vesting "$vesting_entry" '.app_state.auth.accounts += [$vesting]' "$GENESIS_FILE" > "$TEMP_GENESIS"
    mv "$TEMP_GENESIS" "$GENESIS_FILE"
    echo -e "$header_added Added vesting account for address $address"

done < "$CSV_FILE"

echo -e "\nVesting accounts added to $GENESIS_FILE"