SHELL := /bin/bash
COMPILER = rustc
TARGET = $(shell rustc --version --verbose 2> /dev/null | awk "/host:/ { print \$$2 }")
TARGET_DIR = target/release/
DEFAULT = help
CC=clang
CXX=clang++
SGX_MODE ?= HW
ENCLAVE_HOME ?= $(HOME)/.swisstronik-enclave

build:
	@ENCLAVE_HOME=$(ENCLAVE_HOME) SGX_MODE=$(SGX_MODE) make build -C ../sgxvm/
	@protoc --rust_out src/protobuf_generated/ proto/node.proto
	@ENCLAVE_HOME=$(ENCLAVE_HOME) cargo build --release
	@mkdir -p $(ENCLAVE_HOME) && cp $(CURDIR)/../sgxvm/sgx-artifacts/bin/enclave.signed.so $(ENCLAVE_HOME)/enclave.signed.so
	@cp $(CURDIR)/target/release/libsgxvm_wrapper.so $(CURDIR)/internal/api/libsgx_wrapper.x86_64.so
	@protoc --go_out=types --proto_path=../sgxvm/proto/ ../sgxvm/proto/ffi.proto
	@protoc --go_out=types --proto_path=proto/ proto/node.proto
