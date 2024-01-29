VERSION := v1.0.1
COMMIT := $(shell git log -1 --format='%H')
ENCLAVE_HOME ?= $(HOME)/.swisstronik-enclave

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=swisstronik \
	-X github.com/cosmos/cosmos-sdk/version.ServerName=swisstronikd \
	-X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
	-X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT)

ldflags_static = -X github.com/cosmos/cosmos-sdk/version.Name=swisstronik \
	-X github.com/cosmos/cosmos-sdk/version.ServerName=swisstronikd \
	-X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
	-X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
	-w -s -linkmode 'external' -extldflags=-static

BUILD_FLAGS := -ldflags '$(ldflags)'
BUILD_FLAGS_STATIC := -ldflags "$(ldflags_static)"
DOCKER := $(shell which docker)
DOCKER_BUF := $(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace bufbuild/buf
PROJECT_NAME = $(shell git remote get-url origin | xargs basename -s .git)	

###############################################################################
###                                Protobuf                                 ###
###############################################################################

# ------
# NOTE: Link to the tendermintdev/sdk-proto-gen docker images: 
#       https://hub.docker.com/r/tendermintdev/sdk-proto-gen/tags
#
protoVer=0.11.6
protoImageName=ghcr.io/cosmos/proto-builder:$(protoVer)
protoImage=$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace --user 0 $(protoImageName)
# ------
# NOTE: cosmos/proto-builder image is needed because clang-format is not installed
#       on the tendermintdev/sdk-proto-gen docker image.
#		Link to the cosmos/proto-builder docker images:
#       https://github.com/cosmos/cosmos-sdk/pkgs/container/proto-builder
#
protoCosmosVer=0.11.2
protoCosmosName=ghcr.io/cosmos/proto-builder:$(protoCosmosVer)
protoCosmosImage=$(DOCKER) run --network host --rm -v $(CURDIR):/workspace --workdir /workspace $(protoCosmosName)
# ------
# NOTE: Link to the yoheimuta/protolint docker images:
#       https://hub.docker.com/r/yoheimuta/protolint/tags
#
protolintVer=0.42.2
protolintName=yoheimuta/protolint:$(protolintVer)
protolintImage=$(DOCKER) run --network host --rm -v $(CURDIR):/workspace --workdir /workspace $(protolintName)

proto-all: proto-format proto-lint proto-gen

proto-gen:
	@echo "Generating Protobuf files"
	$(protoImage) sh ./scripts/protocgen.sh


proto-format:
	@echo "Formatting Protobuf files"
	$(protoCosmosImage) find ./ -name *.proto -exec clang-format -i {} \;

# NOTE: The linter configuration lives in .protolint.yaml
proto-lint:
	@echo "Linting Protobuf files"
	$(protolintImage) lint ./proto

proto-check-breaking:
	@echo "Checking Protobuf files for breaking changes"
	$(protoImage) buf breaking --against $(HTTPS_GIT)#branch=main


.PHONY: proto-all proto-gen proto-gen-any proto-format proto-lint proto-check-breaking

###############################################################################
###                                  Build                                  ###
###############################################################################

all: install

init:
	@git submodule update --init --recursive

install: go.sum
	go install -mod=readonly $(BUILD_FLAGS) ./cmd/swisstronikd

build: go.sum
	$(MAKE) -C go-sgxvm build
	go build -mod=mod $(BUILD_FLAGS)  -tags osusergo,netgo -o build/swisstronikd ./cmd/swisstronikd

###############################################################################
### 		          Build commands for CLI (without SGX support) 			###
###############################################################################

build-cli: go.sum
	go build -mod=mod $(BUILD_FLAGS) -tags osusergo,netgo,nosgx -o build/$(BINARY_NAME) ./cmd/swisstronikd

build-cli-linux-static: go.sum
	CGO_ENABLED=1  go build -mod=mod $(BUILD_FLAGS_STATIC) -tags osusergo,netgo,nosgx -o build/$(BINARY_NAME) ./cmd/swisstronikd

build-linux:
	GOOS=linux GOARCH=$(if $(findstring aarch64,$(shell uname -m)) || $(findstring arm64,$(shell uname -m)),arm64,amd64) $(MAKE) build

build-macos-cli-amd: 
	BINARY_NAME=swisstronikcli-macos-amd64 GOOS=darwin GOARCH=amd64 $(MAKE) build-cli

build-macos-cli-arm:
	BINARY_NAME=swisstronikcli-macos-arm64 GOOS=darwin GOARCH=arm64 $(MAKE) build-cli

build-linux-cli-amd:
	BINARY_NAME=swisstronikcli-linux-amd64 GOOS=linux GOARCH=amd64 $(MAKE) build-cli-linux-static

build-linux-cli-arm:
	BINARY_NAME=swisstronikcli-linux-arm64 GOOS=linux GOARCH=arm64 $(MAKE) build-cli-linux-static

build-windows-cli:
	BINARY_NAME=swisstronikcli-windows GOOS=windows GOARCH=amd64 $(MAKE) build-cli

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	GO111MODULE=on go mod verify

test:
	go test --cover -short -p 1 ./...

build-docker-local:
	docker build -f docker/node.Dockerfile -t swisstronik --target=local-node --build-arg SGX_MODE=SW .

.PHONY: all install build build-linux build-enclave test build-docker-local

###############################################################################
###                                  Tests                                  ###
###############################################################################

test: test-unit
test-all: test-unit test-race
PACKAGES_UNIT=$(shell go list ./... | grep -Ev 'vendor|importer')
TEST_PACKAGES=./...
TEST_TARGETS := test-unit test-unit-cover test-race

# Test runs-specific rules. To add a new test target, just add
# a new rule, customise ARGS or TEST_PACKAGES ad libitum, and
# append the new rule to the TEST_TARGETS list.
test-unit: ARGS=-timeout=10m -race
test-unit: TEST_PACKAGES=$(PACKAGES_UNIT)

test-race: ARGS=-race
test-race: TEST_PACKAGES=$(PACKAGES_NOSIMULATION)
$(TEST_TARGETS): run-tests

test-unit-cover: ARGS=-timeout=10m -race -coverprofile=coverage.txt -covermode=atomic
test-unit-cover: TEST_PACKAGES=$(PACKAGES_UNIT)

run-tests:
ifneq (,$(shell which tparse 2>/dev/null))
	go test -mod=readonly -json $(ARGS) $(EXTRA_ARGS) $(TEST_PACKAGES) | tparse
else
	go test -mod=readonly $(ARGS)  $(EXTRA_ARGS) $(TEST_PACKAGES)
endif

.PHONY: run-tests test test-all $(TEST_TARGETS)