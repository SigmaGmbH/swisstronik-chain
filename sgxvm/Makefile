SHELL := /bin/bash
COMPILER = rustc
TARGET = $(shell rustc --version --verbose 2> /dev/null | awk "/host:/ { print \$$2 }")
TARGET_DIR = target/release/
DEFAULT = help
CC=clang
CXX=clang++
SGX_MODE ?= HW
ENCLAVE_HOME ?= $(HOME)/.swisstronik-enclave

Trts_Library_Name = sgx_trts
Service_Library_Name = sgx_tservice
Enclave_build_feature = hardware_mode

# ENCLAVE SETTINGS
ifneq ($(SGX_MODE), HW)
	Trts_Library_Name := sgx_trts_sim
	Service_Library_Name := sgx_tservice_sim
	Enclave_build_feature := simulation_mode
else
	Trts_Library_Name := sgx_trts
	Service_Library_Name := sgx_tservice
	Enclave_build_feature := hardware_mode
endif

# DEFINEs

define sgx_clean
	@echo "Cleaning enclave dependencies"
	@rm -rf $(CURDIR)/sgx-artifacts/bin/*
	@rm -rf $(CURDIR)/sgx-artifacts/lib/*
	@rm -f $(CURDIR)/Enclave_u*
	@rm -rf $(CURDIR)/target
	@rm -f $(CURDIR)/Enclave_t*
	@rm -f $(CURDIR)/enclave.unsigned.so
endef

define compile_protobuf
	@echo "Compiling SGXVM protobuf files"
    protoc --rust_out src/protobuf_generated/ proto/ffi.proto
    sed -i -e 's/use protobuf::Message as/\n\nuse std::prelude::v1::*;\nuse protobuf::Message as/g' src/protobuf_generated/ffi.rs
endef

define compile_enclave_rust
	@echo "Building enclave rust code"
	@RUSTFLAGS="-C target-cpu=native" cargo build --release --features $(Enclave_build_feature) --no-default-features
endef

define create_bridge_enclave_rust
	@echo "Create bridge between enclave and rust"
	@/opt/intel/sgxsdk/bin/x64/sgx_edger8r --trusted $(CURDIR)/Enclave.edl --search-path /opt/intel/sgxsdk/include --search-path $(CURDIR)/sgx-sdk/edl --trusted-dir $(CURDIR)/
	@/opt/intel/sgxsdk/bin/x64/sgx_edger8r --untrusted $(CURDIR)/Enclave.edl --search-path /opt/intel/sgxsdk/include --search-path $(CURDIR)/sgx-sdk/edl --untrusted-dir $(CURDIR)/
endef

define compile_enclave_bridge
    @echo "Compile files generated by sgx_edger8r using cc"
	@cc -m64 -O2 -fstack-protector -fPIC -Wno-attributes -I $(CURDIR)/ -I$(CURDIR)/include -I/opt/intel/sgxsdk/include -I$(CURDIR)/sgx-sdk/edl -c $(CURDIR)/Enclave_u.c -o $(CURDIR)/Enclave_u.o
endef

define create_archive
	@echo "Compress files"
	@mkdir -p $(CURDIR)/sgx-artifacts/lib
	@ar rcsD $(CURDIR)/sgx-artifacts/lib/libEnclave_u.a $(CURDIR)/Enclave_u.o
endef

define copy_compiled_enclave
 	@echo "Copy compiled enclave files"
	@cp $(CURDIR)/target/release/libenclave.a $(CURDIR)/sgx-artifacts/lib/libenclave.a
endef

define compile_into_enclave_t_o
	@echo "Compile into Enclave_t.o"
	@cc -m64 -O2 -fstack-protector -ffreestanding -nostdinc -fvisibility=hidden -fpie -fno-strict-overflow -fno-delete-null-pointer-checks -I$(CURDIR)/sgx-sdk/common/inc -I$(CURDIR)/sgx-sdk/edl -I/opt/intel/sgxsdk/include -I/opt/intel/sgxsdk/include/tlibc -I/opt/intel/sgxsdk/include/stlport -I/opt/intel/sgxsdk/include/epid -I $(CURDIR)/ -I$(CURDIR)/include -c $(CURDIR)/Enclave_t.c -o $(CURDIR)/Enclave_t.o
endef

define compile_unsigned_enclave
	@echo "Compile into unsinged enclave"
	@g++ $(CURDIR)/Enclave_t.o -o $(CURDIR)/enclave.unsigned.so -Wl,--no-undefined -nostdlib -nodefaultlibs -nostartfiles -L/opt/intel/sgxsdk/lib64 \
		-Wl,--whole-archive -l$(Trts_Library_Name) -Wl,--no-whole-archive -Wl,--start-group -lsgx_tstdc -lsgx_tcxx -l$(Service_Library_Name) -lsgx_tcrypto -lsgx_tprotected_fs -lpthread \
		-L$(CURDIR)/sgx-artifacts/lib -lenclave -Wl,--end-group -Wl,--version-script=$(CURDIR)/Enclave.lds -Wl,-z,relro,-z,now,-z,noexecstack -Wl,-Bstatic -Wl,-Bsymbolic \
		-Wl,--no-undefined -Wl,-pie,-eenclave_entry -Wl,--export-dynamic -Wl,--gc-sections -Wl,--defsym,__ImageBase=0
endef

define sign_enclave
	@echo "Sign enclave"
	@mkdir -p $(CURDIR)/sgx-artifacts/bin
	@/opt/intel/sgxsdk/bin/x64/sgx_sign sign -key $(CURDIR)/Enclave_private.pem -enclave $(CURDIR)/enclave.unsigned.so -out $(CURDIR)/sgx-artifacts/bin/enclave.signed.so -config $(CURDIR)/Enclave.config.xml
endef

define sgx_build
	$(call sgx_clean)
	$(call compile_protobuf)
	$(call compile_enclave_rust)
	$(call create_bridge_enclave_rust)
	$(call compile_enclave_bridge)
	$(call create_archive)
	$(call copy_compiled_enclave)
	$(call compile_into_enclave_t_o)
	$(call compile_unsigned_enclave)
	$(call sign_enclave)
endef

build:
	$(call sgx_build)
	@echo "Intel SGX enclave built and signed"

clean:
	$(call sgx_clean)