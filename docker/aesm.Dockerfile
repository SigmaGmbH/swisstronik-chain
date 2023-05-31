############ Install Intel SGX SDK & SGX PSW
FROM ghcr.io/initc3/linux-sgx:2.19-jammy as base
RUN wget -qO - https://download.01.org/intel-sgx/sgx_repo/ubuntu/intel-sgx-deb.key | apt-key add -
RUN apt-get update

ENV AESM_PATH=/opt/intel/sgx-aesm-service/aesm
ENV LD_LIBRARY_PATH=/opt/intel/sgx-aesm-service/aesm

WORKDIR /opt/intel/sgx-aesm-service/aesm

ENTRYPOINT ["/opt/intel/sgx-aesm-service/aesm/aesm_service", "--no-daemon"]