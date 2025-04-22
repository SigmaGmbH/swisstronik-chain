#!/bin/bash

SCRIPT_DIR=$(dirname $0)

HOMEDIR="$SCRIPT_DIR/.swisstronik-val2"
KEYRING=test
BINARY="$SCRIPT_DIR/old/bin/swisstronikd"

set -e

$BINARY tx gov submit-proposal draft_proposal.json --from alice -y --gas-prices 7aswtr --home $HOMEDIR
sleep 5
$BINARY tx gov deposit 1 10000000000000000000000aswtr --from alice -y --gas-prices 7aswtr --home $HOMEDIR
sleep 5
$BINARY tx gov vote 1 yes --from alice -y --gas-prices 7aswtr --home $HOMEDIR
sleep 5
$BINARY tx gov vote 1 yes --from bob -y --gas-prices 7aswtr --home $HOMEDIR
sleep 15
$BINARY q gov proposals --home $HOMEDIR