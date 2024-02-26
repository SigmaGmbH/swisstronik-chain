#!/bin/bash

set -e 

make build && cd build && ./swisstronikd enclave request-master-key-dcap rpc.testnet.swisstronik.com:46789