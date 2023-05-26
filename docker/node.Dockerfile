############ Install Intel SGX SDK & SGX PSW
FROM ubuntu:22.04 as base

WORKDIR /root

RUN apt-get update && apt-get install build-essential wget libssl-dev libcurl4-openssl-dev libprotobuf-dev -y
RUN mkdir sgx && mkdir /etc/init
ADD https://download.01.org/intel-sgx/sgx-linux/2.19/distro/ubuntu22.04-server/sgx_linux_x64_sdk_2.19.100.3.bin ./sgx
RUN chmod +x ./sgx/sgx_linux_x64_sdk_2.19.100.3.bin
RUN ./sgx/sgx_linux_x64_sdk_2.19.100.3.bin --prefix /opt/intel
RUN echo "source /opt/intel/sgxsdk/environment" >> /root/.bashrc && rm -rf ./sgx/*

RUN echo 'deb [arch=amd64] https://download.01.org/intel-sgx/sgx_repo/ubuntu jammy main' | tee /etc/apt/sources.list.d/intelsgx.list
RUN wget -qO - https://download.01.org/intel-sgx/sgx_repo/ubuntu/intel-sgx-deb.key | apt-key add -
RUN apt-get update
RUN apt-get install -y libsgx-launch libsgx-urts libsgx-epid libsgx-quote-ex libsgx-dcap-ql sgx-aesm-service libsgx-aesm-launch-plugin libsgx-aesm-epid-plugin



############ Compilation base
FROM base as compile-base

RUN apt-get install -y protobuf-compiler curl

# Install rust
RUN curl https://sh.rustup.rs -sSf | bash -s -- -y > /dev/null 2>&1
ENV PATH="/root/.cargo/bin:${PATH}"

RUN cargo install protobuf-codegen --version "2.8.1" -f > /dev/null 2>&1

# Install golang
ENV GOROOT=/usr/local/go
ENV GOPATH=/go/
ENV PATH=$PATH:/usr/local/go/bin:$GOPATH/bin
ADD https://go.dev/dl/go1.19.linux-amd64.tar.gz go.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go.linux-amd64.tar.gz && rm go.linux-amd64.tar.gz
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest > /dev/null 2>&1



############ Compile enclave & chain
FROM compile-base as compile-chain

RUN apt-get install -y automake autoconf build-essential libtool git

WORKDIR /root

COPY . ./chain

WORKDIR /root/chain

ARG SGX_MODE=HW
ENV SGX_MODE=${SGX_MODE}
ENV SGX_SDK="/opt/intel/sgxsdk"
ENV PATH="${PATH}:${SGX_SDK}/bin:${SGX_SDK}/bin/x64"
ENV PKG_CONFIG_PATH="${PKG_CONFIG_PATH}:${SGX_SDK}/pkgconfig"
ENV LD_LIBRARY_PATH="/opt/intel/sgxsdk/sdk_libs:${LD_LIBRARY_PATH}"

RUN SGX_MODE=${SGX_MODE} make build-enclave
RUN make build



############ Node binary in Hardware Mode
FROM base as hw-node

COPY --from=compile-chain /root/chain/build/swisstronikd /usr/local/bin/swisstronikd
COPY --from=compile-chain /root/.swisstronik-enclave /root/.swisstronik-enclave
COPY --from=compile-chain /root/chain/external/evm-module/librustgo/internal/api/libsgx_wrapper.x86_64.so /lib/x86_64-linux-gnu/libsgx_wrapper.x86_64.so
COPY --from=compile-chain /opt/intel /opt/intel

EXPOSE 26656 26657 1317 9090 8535 8546 8999
CMD ["swisstronikd"]



############ Node binary in Software Mode
FROM ubuntu:22.04 as local-node

WORKDIR /root

RUN apt-get update && apt-get install -y jq

COPY --from=compile-chain /root/chain/build/swisstronikd /usr/local/bin/swisstronikd
COPY --from=compile-chain /root/.swisstronik-enclave /root/.swisstronik-enclave
COPY --from=compile-chain /root/chain/external/evm-module/librustgo/internal/api/libsgx_wrapper.x86_64.so /lib/x86_64-linux-gnu/libsgx_wrapper.x86_64.so
COPY --from=compile-chain /opt/intel/sgxsdk/sdk_libs/* /lib/x86_64-linux-gnu/

# TODO: Run bash script for node setup

EXPOSE 26656 26657 1317 9090 8535 8546 8999