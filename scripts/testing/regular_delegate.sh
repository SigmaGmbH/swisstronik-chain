#!/bin/bash

set -e

CHAIN_ID="swisstronik_1848-1"
KEYRING_BACKEND="test"
BINARY="$(dirname "$0")/../../build/swisstronikd"
DELEGATE_AMOUNT="1swtr"
TREASURY_KEY_NAME="treasury"

# Get validator address
VALIDATOR_ADDRESS=$($BINARY query staking validators --output json | jq -r '.validators[0].operator_address')

if [ -z "$VALIDATOR_ADDRESS" ]; then
  echo "Error: Could not retrieve validator address."
  exit 1
fi

DELEGATOR_ADDRESS=$($BINARY keys show -a $TREASURY_KEY_NAME --keyring-backend $KEYRING_BACKEND)
if [ -z "$DELEGATOR_ADDRESS" ]; then
  echo "Error: Could not retrieve delegator address."
  exit 1
fi

echo "Delegator address: $DELEGATOR_ADDRESS"
echo "Validator address: $VALIDATOR_ADDRESS"

# Delegate tokens
$BINARY tx staking delegate $VALIDATOR_ADDRESS $DELEGATE_AMOUNT --from $TREASURY_KEY_NAME --chain-id $CHAIN_ID --yes --keyring-backend $KEYRING_BACKEND --gas-prices 7aswtr --gas 500000
