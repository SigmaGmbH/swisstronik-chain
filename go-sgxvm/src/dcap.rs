use sgx_types::*;

pub unsafe fn get_quote_size() -> SgxResult<u32> {
    println!("[Enclave Wrapper]: get_quote_size");
    let mut quote_size: u32 = 0;

    match sgx_qe_get_quote_size(&mut quote_size as *mut _) {
        sgx_quote3_error_t::SGX_QL_SUCCESS => Ok(quote_size),
        err => {
            println!("Cannot obtain quote size. Status code: {:?}", err);
            SgxResult::Err(sgx_status_t::SGX_ERROR_UNEXPECTED)
        }
    }
}

pub unsafe fn get_target_info() -> SgxResult<sgx_target_info_t> {
    println!("[Enclave Wrapper]: get_target_info");
    let mut target_info = sgx_target_info_t::default();
    match sgx_qe_get_target_info(&mut target_info as *mut _) {
        sgx_quote3_error_t::SGX_QL_SUCCESS => Ok(target_info),
        err => {
            println!("Cannot obtain target info. Status code: {:?}", err);
            SgxResult::Err(sgx_status_t::SGX_ERROR_UNEXPECTED)
        }
    }
}

// pub unsafe fn get_ecdsa_quote(report: sgx_report_t, quote_size: u32, p_quote: *mut u8) -> SgxResult<()> {
//     println!("[Enclave Wrapper]: get_ecdsa_quote");
//     match sgx_qe_get_quote(&report, quote_size, p_quote) {
//         sgx_quote3_error_t::SGX_QL_SUCCESS => Ok(()),
//         err => {
//             println!("Cannot get ecdsa quote. Status code: {:?}", err);
//             SgxResult::Err(sgx_status_t::SGX_ERROR_UNEXPECTED)
//         }
//     }
// }

pub unsafe fn set_qve_loading_policy(policy: sgx_ql_request_policy_t) -> SgxResult<()> {
    println!("[Enclave Wrapper]: set_qve_loading_policy");
    match sgx_qv_set_enclave_load_policy(policy) {
        sgx_quote3_error_t::SGX_QL_SUCCESS => Ok(()),
        err => {
            println!("Cannot set QvE loading policy. Status code: {:?}", err);
            SgxResult::Err(sgx_status_t::SGX_ERROR_UNEXPECTED)
        }
    }
}
