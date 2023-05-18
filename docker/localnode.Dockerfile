FROM ubuntu:22.04 AS build

RUN \
    apt-get update && \
    apt-get install build-essential automake autoconf  \
    libtool wget python-is-python3 libssl-dev git cmake perl protobuf-compiler curl golang -y

RUN wget https://download.01.org/intel-sgx/sgx-linux/2.19/distro/ubuntu22.04-server/sgx_linux_x64_sdk_2.19.100.3.bin
RUN chmod +x sgx_linux_x64_sdk_2.19.100.3.bin
RUN ./sgx_linux_x64_sdk_2.19.100.3.bin --prefix /opt/intel

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
RUN make build

FROM ubuntu:22.04

WORKDIR /root
COPY --from=build /root/build/swisstronikd /usr/local/bin/swisstronikd
COPY --from=build /root/.swisstronik-enclave /root/.swisstronik-enclave
COPY --from=build /root/external/evm-module/librustgo/internal/api/libsgx_wrapper.x86_64.so /lib/x86_64-linux-gnu/libsgx_wrapper.x86_64.so
COPY --from=build /opt/intel/sgxsdk/sdk_libs/* /lib/x86_64-linux-gnu/

EXPOSE 26656 26657 1317 9090 8535 8546
CMD ["swisstronikd"]