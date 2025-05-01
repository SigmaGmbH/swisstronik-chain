#!/bin/bash

SCRIPT_DIR=$(dirname $0)

MONIKER="localtestnet2"
BINARY="$SCRIPT_DIR/old/bin/swisstronikd"
HOMEDIR="$SCRIPT_DIR/.swisstronik-val2"

$BINARY tx staking create-validator \
  --amount 1000000000000000000000aswtr \
  --commission-max-change-rate "0.05" \
  --commission-max-rate "0.10" \
  --commission-rate "0.05" \
  --min-self-delegation "1" \
  --pubkey $($BINARY tendermint show-validator --home $HOMEDIR) \
  --moniker $MONIKER \
  --from bob \
  --website "http://test.com" \
  --identity "0A6AF02D1557E5B4" \
  --gas-prices 7aswtr --gas 250000 -y --home $HOMEDIR

sleep 5

$BINARY q staking validators  --home $HOMEDIR
  