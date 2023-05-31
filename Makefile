VERSION := 1.0.0
COMMIT := $(shell git log -1 --format='%H')
ENCLAVE_HOME ?= $(HOME)/.swisstronik-enclave

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=swisstronik \
	-X github.com/cosmos/cosmos-sdk/version.ServerName=swisstronikd \
	-X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
	-X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT)

###############################################################################
###                                  Build                                  ###
###############################################################################

BUILD_FLAGS := -ldflags '$(ldflags)'
DOCKER := $(shell which docker)
DOCKER_BUF := $(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace bufbuild/buf
PROJECT_NAME = $(shell git remote get-url origin | xargs basename -s .git)

all: install

init:
	@git submodule update --init --recursive

install: go.sum
	go install -mod=readonly $(BUILD_FLAGS) ./cmd/swisstronikd

build: go.sum
	go build -mod=mod $(BUILD_FLAGS) -o build/swisstronikd ./cmd/swisstronikd

build-linux:
	GOOS=linux GOARCH=$(if $(findstring aarch64,$(shell uname -m)) || $(findstring arm64,$(shell uname -m)),arm64,amd64) $(MAKE) build

build-enclave:
	$(MAKE) -C external/evm-module build-librustgo

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	GO111MODULE=on go mod verify

test:
	go test --cover -short -p 1 ./...

build-docker-local:
	docker build -f docker/node.Dockerfile -t swisstronik --target=local-node .

.PHONY: all install build build-linux build-enclave test build-docker-local
