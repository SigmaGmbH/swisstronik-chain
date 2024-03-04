#![no_std]
#![feature(slice_as_chunks)]

#[macro_use]
extern crate sgx_tstd as std;
extern crate rustls;
extern crate sgx_tse;

extern crate sgx_types;
use sgx_tse::*;
use sgx_types::*;

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
    let hostname = slice::from_raw_parts(hostname, data_len);
    let hostname = match String::from_utf8(hostname.to_vec()) {
        Ok(hostname) => hostname,
        Err(err) => {
            println!("[Enclave] Cannot decode hostname. Reason: {:?}", err);
            return sgx_status_t::SGX_ERROR_UNEXPECTED;
        }
    };

    match attestation::tls::perform_master_key_request(
        hostname,
        socket_fd,
        Some(qe_target_info),
        Some(quote_size),
    ) {
        Ok(_) => sgx_status_t::SGX_SUCCESS,
        Err(err) => err
    }
}

#[no_mangle]
/// Handles incoming request for sharing master key with new node
pub unsafe extern "C" fn ecall_share_seed(socket_fd: c_int) -> sgx_status_t {
    let is_dcap = false;
    let res = match is_dcap {
        false => attestation::tls::perform_master_key_provisioning(socket_fd, None, None),
        true => {
            println!("[Enclave] DCAP master key provisioning is not supported yet");
            return sgx_status_t::SGX_ERROR_UNEXPECTED;
        }
    };

    match res {
        Ok(_) => sgx_status_t::SGX_SUCCESS,
        Err(err) => err
    }
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

    match attestation::tls::perform_master_key_request(hostname, socket_fd, None, None) {
        Ok(_) => sgx_status_t::SGX_SUCCESS,
        Err(err) => err
    }
}
