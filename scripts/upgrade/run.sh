#!/bin/bash

REPO_URL="https://github.com/SigmaGmbH/swisstronik-chain.git"
OLD_TAG="testnet-v1.0.7"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OLD_SOURCES="$SCRIPT_DIR/old"
NEW_SOURCES="$SCRIPT_DIR/../.."

SGX_MODE=SW
PID_FILE="$SCRIPT_DIR/pid"

OLD_BINARY="$SCRIPT_DIR/old/build/swisstronikd"
NEW_BINARY="$SCRIPT_DIR/../../build/swisstronikd"

KEYMANAGER_HOME="$SCRIPT_DIR/.swisstronik-enclave"
ENCLAVE_HOME="$SCRIPT_DIR/.swisstronik-enclave"

HOMEDIR="$SCRIPT_DIR/.swisstronik"
DAEMON_NAME=swisstronikd
DAEMON_HOME=$HOMEDIR
DAEMON_ALLOW_DOWNLOAD_BINARIES=false
DAEMON_RESTART_AFTER_UPGRADE=true

if [ -z "$1" ]; then
  echo "Usage: $0 {build|run|upgrade}"
  exit 1
fi

case "$1" in
  build)
    echo "Building binaries..."

    if [ -d "$OLD_SOURCES" ]; then
      echo "Removing existing '$OLD_SOURCES' folder..."
      rm -rf "$OLD_SOURCES"
    fi

    echo "Cloning $OLD_TAG into '$OLD_SOURCES'..."
    git clone --branch "$OLD_TAG" --depth 1 "$REPO_URL" "$OLD_SOURCES" || { echo "Failed to clone repository."; exit 1; }

    echo "Running 'make build' in '$OLD_SOURCES'..."
    (cd "$OLD_SOURCES" && git submodule update --init --recursive && ENCLAVE_HOME=$ENCLAVE_HOME SGX_MODE=$SGX_MODE make build_d) || { echo "Build failed."; exit 1; }

    echo "Running 'make build' in '$NEW_SOURCES'..."
    (cd "$NEW_SOURCES" && ENCLAVE_HOME=$ENCLAVE_HOME SGX_MODE=$SGX_MODE make build_d) || { echo "Build failed."; exit 1; }

    ;;
  run)
    echo "Running chain..."

    sh init.sh "$SCRIPT_DIR" || { echo "Init script failed."; exit 1; }
    nohup env DAEMON_HOME=$DAEMON_HOME DAEMON_NAME=$DAEMON_NAME ENCLAVE_HOME=$ENCLAVE_HOME KEYMANAGER_HOME=$KEYMANAGER_HOME cosmovisor run start --home $HOMEDIR > swtr.log 2>&1 &
    echo $! > "$PID_FILE"
    echo "Chain started with PID: $(cat "$PID_FILE")"
    ;;
  mock)
    echo "Running chain with mocked genesis..."

    sh init.sh "$SCRIPT_DIR" || { echo "Init script failed."; exit 1; }

    # Replace node files
    cp "$SCRIPT_DIR/misc/genesis.json" "$HOMEDIR/config/genesis.json"
    cp "$SCRIPT_DIR/misc/node_key.json" "$HOMEDIR/config/node_key.json"
    cp "$SCRIPT_DIR/misc/priv_validator_key.json" "$HOMEDIR/config/priv_validator_key.json"
    cp "$SCRIPT_DIR/misc/priv_validator_state.json" "$HOMEDIR/data/priv_validator_state.json"

    nohup env DAEMON_HOME=$DAEMON_HOME DAEMON_NAME=$DAEMON_NAME ENCLAVE_HOME=$ENCLAVE_HOME KEYMANAGER_HOME=$KEYMANAGER_HOME cosmovisor run start --home $HOMEDIR > swtr.log 2>&1 &
    echo $! > "$PID_FILE"
    echo "Chain started with PID: $(cat "$PID_FILE")"
    ;;
  upgrade)
    echo "Proposing upgrade..."
    $OLD_BINARY tx gov submit-proposal "$SCRIPT_DIR/proposal.json" --from alice -y --gas-prices 7aswtr --home $HOMEDIR
    sleep 5
    $OLD_BINARY tx gov deposit 1 10000000000000000000000aswtr --from alice -y --gas-prices 7aswtr --home $HOMEDIR
    sleep 5
    $OLD_BINARY tx gov vote 1 yes --from alice -y --gas-prices 7aswtr --home $HOMEDIR
    sleep 5
    $OLD_BINARY tx gov vote 1 yes --from bob -y --gas-prices 7aswtr --home $HOMEDIR
    sleep 15
    $OLD_BINARY q gov proposals --home $HOMEDIR
    ;;
  stop)
    echo "Stopping the application..."
    if [ ! -f "$PID_FILE" ]; then
      echo "No PID file found. Is the application running?"
      exit 1
    fi

    PID=$(cat "$PID_FILE")
    if kill "$PID" > /dev/null 2>&1; then
      echo "Chain (PID: $PID) stopped."
      rm -f "$PID_FILE"  # Remove the PID file
    else
      echo "Failed to stop chain (PID: $PID)."
      exit 1
    fi
    ;;
  *)
    echo "Invalid command: $1"
    echo "Usage: $0 {build|run|mock|upgrade}"
    exit 1
    ;;
esac

exit 0