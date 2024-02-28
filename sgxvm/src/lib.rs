#![no_std]
#![feature(slice_as_chunks)]
#![feature(core_intrinsics)]

#[macro_use]
extern crate sgx_tstd as std;
extern crate rustls;
extern crate sgx_tse;

extern crate sgx_types;
use sgx_types::*;
use sgx_tse::*;

use std::slice;
use std::string::String;

use crate::querier::GoQuerier;
use crate::types::{Allocation, AllocationWithResult};

mod attestation;
mod backend;
mod coder;
mod encryption;
mod error;
mod handlers;
mod key_manager;
mod memory;
mod ocall;
mod precompiles;
mod protobuf_generated;
mod querier;
mod storage;
mod types;

#[no_mangle]
/// Checks if there is already sealed master key
pub unsafe extern "C" fn ecall_is_initialized() -> i32 {
    if let Err(err) = key_manager::KeyManager::unseal() {
        println!(
            "[Enclave] Cannot restore master key. Reason: {:?}",
            err.as_str()
        );
        return false as i32;
    }
    true as i32
}

#[no_mangle]
/// Allocates provided data inside Intel SGX Enclave and returns
/// pointer to allocated data and data length.
pub extern "C" fn ecall_allocate(data: *const u8, len: usize) -> crate::types::Allocation {
    let slice = unsafe { slice::from_raw_parts(data, len) };
    let mut vector_copy = slice.to_vec();

    let ptr = vector_copy.as_mut_ptr();
    let size = vector_copy.len();
    std::mem::forget(vector_copy);

    Allocation {
        result_ptr: ptr,
        result_size: size,
    }
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
/// Handles incoming request for DCAP Remote Attestation
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

#[no_mangle]
/// Handles incoming request for sharing master key with new node
pub unsafe extern "C" fn ecall_share_seed(socket_fd: c_int) -> sgx_status_t {
    attestation::seed_server::share_seed_inner(socket_fd)
}

#[no_mangle]
/// Handles initialization of a new seed node by creating and sealing master key to seed file
/// If `reset_flag` was set to `true`, it will rewrite existing seed file
pub unsafe extern "C" fn ecall_init_master_key(reset_flag: i32) -> sgx_status_t {
    key_manager::init_master_key_inner(reset_flag)
}

#[no_mangle]
/// Handles incoming request for EPID Remote Attestation
pub unsafe extern "C" fn ecall_request_seed(
    hostname: *const u8,
    data_len: usize,
    socket_fd: c_int,
) -> sgx_status_t {
    let hostname = slice::from_raw_parts(hostname, data_len);
    let hostname = match String::from_utf8(hostname.to_vec()) {
        Ok(hostname) => hostname,
        Err(err) => {
            println!(
                "[Enclave] Seed Client. Cannot decode hostname. Reason: {:?}",
                err
            );
            return sgx_status_t::SGX_ERROR_UNEXPECTED;
        }
    };

    attestation::seed_client::request_seed_inner(hostname, socket_fd)
}

#[no_mangle]
pub unsafe extern "C" fn ecall_create_report(
    p_qe3_target: *const sgx_target_info_t,
    p_report: *mut sgx_report_t,
) -> sgx_status_t {    
    let report_data = sgx_report_data_t::default();
    unsafe { sgx_create_report(p_qe3_target, &report_data, p_report) }
}

#[no_mangle]
pub unsafe extern "C" fn ecall_get_target_info(
    target_info: *mut sgx_target_info_t,
) -> sgx_status_t {
    sgx_self_target(target_info)
}



// Fix https://github.com/apache/incubator-teaclave-sgx-sdk/issues/373 for debug mode
#[cfg(debug_assertions)]
#[no_mangle]
pub extern "C" fn __assert_fail(
    __assertion: *const u8,
    __file: *const u8,
    __line: u32,
    __function: *const u8,
) -> ! {
    use core::intrinsics::abort;
    abort()
}