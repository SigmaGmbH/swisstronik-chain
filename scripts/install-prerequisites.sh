sudo apt update
sudo apt upgrade -y
sudo apt install build-essential gcc git jq chrony -y
sudo apt -y install cargo
wget -q -O - https://raw.githubusercontent.com/canha/golang-tools-install-script/master/goinstall.sh | bash -s -- --version 1.19
source ~/.profile
cargo install protobuf-codegen --version "2.8.1" -f
go install cosmossdk.io/tools/cosmovisor/cmd/cosmovisor@latest
source /opt/intel/sgxsdk/environment
sudo rm -rf $HOME/chain
git clone --recurse-submodules https://github.com/SigmaGmbH/chain.git
cd $HOME/chain/ && go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
cd $HOME/chain/ && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
cd $HOME/chain/external/evm-module && go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
cd $HOME/chain/external/evm-module && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
cd $HOME/chain/ && SGX_MODE=SW make build-enclave
cd $HOME/chain/ && make install
export DAEMON_NAME=swisstronikd
export DAEMON_HOME=$HOME/.swisstronik
source ~/.profile
