# Swisstronik Blockchain

Swisstronik is an identity-based hybrid layer-1 blockchain ecosystem. 
It lets Web 3.0 and traditional companies build KYC, AML and DPR compliant applications with enhanced data privacy

## Build

Install submodules by running
```sh 
make init 
```

Build an enclave. For testing purposes you can build enclave in simulation mode by adding `SGX_MODE=SW` 
```sh
make build-enclave
```

Build a chain
```sh
make build
```

## Docker

### Local development node
To build a Docker image, that contains binary for local Swisstronik node, run the following command:
```sh
make build-docker-local
```
This will create an image with the name `swisstronik`. Now it is possible to run the `swisstronikd` binary in the container, 
e.g. checking stored keys:
```sh
docker run -it --rm swisstronik swisstronikd keys list
```