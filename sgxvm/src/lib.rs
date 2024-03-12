#![no_std]
#![feature(slice_as_chunks)]

#[macro_use]
extern crate sgx_tstd as std;
extern crate rustls;
extern crate sgx_tse;
extern crate sgx_types;

use sgx_types::*;
use sgx_tcrypto::*;

use std::slice;
use std::string::String;

use crate::querier::GoQuerier;
use crate::types::{Allocation, AllocationWithResult};
use crate::attestation::dcap::get_qe_quote;

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
pub unsafe extern "C" fn ecall_request_master_key_dcap(
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
        true,
    ) {
        Ok(_) => sgx_status_t::SGX_SUCCESS,
        Err(err) => err,
    }
}

#[no_mangle]
/// Handles incoming request for sharing master key with new node using DCAP attestation
pub unsafe extern "C" fn ecall_attest_peer_dcap(
    socket_fd: c_int,
    qe_target_info: &sgx_target_info_t,
    quote_size: u32,
) -> sgx_status_t {
    match attestation::tls::perform_master_key_provisioning(socket_fd, Some(qe_target_info), Some(quote_size), true) {
        Ok(_) => sgx_status_t::SGX_SUCCESS,
        Err(err) => err,
    }
}

#[no_mangle]
/// Handles incoming request for sharing master key with new node using EPID attestation
pub unsafe extern "C" fn ecall_attest_peer_epid(socket_fd: c_int) -> sgx_status_t {
    match attestation::tls::perform_master_key_provisioning(socket_fd, None, None, false) {
        Ok(_) => sgx_status_t::SGX_SUCCESS,
        Err(err) => err,
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
pub unsafe extern "C" fn ecall_request_master_key_epid(
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

    match attestation::tls::perform_master_key_request(hostname, socket_fd, None, None, false) {
        Ok(_) => sgx_status_t::SGX_SUCCESS,
        Err(err) => err,
    }
}

#[no_mangle]
pub unsafe extern "C" fn ecall_dump_dcap_quote(
    qe_target_info: &sgx_target_info_t,
    quote_size: u32,
) -> AllocationWithResult {
    let ecc_handle = SgxEccHandle::new();
    let _ = ecc_handle.open();
    let (_, pub_k) = match ecc_handle.create_key_pair() {
        Ok(res) => res,
        Err(status_code) => {
            println!("[Enclave] Cannot create key pair using SgxEccHandle. Reason: {:?}", status_code);
            return AllocationWithResult::default();
        }
    };

    let qe_quote = match get_qe_quote(&pub_k, qe_target_info, quote_size) {
        Ok(quote) => quote,
        Err(status_code) => {
            println!("[Enclave] Cannot generate QE quote. Reason: {:?}", status_code);
            return AllocationWithResult::default();
        } 
    };

    let _ = ecc_handle.close();

    handlers::allocate_inner(qe_quote)
}

#[no_mangle]
pub unsafe extern "C" fn ecall_verify_dcap_quote(
    quote_ptr: *const u8,
    quote_len: u32,
) -> sgx_status_t {
    let slice = unsafe { slice::from_raw_parts(quote_ptr, quote_len as usize) };
    let quote_buf = slice.to_vec();

    match attestation::dcap::verify_dcap_quote(quote_buf.to_vec()) {
        Ok(_) => {
            println!("[Enclave] Quote verified");
            sgx_status_t::SGX_SUCCESS
        },
        Err(err) => {
            println!("[Enlcave] Quote verification failed. Status code: {:?}", err);
            err
        }
    }
}
