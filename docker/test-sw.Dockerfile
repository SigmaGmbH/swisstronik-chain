############ Install Intel SGX SDK & SGX PSW
FROM ghcr.io/sigmagmbh/sgx:2.23-jammy-554238b as base
RUN wget -qO - https://download.01.org/intel-sgx/sgx_repo/ubuntu/intel-sgx-deb.key | apt-key add -
RUN apt-get update


############ Compilation base
FROM base as compile-base

RUN apt-get install -y protobuf-compiler curl

# Install rust
ENV PATH="/usr/local/go/bin:/go/bin:/root/.cargo/bin:${PATH}"
ENV GOROOT=/usr/local/go
ENV GOPATH=/go/

RUN curl https://sh.rustup.rs -sSf | bash -s -- -y > /dev/null 2>&1
RUN cargo install protobuf-codegen --version "2.8.1" -f
 
# Install golang
ADD https://go.dev/dl/go1.22.5.linux-amd64.tar.gz go.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go.linux-amd64.tar.gz && rm go.linux-amd64.tar.gz
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.2 && \
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.4.0 > /dev/null 2>&1



############ Compile enclave & chain
FROM compile-base as compile-chain

RUN apt-get install -y automake autoconf build-essential libtool git 

ARG SGX_MODE=SW
ENV SGX_MODE=${SGX_MODE}
ARG PRODUCTION_MODE=true
ENV PRODUCTION_MODE=${PRODUCTION_MODE}
ENV SGX_SDK="/opt/intel/sgxsdk"
ENV PATH="${PATH}:${SGX_SDK}/bin:${SGX_SDK}/bin/x64"
ENV PKG_CONFIG_PATH="${PKG_CONFIG_PATH}:${SGX_SDK}/pkgconfig"
ENV LD_LIBRARY_PATH="/opt/intel/sgxsdk/sdk_libs:${LD_LIBRARY_PATH}"

COPY . /root/chain
WORKDIR /root/chain
RUN make build
RUN ./build/swisstronikd testnet init-testnet-enclave
RUN make test-all