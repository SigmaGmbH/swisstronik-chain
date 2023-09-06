#!/usr/bin/env bash

# --------------
# Commands to run locally
# docker run --network host --rm -v $(CURDIR):/workspace --workdir /workspace ghcr.io/cosmos/proto-builder:v0.11.6 sh ./protocgen.sh
#
set -eo pipefail

echo "Generating gogo proto code"
proto_dirs=$(find ./proto -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  proto_files=$(find "${dir}" -maxdepth 1 -name '*.proto')
  for file in $proto_files; do
    # Check if the go_package in the file is pointing to swisstronik
    if grep -q "option go_package.*swisstronik" "$file"; then
      buf generate --template proto/buf.gen.gogo.yaml "$file"
    fi

    # Check if the go_package in the file is pointing to evmos / ethermint
    if grep -q "option go_package.*evmos" "$file"; then
      buf generate --template proto/buf.gen.gogo.yaml "$file"
    fi
  done
done

cp -r github.com/evmos/ethermint/* ./
cp -r swisstronik/* ./
rm -rf github.com
rm -rf swisstronik