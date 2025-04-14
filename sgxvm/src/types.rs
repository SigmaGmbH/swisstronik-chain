use std::vec::Vec;
use std::boxed::Box;
use sgx_types::*;

// Struct for allocated buffer outside SGX Enclave
#[repr(C)]
#[allow(dead_code)]
pub struct AllocatedBuffer {
    pub ptr: *mut u8,
}

/// Recovers boxed value from pointer
#[allow(dead_code)]
pub unsafe fn recover_buffer(buf: AllocatedBuffer) -> Option<Vec<u8>> {
    if buf.ptr.is_null() {
        return None;
    }
    let boxed_vector = Box::from_raw(buf.ptr as *mut Vec<u8>);
    Some(*boxed_vector)
}

#[repr(C)]
pub struct AllocationWithResult {
    pub result_ptr: *mut u8,
    pub result_len: usize,
    pub status: sgx_status_t
}

impl Default for AllocationWithResult {
    fn default() -> Self {
        AllocationWithResult {
            result_ptr: std::ptr::null_mut(),
            result_len: 0,
            status: sgx_status_t::SGX_ERROR_UNEXPECTED,
        }
    }
}

#[repr(C)]
pub struct Allocation {
    pub result_ptr: *mut u8,
    pub result_size: usize,
}