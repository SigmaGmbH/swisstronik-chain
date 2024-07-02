#!/bin/bash

HOMEDIR=~/.swisstronik
KEYRING=test
BINARY=./build/swisstronikd

set -e

$BINARY tx gov submit-proposal draft_proposal.json --from alice -y --gas-prices 7aswtr
sleep 3
$BINARY tx gov deposit 1 10000000000aswtr --from alice -y --gas-prices 7aswtr
sleep 3
$BINARY tx gov vote 1 yes --from alice -y --gas-prices 7aswtr
$BINARY tx gov vote 1 yes --from bob -y --gas-prices 7aswtr