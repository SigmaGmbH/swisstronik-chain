# Swisstronik Blockchain

Swisstronik is an identity-based hybrid layer-1 blockchain ecosystem. 
It lets Web 3.0 and traditional companies build KYC, AML and DPR compliant applications with enhanced data privacy

[![Build local node docker image](https://github.com/SigmaGmbH/swisstronik-chain/actions/workflows/docker-local.yml/badge.svg)](https://github.com/SigmaGmbH/swisstronik-chain/actions/workflows/docker-local.yml)

## Build

Install submodules by running
```sh 
make init 
```

Build an enclave. For testing purposes you can build enclave in simulation mode by adding `SGX_MODE=SW`. To specify path for output with compiled enclave you can use `ENCLAVE_HOME` env arg.
```sh
make build-enclave
```

Build a chain
```sh
make build
```

## Docker

### Local development node
Before building ensure that you initialized all submodules. You can do that by running:
```sh
make init
```

To build a Docker image, that contains binary for local Swisstronik node, run the following command:
```sh
make build-docker-local
```
This will create an image with the name `swisstronik` and `latest` version tag. Now it is possible to run the `swisstronikd` binary in the container, 
e.g. checking stored keys:
```sh
docker run -it --rm swisstronik swisstronikd keys list
```

### Local testnet
To setup local test network with multiple validators, run:
```sh
swisstronikd testnet init-config --starting-ip-address 192.167.10.1 --chain-id swisstronik_1291-1
```

Then run:
```sh
docker-compose -f local-network.yml up
```

### Monitoring

#### Enable monitoring
To enable monitoring for your node, first check if prometheus is enabled (`prometheus = true`) in `config.toml`,
located at `$HOME/.swisstronik/config` by default. To enable it, simply run:
```sh
sed -i 's/prometheus = false/prometheus = true/g' <YOUR-NODE-HOMEDIR>/config/config.toml
```
Also, you need to enable telemetry in `app.toml`. To enable it, change `enabled` to `true` 
```
[telemetry]
  enable-hostname = false
  enable-hostname-label = false
  enable-service-label = false
  enabled = false 
  global-labels = []
  prometheus-retention-time = 0
  service-name = ""
```

Then you need to restart your node. After that you should be able to access the tendermint metrics (default port is 26660)

#### Configure Prometheus Targets
Update target with address of your node in `monitoring/prometheus.yml`. This will tell prometheus from where it should obtain metrics

#### Setup Prometheus and Grafana
You can start docker containers with prometheus and grafana using docker-compose. To do it, run:
```sh
docker-compose run up -d
```