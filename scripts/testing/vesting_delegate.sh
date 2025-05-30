#!/bin/bash

set -e

CHAIN_ID="swisstronik_1848-1"
KEYRING_BACKEND="test"
BINARY="$(dirname "$0")/../../build/swisstronikd"
DELEGATE_AMOUNT="100000000aswtr"
TREASURY_KEY_NAME="treasury"

# Get validator address
VALIDATOR_ADDRESS=$($BINARY query staking validators --output json | jq -r '.validators[0].operator_address')

if [ -z "$VALIDATOR_ADDRESS" ]; then
  echo "Error: Could not retrieve validator address."
  exit 1
fi

echo "Validator address: $VALIDATOR_ADDRESS"

# Get first vesting account address
VESTING_ACCOUNT=$($BINARY keys list --keyring-backend $KEYRING_BACKEND --output json | jq -r '.[0].address')

if [ -z "$VESTING_ACCOUNT" ]; then
  echo "Error: Could not retrieve vesting account address."
  exit 1
fi

echo "Vesting account address: $VESTING_ACCOUNT"

# Fund created vesting account
$BINARY tx bank send $TREASURY_KEY_NAME $VESTING_ACCOUNT 3500000aswtr -y --gas-prices 7aswtr

sleep 5

# Delegate tokens
$BINARY tx staking delegate $VALIDATOR_ADDRESS $DELEGATE_AMOUNT --from $VESTING_ACCOUNT --chain-id $CHAIN_ID --yes --keyring-backend $KEYRING_BACKEND --gas-prices 7aswtr --gas 500000