#![no_std]
#![feature(slice_as_chunks)]

#[macro_use]
extern crate sgx_tstd as std;
extern crate rustls;

extern crate sgx_types;
use sgx_types::sgx_status_t;

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

// TODO: move all ECALLs to lib.rs

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
    let request_slice = unsafe { slice::from_raw_parts(request_data, len) };

    let ffi_request = match protobuf::parse_from_bytes::<FFIRequest>(request_slice) {
        Ok(ffi_request) => ffi_request,
        Err(err) => {
            println!("Got error during protobuf decoding: {:?}", err);
            return AllocationWithResult::default();
        }
    };

    match ffi_request.req {
        Some(req) => {
            match req {
                FFIRequest_oneof_req::callRequest(data) => {
                    handlers::handle_evm_call_request(querier, data)
                },
                FFIRequest_oneof_req::createRequest(data) => {
                    handlers::handle_evm_create_request(querier, data)
                },
                FFIRequest_oneof_req::publicKeyRequest(_) => {
                    handlers::handle_public_key_request()
                }
            }
        }
        None => {
            println!("Got empty request during protobuf decoding");
            AllocationWithResult::default()
        }
    }
}
