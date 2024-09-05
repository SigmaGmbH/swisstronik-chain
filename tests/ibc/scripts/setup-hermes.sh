#!/bin/sh
set -e

DEFAULT_CHAIN_A_ID="osmosis"
DEFAULT_CHAIN_A_MNEMONIC="black frequent sponsor nice claim rally hunt suit parent size stumble expire forest avocado mistake agree trend witness lounge shiver image smoke stool chicken"
DEFAULT_CHAIN_A_RPC="http://localhost:36657"
DEFAULT_CHAIN_B_ID="swisstronik_1291-1"
DEFAULT_CHAIN_B_MNEMONIC="black frequent sponsor nice claim rally hunt suit parent size stumble expire forest avocado mistake agree trend witness lounge shiver image smoke stool chicken"
DEFAULT_CHAIN_B_RPC="http://localhost:26657"

CHAIN_A_MNEMONIC=${CHAIN_A_MNEMONIC:-$DEFAULT_CHAIN_A_MNEMONIC}
CHAIN_A_ID=${CHAIN_A_ID:-$DEFAULT_CHAIN_A_ID}
CHAIN_A_RPC=${CHAIN_A_RPC:-$DEFAULT_CHAIN_A_RPC}
CHAIN_B_MNEMONIC=${CHAIN_B_MNEMONIC:-$DEFAULT_CHAIN_B_MNEMONIC}
CHAIN_B_ID=${CHAIN_B_ID:-$DEFAULT_CHAIN_B_ID}
CHAIN_B_RPC=${CHAIN_B_RPC:-$DEFAULT_CHAIN_B_RPC}

install_prerequisites(){
    echo "🧰 Install prerequisites"
    apt update
    apt -y install curl
}

add_keys(){

    echo "🔑 Adding key for $CHAIN_A_ID"
    mkdir -p /home/hermes/mnemonics/
    echo $CHAIN_A_MNEMONIC > /home/hermes/mnemonics/$CHAIN_A_ID

    hermes keys add \
    --chain $CHAIN_A_ID \
    --mnemonic-file /home/hermes/mnemonics/$CHAIN_A_ID \
    --key-name $CHAIN_A_ID \
    --overwrite

    echo "🔑 Adding key for $CHAIN_B_ID"
    echo $CHAIN_B_MNEMONIC > /home/hermes/mnemonics/$CHAIN_B_ID

    hermes keys add \
    --chain $CHAIN_B_ID \
    --mnemonic-file /home/hermes/mnemonics/$CHAIN_B_ID \
    --key-name $CHAIN_B_ID \
    --overwrite
}

create_channel(){
    echo "🥱 Waiting for $CHAIN_A_ID to start"
    COUNTER=0
    until $(curl --output /dev/null --silent --head --fail $CHAIN_A_RPC/status); do
        printf '.'
        sleep 2
    done

    echo "🥱 Waiting for $CHAIN_B_ID to start"
    COUNTER=0
    until $(curl --output /dev/null --silent --head --fail $CHAIN_B_RPC/status); do
        printf '.'
        sleep 5
    done

    echo "📺 Creating channel $CHAIN_A_ID <> $CHAIN_B_ID"
    hermes create channel \
    --a-chain $CHAIN_A_ID \
    --b-chain $CHAIN_B_ID \
    --a-port transfer \
    --b-port transfer \
    --new-client-connection --yes
}

install_prerequisites
add_keys
create_channel

echo "✉️ Start Hermes"
hermes start