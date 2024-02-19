use sgx_types::*;

pub unsafe fn get_quote_size() -> SgxResult<u32> {
    println!("[Enclave Wrapper]: get_quote_size");
    let mut quote_size: u32 = 0;
    let qe3_result = sgx_qe_get_quote_size(&mut quote_size as *mut _);

    match qe3_result {
        sgx_quote3_error_t::SGX_QL_SUCCESS => Ok(quote_size),
        _ => {
            println!("Cannot obtain quote size. Status code: {:?}", qe3_result);
            SgxResult::Err(sgx_status_t::SGX_ERROR_UNEXPECTED)
        }
    }
}

pub unsafe fn get_target_info() -> SgxResult<sgx_target_info_t> {
    println!("[Enclave Wrapper]: get_target_info");
    let mut target_info = sgx_target_info_t::default();
    let qe3_result = sgx_qe_get_target_info(&mut target_info as *mut _);
    match qe3_result {
        sgx_quote3_error_t::SGX_QL_SUCCESS => Ok(target_info),
        _ => {
            println!("Cannot obtain target info. Status code: {:?}", qe3_result);
            SgxResult::Err(sgx_status_t::SGX_ERROR_UNEXPECTED)
        }
    }
}

pub unsafe fn get_ecdsa_quote(report: sgx_report_t, quote_size: u32, p_quote: *mut u8) -> SgxResult<()> {
    println!("[Enclave Wrapper]: get_ecdsa_quote");
    let qe3_result = sgx_qe_get_quote(&report, quote_size, p_quote);
    match qe3_result {
        sgx_quote3_error_t::SGX_QL_SUCCESS => Ok(()),
        _ => {
            println!("Cannot get ecdsa quote. Status code: {:?}", qe3_result);
            SgxResult::Err(sgx_status_t::SGX_ERROR_UNEXPECTED)
        }
    }
}
