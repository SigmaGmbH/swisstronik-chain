use std::vec::Vec;
use sgx_types::sgx_status_t;

use crate::{AllocationWithResult, Allocation};
use crate::ocall;
use crate::key_manager::KeyManager;
use crate::protobuf_generated::ffi::NodePublicKeyResponse;

pub mod tx;

/// Allocates provided data outside of enclave
/// * data - bytes to allocate outside of enclave
/// 
/// Returns allocation result with pointer to allocated memory, length of allocated data and status of allocation
pub fn allocate_inner(data: Vec<u8>) -> AllocationWithResult {
    let mut ocall_result = std::mem::MaybeUninit::<Allocation>::uninit();
    let sgx_result = unsafe { 
        ocall::ocall_allocate(
            ocall_result.as_mut_ptr(),
            data.as_ptr(),
            data.len()
        ) 
    };
    match sgx_result {
        sgx_status_t::SGX_SUCCESS => {
            let ocall_result = unsafe { ocall_result.assume_init() };
            AllocationWithResult {
                result_ptr: ocall_result.result_ptr,
                result_len: data.len(),
                status: sgx_status_t::SGX_SUCCESS
            }
        },
        _ => {
            println!("ocall_allocate failed: {:?}", sgx_result.as_str());
            AllocationWithResult::default()
        }
    }
}

/// Handles incoming request for node public key, which can be used
/// to derive shared encryption key to encrypt transaction data or 
/// decrypt node response
pub fn handle_public_key_request() -> AllocationWithResult {
    let key_manager = match KeyManager::unseal() {
        Ok(manager) => manager,
        Err(err) => {
            return AllocationWithResult::default()
        }
    };

    let public_key = key_manager.get_public_key();

    let mut response = NodePublicKeyResponse::new();
    response.set_publicKey(public_key);

    let encoded_response = match response.write_to_bytes() {
        Ok(res) => res,
        Err(err) => {
            println!("Cannot encode protobuf result. Reason: {:?}", err);
            return AllocationWithResult::default();
        }
    };
    
    super::allocate_inner(encoded_response)
}