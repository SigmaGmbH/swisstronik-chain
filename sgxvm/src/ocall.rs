/// This file contains signatures of `OCALL` functions
use crate::{querier::GoQuerier, Allocation, AllocationWithResult};
use sgx_types::sgx_status_t;
use sgx_types::*;

extern "C" {
    pub fn ocall_query_raw(
        _ret_val: *mut AllocationWithResult,
        _querier: *mut GoQuerier,
        _request: *const u8,
        _len: usize,
    ) -> sgx_status_t;

    pub fn ocall_allocate(
        _ret_val: *mut Allocation,
        _data: *const u8,
        _len: usize) -> sgx_status_t;

    pub fn ocall_sgx_init_quote(
        _ret_val: *mut sgx_status_t,
        _ret_ti: *mut sgx_target_info_t,
        _ret_gid: *mut sgx_epid_group_id_t,
    ) -> sgx_status_t;

    pub fn ocall_get_ias_socket(
        _ret_val: *mut sgx_status_t,
        _ret_fd: *mut i32,
    ) -> sgx_status_t;

    pub fn ocall_get_quote(
        _ret_val: *mut sgx_status_t,
        _p_sigrl: *const u8,
        _sigrl_len: u32,
        _p_report: *const sgx_report_t,
        _quote_type: sgx_quote_sign_type_t,
        _p_spid: *const sgx_spid_t,
        _p_nonce: *const sgx_quote_nonce_t,
        _p_qe_report: *mut sgx_report_t,
        _p_quote: *mut u8,
        _maxlen: u32,
        _p_quote_len: *mut u32,
    ) -> sgx_status_t;

    pub fn ocall_get_ecdsa_quote(
		_ret_val: *mut sgx_status_t,
		_p_report: *const sgx_report_t,
		_p_quote: *mut u8,
		_quote_size: u32,
	) -> sgx_status_t;

    pub fn ocall_get_qve_report(
		_ret_val: *mut sgx_status_t,
		_p_quote: *const u8,
		_quote_len: u32,
		_timestamp: i64,
		_p_collateral_expiration_status: *mut u32,
		_p_quote_verification_result: *mut sgx_ql_qv_result_t,
		_p_qve_report_info: *mut sgx_ql_qe_report_info_t,
		_p_supplemental_data: *mut u8,
		_supplemental_data_size: u32,
        _p_collateral: *const u8,
        _collateral_len: u32,
    ) -> sgx_status_t;

    pub fn ocall_get_supplemental_data_size(
        _ret_val: *mut sgx_status_t,
        _data_size: *mut u32,
    ) -> sgx_status_t;

    pub fn ocall_get_quote_ecdsa_collateral(
        _ret_val: *mut sgx_status_t,
        _p_quote: *const u8,
        _n_quote: u32,
        _p_col: *mut u8,
        _n_col: u32,
        _p_col_out: *mut u32,
    ) -> sgx_status_t;
}
