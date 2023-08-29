# go-sgxvm

This is a wrapper around the [Sputnik VM](https://github.com/rust-blockchain/evm).
It allows you to compile, initialize and execute EVM smart contracts from Go applications.

It was forked from https://github.com/CosmWasm/wasmvm 


## Build SGX-EVM & SGX-Wrapper

Ensure that SGX SDK was installed to `/opt/intel/` directory

Then run:
`source /opt/intel/sgxsdk/environment`

Now you are ready to build enclave with SGX-EVM and wrapper around it. To do it, run:
`make build`

If you want to build SGX-EVM enclave only, run: `make sgx`

## Structure

This repo contains both Rust and Go code. The rust code is compiled into a dll/so
to be linked via cgo and wrapped with a pleasant Go API. The full build step
involves compiling rust -> C library, and linking that library to the Go code.
For ergonomics of the user, we will include pre-compiled libraries to easily
link with, and Go developers should just be able to import this directly.

## Docs

Run `(cd libsgx_wrapper && cargo doc --no-deps --open)`.

## Design

To understand how Cosmos SDK and EVM Keeper interacts with this library, you can refer to diagram below:

![plot](./spec/sgxsequence.png)

## Development

There are two halfs to this code - go and rust. The first step is to ensure that there is
a proper dll built for your platform. This should be `api/libsgx_wrapper.X`, where X is:

- `so` for Linux systems
- `dylib` for MacOS
- `dll` for Windows - Not currently supported due to upstream dependency

If this is present, then `make test` will run the Go test suite and you can import this code freely.
If it is not present you will have to build it for your system, and ideally add it to this repo
with a PR (on your fork). We will set up a proper CI system for building these binaries,
but we are not there yet.

To build the rust side, try `make build-rust` and wait for it to compile. This depends on
`cargo` being installed with `rustc` version 1.47+. Generally, you can just use `rustup` to
install all this with no problems.

## Toolchain

For development you should be able to use any reasonably up-to-date Rust stable.
