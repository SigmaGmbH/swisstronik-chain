HOMEDIR="$HOME/.swisstronik"
CONFIG=$HOMEDIR/config/config.toml
APP_TOML=$HOMEDIR/config/app.toml
GENESIS=$HOMEDIR/config/genesis.json
TMP_GENESIS=$HOMEDIR/config/tmp_genesis.json

# validate dependencies are installed
command -v jq >/dev/null 2>&1 || {
	echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"
	exit 1
}
sudo rm -rf $HOMEDIR
cd $HOME/chain/ && git pull
cd $HOME/chain/ && SGX_MODE=SW make build-enclave
cd $HOME/chain/ && make install
mkdir -p $HOMEDIR/cosmovisor/genesis/bin && mkdir -p $HOMEDIR/cosmovisor/upgrades
cp $HOME/go/bin/swisstronikd $HOMEDIR/cosmovisor/genesis/bin
swisstronikd init validator --chain-id swisstronik_1291-1
echo "pet apart myth reflect stuff force attract taste caught fit exact ice slide sheriff state since unusual gaze practice course mesh magnet ozone purchase" | swisstronikd keys add validator --keyring-backend test --recover
echo "bottom soccer blue sniff use improve rough use amateur senior transfer quarter" | swisstronikd keys add validator1 --keyring-backend test --recover
echo "wreck layer draw very fame person frown essence approve lyrics sustain spoon" | swisstronikd keys add validator2 --keyring-backend test --recover
echo "exotic merit wrestle sad bundle age purity ability collect immense place tone" | swisstronikd keys add validator3 --keyring-backend test --recover
echo "faculty head please solid picnic benefit hurt gloom flag transfer thrive zebra" | swisstronikd keys add validator4 --keyring-backend test --recover
echo "betray theory cargo way left cricket doll room donkey wire reunion fall left surprise hamster corn village happy bulb token artist twelve whisper expire" | swisstronikd keys add test1 --keyring-backend test --recover
echo "toss sense candy point cost rookie jealous snow ankle electric sauce forward oblige tourist stairs horror grunt tenant afford master violin final genre reason" | swisstronikd keys add test2 --keyring-backend test --recover
swisstronikd add-genesis-account $(swisstronikd keys show validator -a --keyring-backend test) 100000000000000000000000uswtr
swisstronikd add-genesis-account $(swisstronikd keys show validator1 -a --keyring-backend test) 110000000000000000000000uswtr
swisstronikd add-genesis-account $(swisstronikd keys show validator2 -a --keyring-backend test) 120000000000000000000000uswtr
swisstronikd add-genesis-account $(swisstronikd keys show validator3 -a --keyring-backend test) 130000000000000000000000uswtr
swisstronikd add-genesis-account $(swisstronikd keys show validator4 -a --keyring-backend test) 140000000000000000000000uswtr
swisstronikd add-genesis-account $(swisstronikd keys show test1 -a --keyring-backend test) 10000000000000000000000uswtr
swisstronikd add-genesis-account $(swisstronikd keys show test2 -a --keyring-backend test) 10000000000000000000000uswtr
swisstronikd gentx validator 90000000000000000000000uswtr --keyring-backend test --chain-id swisstronik_1291-1
swisstronikd collect-gentxs
sed -i 's/stake/uswtr/g' "$GENESIS"
sed -i 's/pruning = "default"/pruning = "custom"/g' "$CONFIG"
sed -i 's/pruning-keep-recent = "0"/pruning-keep-recent = "2"/g' "$APP_TOML"
sed -i 's/pruning-interval = "0"/pruning-interval = "10"/g' "$APP_TOML"
sed -i 's/127.0.0.1:26657/0.0.0.0:26657/g' "$CONFIG"
sed -i 's/cors_allowed_origins\s*=\s*\[\]/cors_allowed_origins = ["*",]/g' "$CONFIG"
# Enable prometheus on genesis node
sed -i 's/prometheus = false/prometheus = true/g' "$CONFIG"
jq '.app_state["evm"]["params"]["evm_denom"]="uswtr"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.consensus_params["block"]["max_gas"]="10000000000"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["feemarket"]["last_block_gas"]="10000000000"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"