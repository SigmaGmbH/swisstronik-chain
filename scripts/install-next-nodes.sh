#!/bin/bash
# Require 2 parameters
# $1 = moniker and wallet name
# $2 = RPC URL (Example http://localhost:26657)

HOMEDIR="$HOME/.swisstronik"
CONFIG=$HOMEDIR/config/config.toml
APP_TOML=$HOMEDIR/config/app.toml
GENESIS=$HOMEDIR/config/genesis.json
TMP_GENESIS=$HOMEDIR/config/tmp_genesis.json
MONIKER=$1
RPC=$2
if [[ -z $MONIKER ]]; then
	echo >&2 "invalid MONIKER"
    exit 1
fi
if [[ -z $RPC ]]; then
	echo >&2 "invalid RPC"
    exit 1
fi

# validate dependencies are installed
command -v jq >/dev/null 2>&1 || {
	echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"
	exit 1
}

sudo rm -rf $HOMEDIR
cd $HOME/chain/ && git pull
cd $HOME/chain/ && SGX_MODE=SW make build-enclave
cd $HOME/chain/ && make install
swisstronikd init $MONIKER --chain-id swisstronik_1291-1
swisstronikd keys add $MONIKER --keyring-backend test
curl $RPC/genesis? | jq ".result.genesis" > $HOMEDIR/config/genesis.json
swisstronikd config node $RPC
sed -i 's/pruning = "default"/pruning = "custom"/g' "$CONFIG"
sed -i 's/pruning-keep-recent = "0"/pruning-keep-recent = "2"/g' "$APP_TOML"
sed -i 's/pruning-interval = "0"/pruning-interval = "10"/g' "$APP_TOML"
sed -i 's/127.0.0.1:26657/0.0.0.0:26657/g' "$CONFIG"
sed -i 's/cors_allowed_origins\s*=\s*\[\]/cors_allowed_origins = ["*",]/g' "$CONFIG"