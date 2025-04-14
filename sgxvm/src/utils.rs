use std::vec::Vec;
use crate::ocall;
use crate::types::{Allocation, AllocationWithResult};
use sgx_types::sgx_status_t;

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