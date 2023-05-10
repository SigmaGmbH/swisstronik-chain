init:
	@git submodule update --init --recursive

build-sgx:
	$(MAKE) -C external/evm-module build-librustgo