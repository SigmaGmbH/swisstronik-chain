use crate::errors::{handle_c_error_default, Error};
use crate::memory::{ByteSliceView, UnmanagedVector};
use crate::protobuf_generated::node;
use crate::types::{Allocation, AllocationWithResult, GoQuerier};

use lazy_static::lazy_static;
use protobuf::Message;
use sgx_types::*;
use std::panic::catch_unwind;

pub mod enclave_api;
pub mod doorbell;

lazy_static! {
    pub static ref ENCLAVE_DOORBELL: doorbell::EnclaveDoorbell = doorbell::EnclaveDoorbell::new();
}

// store some common string for argument names
pub const PB_REQUEST_ARG: &str = "pb_request";

#[allow(dead_code)]
extern "C" {
    pub fn handle_request(
        eid: sgx_enclave_id_t,
        retval: *mut AllocationWithResult,
        querier: *mut GoQuerier,
        request: *const u8,
        len: usize,
    ) -> sgx_status_t;

    pub fn ecall_allocate(
        eid: sgx_enclave_id_t,
        retval: *mut Allocation,
        data: *const u8,
        len: usize,
    ) -> sgx_status_t;

    pub fn ecall_init_master_key(
        eid: sgx_enclave_id_t,
        retval: *mut sgx_status_t,
        reset_flag: i32,
    ) -> sgx_status_t;

    pub fn ecall_is_initialized(eid: sgx_enclave_id_t, retval: *mut i32) -> sgx_status_t;

    pub fn ecall_share_seed(
        eid: sgx_enclave_id_t,
        retval: *mut sgx_status_t,
        socket_fd: c_int,
    ) -> sgx_status_t;

    pub fn ecall_request_seed(
        eid: sgx_enclave_id_t,
        retval: *mut sgx_status_t,
        hostname: *const u8,
        data_len: usize,
        socket_fd: c_int,
    ) -> sgx_status_t;

    pub fn ecall_status(eid: sgx_enclave_id_t, retval: *mut sgx_status_t) -> sgx_status_t;

    pub fn ecall_dcap_attestation(
        eid: sgx_enclave_id_t,
        retval: *mut sgx_status_t,
        hostname: *const u8,
        data_len: usize,
        socket_fd: c_int,
        qe_target_info: &sgx_target_info_t,
        quote_size: u32,
    ) -> sgx_status_t;

    pub fn ecall_create_report(
        eid: sgx_enclave_id_t,
        retval: *mut sgx_status_t,
        p_qe3_target: *const sgx_target_info_t,
        p_report: *mut sgx_report_t,
    ) -> sgx_status_t;

    pub fn ecall_get_target_info(
        eid: sgx_enclave_id_t,
        retval: *mut sgx_status_t,
        target_info: *mut sgx_target_info_t,
    ) -> sgx_status_t;

    pub fn sgx_tvl_verify_qve_report_and_identity(
        eid: sgx_enclave_id_t,
        retval: *mut sgx_quote3_error_t,
        p_quote: *const uint8_t,
        quote_size: uint32_t,
        p_qve_report_info: *const sgx_ql_qe_report_info_t,
        expiration_check_date: time_t,
        collateral_expiration_status: uint32_t,
        quote_verification_result: sgx_ql_qv_result_t,
        p_supplemental_data: *const uint8_t,
        supplemental_data_size: uint32_t,
        qve_isvsvn_threshold: sgx_isv_svn_t,
    ) -> sgx_quote3_error_t;
}

#[no_mangle]
/// Handles all incoming protobuf-encoded requests related to node setup
/// such as generating of attestation certificate, keys, etc.
pub unsafe extern "C" fn handle_initialization_request(
    request: ByteSliceView,
    error_msg: Option<&mut UnmanagedVector>,
) -> UnmanagedVector {
    let r = catch_unwind(|| {
        // Check if request is correct
        let req_bytes = request
            .read()
            .ok_or_else(|| Error::unset_arg(PB_REQUEST_ARG))?;

        let request = match protobuf::parse_from_bytes::<node::SetupRequest>(req_bytes) {
            Ok(request) => request,
            Err(e) => {
                return Err(Error::protobuf_decode(e.to_string()));
            }
        };

        let enclave_access_token = crate::enclave::ENCLAVE_DOORBELL
            .get_access(1) // This can never be recursive
            .ok_or(sgx_status_t::SGX_ERROR_BUSY)?;
        let evm_enclave = (*enclave_access_token)?;

        let result = match request.req {
            Some(req) => {
                match req {
                    node::SetupRequest_oneof_req::nodeStatus(_req) => {
                        enclave_api::EnclaveApi::check_node_status(evm_enclave.geteid())?;
                        let response = node::NodeStatusResponse::new();
                        let response_bytes: Vec<u8> = response.write_to_bytes()?;
                        Ok(response_bytes)
                    }
                    node::SetupRequest_oneof_req::initializeMasterKey(req) => {
                        enclave_api::EnclaveApi::initialize_master_key(evm_enclave.geteid(), req.shouldReset)?;
                        let response = node::InitializeMasterKeyResponse::new();
                        let response_bytes = response.write_to_bytes()?;
                        Ok(response_bytes)
                    }
                    node::SetupRequest_oneof_req::startBootstrapServer(req) => {
                        enclave_api::EnclaveApi::start_bootstrap_server(evm_enclave.geteid(), req.fd)?;
                        let response = node::StartBootstrapServerResponse::new();
                        let response_bytes = response.write_to_bytes()?;
                        Ok(response_bytes)
                    }
                    node::SetupRequest_oneof_req::epidAttestationRequest(req) => {
                        enclave_api::EnclaveApi::perform_epid_attestation(evm_enclave.geteid(), req.hostname, req.fd)?;
                        let response = node::EPIDAttestationResponse::new();
                        let response_bytes = response.write_to_bytes()?;
                        Ok(response_bytes)
                    }
                    node::SetupRequest_oneof_req::dcapAttestationRequest(req) => {
                        enclave_api::EnclaveApi::perform_dcap_attestation(evm_enclave.geteid(), req.hostname, req.fd)?;
                        let response = node::DCAPAttestationResponse::new();
                        let response_bytes = response.write_to_bytes()?;
                        Ok(response_bytes)
                    }
                    node::SetupRequest_oneof_req::isInitialized(_) => {
                        let is_initialized = enclave_api::EnclaveApi::is_enclave_initialized(evm_enclave.geteid())?;
                        let mut response = node::IsInitializedResponse::new();
                        response.isInitialized = is_initialized;
                        let response_bytes = response.write_to_bytes()?;
                        Ok(response_bytes)
                    }
                }
            }
            None => Err(Error::protobuf_decode("Request unwrapping failed")),
        };

        result
    })
    .unwrap_or_else(|_| Err(Error::panic()));

    let data = handle_c_error_default(r, error_msg);
    UnmanagedVector::new(Some(data))
}

#[no_mangle]
pub extern "C" fn make_pb_request(
    querier: GoQuerier,
    request: ByteSliceView,
    error_msg: Option<&mut UnmanagedVector>,
) -> UnmanagedVector {
    let r = catch_unwind(|| {
        // Check if request is correct
        let req_bytes = request
            .read()
            .ok_or_else(|| Error::unset_arg(PB_REQUEST_ARG))?;

        let enclave_access_token = ENCLAVE_DOORBELL
            .get_access(1) // This can never be recursive
            .ok_or(sgx_status_t::SGX_ERROR_BUSY)?;
        let evm_enclave = (*enclave_access_token)?;

        enclave_api::EnclaveApi::handle_evm_request(evm_enclave.geteid(), req_bytes, querier)
    }).unwrap_or_else(|_| Err(Error::panic()));

    let data = handle_c_error_default(r, error_msg);
    UnmanagedVector::new(Some(data))
}