#!/bin/bash

SCRIPT_DIR=$(dirname $0)

CHAINID="swisstronik_1848-1"
MONIKER="localtestnet"
KEYRING="test"
KEYALGO="eth_secp256k1"
BINARY="$SCRIPT_DIR/../../build/swisstronikd"
HOMEDIR="$SCRIPT_DIR/.swisstronik"

ENCLAVE_HOME=$SCRIPT_DIR/.swisstronik-enclave
KEYMANAGER_HOME=$SCRIPT_DIR/.swisstronik-enclave

# Start chain
KEYMANAGER_HOME=$KEYMANAGER_HOME ENCLAVE_HOME=$ENCLAVE_HOME $BINARY start --home $HOMEDIR --json-rpc.ws-address-unencrypted 127.0.0.1:8548 --json-rpc.address-unencrypted 127.0.0.1:8547