#!/bin/bash

set -e

CHAIN_ID="swisstronik_1291-1"
KEYRING_BACKEND="test"
BINARY="$(dirname "$0")/../../build/swisstronikd"
DELEGATE_AMOUNT="100000000aswtr"

# Get validator address
VALIDATOR_ADDRESS=$($BINARY query staking validators --output json | jq -r '.validators[0].operator_address')

if [ -z "$VALIDATOR_ADDRESS" ]; then
  echo "Error: Could not retrieve validator address."
  exit 1
fi

echo "Validator address: $VALIDATOR_ADDRESS"

# Create new vesting account using CLI
VESTING_ACCOUNT=vesting
TREASURY_KEY_NAME="treasury"

$BINARY keys add $VESTING_ACCOUNT --keyring-backend $KEYRING_BACKEND

VESTING_ACC_ADDRESS=$($BINARY keys show $VESTING_ACCOUNT -a --keyring-backend $KEYRING_BACKEND)

echo "Vesting account address: $VESTING_ACC_ADDRESS"

$BINARY tx vesting create-monthly-vesting-account $VESTING_ACC_ADDRESS 0 10 100000000000000aswtr -y --from $TREASURY_KEY_NAME --gas-prices 7aswtr

sleep 3

# Fund created vesting account
$BINARY tx bank send $TREASURY_KEY_NAME $VESTING_ACC_ADDRESS 3500000aswtr -y --gas-prices 7aswtr

sleep 3

# Delegate tokens
$BINARY tx staking delegate $VALIDATOR_ADDRESS $DELEGATE_AMOUNT --from $VESTING_ACCOUNT --chain-id $CHAIN_ID --yes --keyring-backend $KEYRING_BACKEND --gas-prices 7aswtr --gas 500000