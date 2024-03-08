use sgx_types::*;
use crate::errors::Error;

/// Returns target info from Quoting Enclave (QE)
pub fn get_qe_target_info() -> Result<sgx_target_info_t, Error> {
    let mut qe_target_info = sgx_target_info_t::default();
    let qe3_ret = unsafe { sgx_qe_get_target_info(&mut qe_target_info) };
    if qe3_ret != sgx_quote3_error_t::SGX_QL_SUCCESS {
        println!("[Enclave Wrapper] sgx_qe_get_target_info failed. Status code: {:?}", qe3_ret);
        return Err(Error::enclave_error("sgx_qe_get_target_info failed"));
    }

    Ok(qe_target_info)
}

/// Returns size of buffer to allocate for the quote
pub fn get_quote_size() -> Result<u32, Error> {
    let mut quote_size = 0u32;
    let qe3_ret = unsafe { sgx_qe_get_quote_size(&mut quote_size) };
    if qe3_ret != sgx_quote3_error_t::SGX_QL_SUCCESS {
        println!("[Enclave Wrapper] sgx_qe_get_quote_size failed. Status code: {:?}", qe3_ret);
        return Err(Error::enclave_error("sgx_qe_get_quote_size failed"));
    }

    Ok(quote_size)
}

/// Returns DCAP quote from QE
pub fn get_qe_quote(report: sgx_report_t, quote_size: u32, p_quote: *mut u8) -> SgxResult<()> {
    println!("[Enclave Wrapper]: get_qe_quote");
    match unsafe {sgx_qe_get_quote(&report, quote_size, p_quote)} {
        sgx_quote3_error_t::SGX_QL_SUCCESS => Ok(()),
        err => {
            println!("Cannot get quote from QE. Status code: {:?}", err);
            SgxResult::Err(sgx_status_t::SGX_ERROR_UNEXPECTED)
        }
    }
}