/// This file contains signatures of `OCALL` functions
use crate::{querier::GoQuerier, Allocation, AllocationWithResult};
use sgx_types::sgx_status_t;
use sgx_types::*;
use std::vec::Vec;

extern "C" {
    pub fn ocall_query_raw(
        ret_val: *mut AllocationWithResult,
        querier: *mut GoQuerier,
        request: *const u8,
        len: usize,
    ) -> sgx_status_t;

    pub fn ocall_allocate(ret_val: *mut Allocation, data: *const u8, len: usize) -> sgx_status_t;

    pub fn ocall_sgx_init_quote(
        ret_val: *mut sgx_status_t,
        ret_ti: *mut sgx_target_info_t,
        ret_gid: *mut sgx_epid_group_id_t,
    ) -> sgx_status_t;

    pub fn ocall_get_ias_socket(ret_val: *mut sgx_status_t, ret_fd: *mut i32) -> sgx_status_t;

    pub fn ocall_get_quote(
        ret_val: *mut sgx_status_t,
        p_sigrl: *const u8,
        sigrl_len: u32,
        p_report: *const sgx_report_t,
        quote_type: sgx_quote_sign_type_t,
        p_spid: *const sgx_spid_t,
        p_nonce: *const sgx_quote_nonce_t,
        p_qe_report: *mut sgx_report_t,
        p_quote: *mut u8,
        maxlen: u32,
        p_quote_len: *mut u32,
    ) -> sgx_status_t;
}

pub fn make_request(querier: *mut GoQuerier, request: Vec<u8>) -> Option<Vec<u8>> {
    let mut allocation = std::mem::MaybeUninit::<AllocationWithResult>::uninit();

    let result = unsafe {
        ocall_query_raw(
            allocation.as_mut_ptr(),
            querier,
            request.as_ptr(),
            request.len(),
        )
    };

    match result {
        sgx_status_t::SGX_SUCCESS => {
            let allocation = unsafe { allocation.assume_init() };
            let result_vec = unsafe {
                Vec::from_raw_parts(
                    allocation.result_ptr,
                    allocation.result_len,
                    allocation.result_len,
                )
            };

            return Some(result_vec);
        }
        _ => {
            println!("make_request failed: Reason: {:?}", result.as_str());
            return None;
        }
    };
}
