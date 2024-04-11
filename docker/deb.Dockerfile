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
ADD https://go.dev/dl/go1.19.linux-amd64.tar.gz go.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go.linux-amd64.tar.gz && rm go.linux-amd64.tar.gz
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest > /dev/null 2>&1



############ Compile enclave & chain
FROM compile-base as compile-chain

RUN apt-get install -y automake autoconf build-essential libtool git 

ARG SGX_MODE=HW
ENV SGX_MODE=${SGX_MODE}
ENV SGX_SDK="/opt/intel/sgxsdk"
ENV PATH="${PATH}:${SGX_SDK}/bin:${SGX_SDK}/bin/x64"
ENV PKG_CONFIG_PATH="${PKG_CONFIG_PATH}:${SGX_SDK}/pkgconfig"
ENV LD_LIBRARY_PATH="/opt/intel/sgxsdk/sdk_libs:${LD_LIBRARY_PATH}"

COPY . /root/chain
WORKDIR /root/chain
RUN make build


############ Node binary for deb package
FROM compile-base as build-deb

ARG BUILD_VERSION="v1.0.1"
ENV VERSION=${BUILD_VERSION}
ARG DEB_BIN_DIR=/usr/local/bin
ENV DEB_BIN_DIR=${DEB_BIN_DIR}
ARG DEB_LIB_DIR=/usr/lib
ENV DEB_LIB_DIR=${DEB_LIB_DIR}
ARG ENCLAVE_HOME=${DEB_LIB_DIR}
ARG ENCLAVE_HOME=${ENCLAVE_HOME}

WORKDIR /root

# Copy over binaries from the build-env
COPY --from=compile-chain /root/chain/build/swisstronikd swisstronikd
COPY --from=compile-chain /root/.swisstronik-enclave /usr/lib/.swisstronik-enclave
COPY --from=compile-chain /root/chain/go-sgxvm/internal/api/libsgx_wrapper.x86_64.so /usr/lib/.swisstronik-enclave/libsgx_wrapper.x86_64.so

COPY ./deb ./deb
COPY ./scripts/build_deb.sh .

RUN chmod +x build_deb.sh

CMD ["/bin/bash", "build_deb.sh"]
