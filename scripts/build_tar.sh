#!/bin/bash

# used to exit on first error (any non-zero exit code)
set -e

rm -rf /tmp/swisstronik-build

mkdir -p /tmp/swisstronik-build
cp -f ./swisstronikd /tmp/swisstronik-build/swisstronikd
chmod +x /tmp/swisstronik-build/swisstronikd

cp -f /usr/lib/.swisstronik-enclave/* /tmp/swisstronik-build
chmod +x /tmp/swisstronik-build/*.so

tar -czvf swisstronik_"$VERSION"_amd64.tar.gz tmp/swisstronik-build

cp ./swisstronik_"$VERSION"_amd64.tar.gz /build/

