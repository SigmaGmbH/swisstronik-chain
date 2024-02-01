use std::vec::Vec;
use sgx_types::sgx_status_t;
use protobuf::Message;

use crate::{AllocationWithResult, Allocation};
use crate::ocall;
use crate::key_manager::KeyManager;
use crate::protobuf_generated::ffi::{
    NodePublicKeyResponse,
    SGXVMCallRequest, 
    SGXVMCreateRequest,
};
use crate::GoQuerier;

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
        Err(_) => {
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
    
    allocate_inner(encoded_response)
}

/// Handles incoming request for calling contract or transferring value
/// * querier – GoQuerier which is used to interact with Go (Cosmos) from SGX Enclave
/// * data – EVM call data (destination, value, etc.)
pub fn handle_evm_call_request(querier: *mut GoQuerier, data: SGXVMCallRequest) -> AllocationWithResult {
    let res = tx::handle_call_request_inner(querier, data);
    tx::convert_and_allocate_transaction_result(res)
}

/// Handles incoming request for creation of a new contract
/// * querier – GoQuerier which is used to interact with Go (Cosmos) from SGX Enclave
/// * data – EVM call data (value, tx.data, etc.)
pub fn handle_evm_create_request(querier: *mut GoQuerier, data: SGXVMCreateRequest) -> AllocationWithResult {
    let res = tx::handle_create_request_inner(querier, data);
    tx::convert_and_allocate_transaction_result(res)
}