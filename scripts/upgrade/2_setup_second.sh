#!/bin/bash

SCRIPT_DIR=$(dirname $0)

CHAINID="swisstronik_1848-1"
MONIKER="localtestnet2"
KEYRING="test"
KEYALGO="eth_secp256k1"
BINARY="./old/bin/swisstronikd"
HOMEDIR="$SCRIPT_DIR/.swisstronik-val2"
OTHER_HOMEDIR="$SCRIPT_DIR/.swisstronik"

# Path variables
CONFIG=$HOMEDIR/config/config.toml
APP_TOML=$HOMEDIR/config/app.toml
GENESIS=$HOMEDIR/config/genesis.json
TMP_GENESIS=$HOMEDIR/config/tmp_genesis.json
ENCLAVE_HOME=$SCRIPT_DIR/.swisstronik-enclave-val2
KEYMANAGER_HOME=$SCRIPT_DIR/.swisstronik-enclave-val2

# validate dependencies are installed
command -v jq >/dev/null 2>&1 || {
	echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"
	exit 1
}

# used to exit on first error (any non-zero exit code)
set -e

rm -rf "$HOMEDIR"
rm -rf "$ENCLAVE_HOME"

cp -r $SCRIPT_DIR/.swisstronik-enclave $SCRIPT_DIR/.swisstronik-enclave-val2

$BINARY config keyring-backend $KEYRING --home "$HOMEDIR"
$BINARY config chain-id $CHAINID --home "$HOMEDIR"

echo "betray theory cargo way left cricket doll room donkey wire reunion fall left surprise hamster corn village happy bulb token artist twelve whisper expire" | $BINARY keys add alice --keyring-backend $KEYRING --home $HOMEDIR --recover
echo "toss sense candy point cost rookie jealous snow ankle electric sauce forward oblige tourist stairs horror grunt tenant afford master violin final genre reason" | $BINARY keys add bob --keyring-backend $KEYRING --home $HOMEDIR --recover
echo "offer feel open ancient relax habit field right evoke ball organ beauty" | $BINARY keys add test1 --recover  --keyring-backend $KEYRING --home "$HOMEDIR"
echo "olympic such citizen any bind small neutral hidden prefer pupil trash lemon" | $BINARY keys add test2 --recover  --keyring-backend $KEYRING --home "$HOMEDIR"
echo "cup hip eyebrow flock slogan filter gas tent angle purpose rose setup" | $BINARY keys add operator --recover --keyring-backend $KEYRING --home "$HOMEDIR"

$BINARY init $MONIKER -o --chain-id $CHAINID --home "$HOMEDIR"

# expose ports
sed -i 's/127.0.0.1:26657/0.0.0.0:36657/g' "$CONFIG"
sed -i 's/0.0.0.0:26656/0.0.0.0:36656/g' "$CONFIG"
sed -i 's/127.0.0.1:8545/0.0.0.0:9545/g' "$APP_TOML"
sed -i 's/127.0.0.1:8546/0.0.0.0:9546/g' "$APP_TOML"
sed -i 's/127.0.0.1:8547/0.0.0.0:9547/g' "$APP_TOML"
sed -i 's/127.0.0.1:8548/0.0.0.0:9548/g' "$APP_TOML"
sed -i 's/localhost:9090/localhost:10090/g' "$APP_TOML"
sed -i 's/localhost:9091/localhost:10091/g' "$APP_TOML"


# enable prometheus metrics
sed -i 's/prometheus-retention-time  = "0"/prometheus-retention-time  = "1000000000000"/g' "$APP_TOML"
sed -i 's/enabled = false/enabled = true/g' "$APP_TOML"

# set min gas price
sed -i 's/minimum-gas-prices = ""/minimum-gas-prices = "0aswtr"/' "$APP_TOML"

# set custom pruning settings
sed -i.bak 's/pruning = "default"/pruning = "custom"/g' "$APP_TOML"
sed -i.bak 's/pruning-keep-recent = "0"/pruning-keep-recent = "2"/g' "$APP_TOML"
sed -i.bak 's/pruning-interval = "0"/pruning-interval = "10"/g' "$APP_TOML"

# request genesis.json
curl http://localhost:26657/genesis? | jq ".result.genesis" > $HOMEDIR/config/genesis.json

# add validator 1 to persistent peers
FIRST_NODE_ID=$($BINARY tendermint show-node-id --home $OTHER_HOMEDIR)
sed -i 's/persistent_peers = ""/persistent_peers = "'$FIRST_NODE_ID'@127.0.0.1:26656"/' $CONFIG

# configure cosmovisor
DAEMON_NAME=swisstronikd
DAEMON_HOME=$HOMEDIR
DAEMON_ALLOW_DOWNLOAD_BINARIES=false
DAEMON_RESTART_AFTER_UPGRADE=true

DAEMON_HOME=$DAEMON_HOME DAEMON_NAME=$DAEMON_NAME cosmovisor init $BINARY

# add binary for upgrade
NEW_BINARY="$SCRIPT_DIR/../../build/swisstronikd"
mkdir -p $HOMEDIR/cosmovisor/upgrades/v1.0.1/bin
cp $NEW_BINARY $HOMEDIR/cosmovisor/upgrades/v1.0.1/bin

# Start chain
DAEMON_HOME=$DAEMON_HOME DAEMON_NAME=$DAEMON_NAME ENCLAVE_HOME=$ENCLAVE_HOME KEYMANAGER_HOME=$KEYMANAGER_HOME cosmovisor run start --home $HOMEDIR