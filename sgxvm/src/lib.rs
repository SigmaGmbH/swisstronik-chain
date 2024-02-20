#![no_std]
#![feature(slice_as_chunks)]

#[macro_use]
extern crate sgx_tstd as std;
extern crate rustls;

extern crate sgx_types;
use sgx_types::*;

use std::slice;

use crate::protobuf_generated::ffi::{FFIRequest, FFIRequest_oneof_req};
use crate::querier::GoQuerier;
use crate::types::{Allocation, AllocationWithResult};

mod backend;
mod coder;
mod error;
mod memory;
mod ocall;
mod protobuf_generated;
mod querier;
mod storage;
mod encryption;
mod attestation;
mod key_manager;
mod handlers;
mod types;
mod precompiles;

#[no_mangle]
/// Checks if there is already sealed master key
pub unsafe extern "C" fn ecall_is_initialized() -> i32 {
    if let Err(err) = key_manager::KeyManager::unseal() {
        println!("[Enclave] Cannot restore master key. Reason: {:?}", err.as_str());
        return false as i32
    }
    true as i32
} 

#[no_mangle]
/// Allocates provided data inside Intel SGX Enclave and returns 
/// pointer to allocated data and data length.
pub extern "C" fn ecall_allocate(
    data: *const u8,
    len: usize,
) -> crate::types::Allocation {
    let slice = unsafe { slice::from_raw_parts(data, len) };
    let mut vector_copy = slice.to_vec();

    let ptr = vector_copy.as_mut_ptr();
    let size = vector_copy.len();
    std::mem::forget(vector_copy);

    Allocation { result_ptr: ptr, result_size: size }
}

#[no_mangle]
/// Performes self attestation and outputs if system was configured
/// properly and node can pass Remote Attestation.
pub extern "C" fn ecall_status() -> sgx_status_t {
    attestation::self_attestation::self_attest()
}

#[no_mangle]
/// Handles incoming protobuf-encoded request
pub extern "C" fn handle_request(
    querier: *mut GoQuerier,
    request_data: *const u8,
    len: usize,
) -> AllocationWithResult {
    handlers::handle_protobuf_request_inner(querier, request_data, len)
}

#[no_mangle]
pub unsafe extern "C" fn ecall_dcap_attestation(
    hostname: *const u8,
    data_len: usize,
    socket_fd: c_int,
    qe_target_info: &sgx_target_info_t,
	quote_size: u32,
) -> sgx_status_t {
    attestation::dcap::perform_dcap_attestation(
        hostname,
        data_len,
        socket_fd,
        qe_target_info,
        quote_size,
    )
}
