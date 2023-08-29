use crate::{AllocationWithResult, Allocation};
use std::vec::Vec;
use sgx_types::sgx_status_t;
use crate::ocall;

pub mod tx;
pub mod node;

/// Allocates provided data outside of enclave
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