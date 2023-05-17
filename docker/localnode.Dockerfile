FROM ubuntu:22.04

# Install Intel SGX SDK
RUN apt-get update
RUN apt-get install build-essential ocaml ocamlbuild automake autoconf libtool wget python-is-python3 libssl-dev git cmake perl -y

RUN wget https://download.01.org/intel-sgx/sgx-linux/2.19/distro/ubuntu22.04-server/sgx_linux_x64_sdk_2.19.100.3.bin
RUN chmod +x sgx_linux_x64_sdk_2.19.100.3.bin
RUN ./sgx_linux_x64_sdk_2.19.100.3.bin --prefix /opt/intel

# Build enclave
RUN apt-get install protobuf-compiler curl golang -y
RUN curl https://sh.rustup.rs -sSf | bash -s -- -y
ENV PATH="/root/.cargo/bin:${PATH}"
RUN cargo install protobuf-codegen --version "2.8.1" -f

WORKDIR /root

COPY . .

ENV SGX_SDK="/opt/intel/sgxsdk"
ENV PATH="${PATH}:${SGX_SDK}/bin:${SGX_SDK}/bin/x64"
ENV PKG_CONFIG_PATH="${PKG_CONFIG_PATH}:${SGX_SDK}/pkgconfig"
ENV LD_LIBRARY_PATH="/opt/intel/sgxsdk/sdk_libs:${LD_LIBRARY_PATH}"

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
ENV GO_PATH="/root/go"
ENV PATH="${PATH}:/${GO_PATH}/bin"

RUN SGX_MODE=SW make build-enclave
RUN make install

EXPOSE 26656 26657 1317 9090 8535 8546
CMD ["swisstronikd"]