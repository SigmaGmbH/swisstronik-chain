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

############ Node binary
FROM base as node

COPY --from=compile-chain /root/chain/build/swisstronikd /usr/local/bin/swisstronikd
COPY --from=compile-chain /root/.swisstronik-enclave /root/.swisstronik-enclave
COPY --from=compile-chain /root/chain/external/evm-module/librustgo/internal/api/libsgx_wrapper.x86_64.so /lib/x86_64-linux-gnu/libsgx_wrapper.x86_64.so
COPY --from=compile-chain /opt/intel /opt/intel

EXPOSE 26656 26657 1317 9090 8535 8546 8999
CMD ["swisstronikd"]





















# FROM ubuntu:22.04 AS build
# 
# RUN \
#     apt-get update && \
#     apt-get install build-essential automake autoconf  \
#     libtool wget python-is-python3 libssl-dev git cmake perl protobuf-compiler curl golang -y
# 
# RUN wget https://download.01.org/intel-sgx/sgx-linux/2.19/distro/ubuntu22.04-server/sgx_linux_x64_sdk_2.19.100.3.bin
# RUN chmod +x sgx_linux_x64_sdk_2.19.100.3.bin
# RUN ./sgx_linux_x64_sdk_2.19.100.3.bin --prefix /opt/intel
# 
# RUN curl https://sh.rustup.rs -sSf | bash -s -- -y
# ENV PATH="/root/.cargo/bin:${PATH}"
# RUN cargo install protobuf-codegen --version "2.8.1" -f
# 
# WORKDIR /root
# 
# COPY . .
# 
# ENV SGX_SDK="/opt/intel/sgxsdk"
# ENV PATH="${PATH}:${SGX_SDK}/bin:${SGX_SDK}/bin/x64"
# ENV PKG_CONFIG_PATH="${PKG_CONFIG_PATH}:${SGX_SDK}/pkgconfig"
# ENV LD_LIBRARY_PATH="/opt/intel/sgxsdk/sdk_libs:${LD_LIBRARY_PATH}"
# 
# RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
# RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
# ENV GO_PATH="/root/go"
# ENV PATH="${PATH}:/${GO_PATH}/bin"
# 
# RUN make init
# RUN SGX_MODE=SW make build-enclave
# RUN make build
# 
# FROM ubuntu:22.04
# 
# WORKDIR /root
# COPY --from=build /root/build/swisstronikd /usr/local/bin/swisstronikd
# COPY --from=build /root/.swisstronik-enclave /root/.swisstronik-enclave
# COPY --from=build /root/external/evm-module/librustgo/internal/api/libsgx_wrapper.x86_64.so /lib/x86_64-linux-gnu/libsgx_wrapper.x86_64.so
# COPY --from=build /opt/intel/sgxsdk/sdk_libs/* /lib/x86_64-linux-gnu/
# 
# EXPOSE 26656 26657 1317 9090 8535 8546
# CMD ["swisstronikd"]