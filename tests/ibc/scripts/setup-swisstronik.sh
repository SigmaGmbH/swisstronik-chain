#!/bin/bash
set -eu

DEFAULT_CHAIN_ID="swisstronik_1291-1"
DEFAULT_VALIDATOR_MONIKER="validator"
DEFAULT_VALIDATOR_MNEMONIC="bottom loan skill merry east cradle onion journey palm apology verb edit desert impose absurd oil bubble sweet glove shallow size build burst effort"
DEFAULT_FAUCET_MNEMONIC="increase bread alpha rigid glide amused approve oblige print asset idea enact lawn proof unfold jeans rabbit audit return chuckle valve rather cactus great"
DEFAULT_RELAYER_MNEMONIC="black frequent sponsor nice claim rally hunt suit parent size stumble expire forest avocado mistake agree trend witness lounge shiver image smoke stool chicken"

# Override default values with environment variables
CHAIN_ID=${CHAIN_ID:-$DEFAULT_CHAIN_ID}
VALIDATOR_MONIKER=${VALIDATOR_MONIKER:-$DEFAULT_VALIDATOR_MONIKER}
VALIDATOR_MNEMONIC=${VALIDATOR_MNEMONIC:-$DEFAULT_VALIDATOR_MNEMONIC}
FAUCET_MNEMONIC=${FAUCET_MNEMONIC:-$DEFAULT_FAUCET_MNEMONIC}
RELAYER_MNEMONIC=${RELAYER_MNEMONIC:-$DEFAULT_RELAYER_MNEMONIC}

SWISSTRONIK_HOME=$HOME/.swisstronik/
CONFIG_FOLDER=$SWISSTRONIK_HOME/config

#install_prerequisites () {
#  apk add dasel
#}

edit_genesis () {
  GENESIS=$CONFIG_FOLDER/genesis.json
  TMP_GENESIS=$CONFIG_FOLDER/tmp_genesis.jsonn

  jq '.app_state["feemarket"]["params"]["base_fee"]="7"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
  jq '.app_state["staking"]["params"]["bond_denom"]="aswtr"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
  jq '.app_state["staking"]["params"]["unbonding_time"]="1s"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
  jq '.app_state["crisis"]["constant_fee"]["denom"]="aswtr"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
  jq '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="aswtr"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
  jq '.app_state["gov"]["params"]["min_deposit"][0]["denom"]="aswtr"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
  jq '.app_state["evm"]["params"]["evm_denom"]="aswtr"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
  jq '.app_state["inflation"]["params"]["mint_denom"]="aswtr"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
  jq '.app_state["mint"]["params"]["mint_denom"]="aswtr"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
  jq '.consensus_params["block"]["max_gas"]="10000000"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
  jq '.app_state["compliance"]["operators"]=[{"operator":"swtr1ml2knanpk8sv94f8h9g8vaf9k3yyfva4fykyn9", "operator_type": 1}]' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

  # Change proposal periods to pass within a reasonable time for local testing
  sed -i.bak 's/"max_deposit_period": "172800s"/"max_deposit_period": "30s"/g' "$GENESIS"
  sed -i.bak 's/"voting_period": "172800s"/"voting_period": "30s"/g' "$GENESIS"
}

add_genesis_accounts () {
  # Validator
  echo "⚖️ Add validator account"
  echo $VALIDATOR_MNEMONIC | swisstronikd keys add $VALIDATOR_MONIKER --recover --keyring-backend=test --home $SWISSTRONIK_HOME
  VALIDATOR_ACCOUNT=$(swisstronikd keys show -a $VALIDATOR_MONIKER --keyring-backend test --home $SWISSTRONIK_HOME)
  swisstronikd add-genesis-account $VALIDATOR_ACCOUNT 100000000swtr --home $SWISSTRONIK_HOME

  # Faucet
  echo "🚰 Add faucet account"
  echo $FAUCET_MNEMONIC | swisstronikd keys add faucet --recover --keyring-backend=test --home $SWISSTRONIK_HOME
  FAUCET_ACCOUNT=$(swisstronikd keys show -a faucet --keyring-backend test --home $SWISSTRONIK_HOME)
  swisstronikd add-genesis-account $FAUCET_ACCOUNT 100000000swtr --home $SWISSTRONIK_HOME

  # Relayer
  echo "🔗 Add relayer account"
  echo $RELAYER_MNEMONIC | swisstronikd keys add relayer --recover --keyring-backend=test --home $SWISSTRONIK_HOME
  RELAYER_ACCOUNT=$(swisstronikd keys show -a relayer --keyring-backend test --home $SWISSTRONIK_HOME)
  swisstronikd add-genesis-account $RELAYER_ACCOUNT 100000000swtr --home $SWISSTRONIK_HOME

  swisstronikd gentx $VALIDATOR_MONIKER 1000000000000000000000aswtr --keyring-backend=test --chain-id=$CHAIN_ID --home $SWISSTRONIK_HOME
  swisstronikd collect-gentxs --home $SWISSTRONIK_HOME

  swisstronikd validate-genesis --home $SWISSTRONIK_HOME
}

edit_config () {
  CONFIG=$CONFIG_FOLDER/config.toml
  APP_TOML=$CONFIG_FOLDER/app.toml

  # expose ports
  sed -i 's/127.0.0.1:26657/0.0.0.0:26657/g' "$CONFIG"
  sed -i 's/127.0.0.1:8545/0.0.0.0:8545/g' "$APP_TOML"
  sed -i 's/127.0.0.1:8546/0.0.0.0:8546/g' "$APP_TOML"

  # enable prometheus metrics
  sed -i 's/prometheus = false/prometheus = true/' "$CONFIG"
  sed -i 's/prometheus-retention-time  = "0"/prometheus-retention-time  = "1000000000000"/g' "$APP_TOML"
  sed -i 's/enabled = false/enabled = true/g' "$APP_TOML"

  # set min gas price
  sed -i 's/minimum-gas-prices = ""/minimum-gas-prices = "0aswtr"/' "$APP_TOML"

  # set custom pruning settings
  sed -i.bak 's/pruning = "default"/pruning = "custom"/g' "$APP_TOML"
  sed -i.bak 's/pruning-keep-recent = "0"/pruning-keep-recent = "2"/g' "$APP_TOML"
  sed -i.bak 's/pruning-interval = "0"/pruning-interval = "10"/g' "$APP_TOML"
}

init_epoch_keys () {
  # Initialize epoch keys for local testnet
  swisstronikd testnet init-testnet-enclave
}

if [[ ! -d $CONFIG_FOLDER ]]
then
#  install_prerequisites
  echo "🧪 Creating Swisstronik home for $VALIDATOR_MONIKER"
  echo $VALIDATOR_MNEMONIC | swisstronikd init -o --chain-id=$CHAIN_ID --home $SWISSTRONIK_HOME --recover $VALIDATOR_MONIKER
  edit_genesis
  add_genesis_accounts
  edit_config
  init_epoch_keys
fi

echo "🏁 Starting $CHAIN_ID..."
swisstronikd start --home $SWISSTRONIK_HOME
