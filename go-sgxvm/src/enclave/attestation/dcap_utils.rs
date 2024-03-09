use std::io::Write;

use crate::enclave;
use crate::errors::Error;
use crate::types;
use sgx_types::*;

/// Returns target info from Quoting Enclave (QE)
pub fn get_qe_target_info() -> Result<sgx_target_info_t, Error> {
    let mut qe_target_info = sgx_target_info_t::default();
    let qe3_ret = unsafe { sgx_qe_get_target_info(&mut qe_target_info) };
    if qe3_ret != sgx_quote3_error_t::SGX_QL_SUCCESS {
        println!(
            "[Enclave Wrapper] sgx_qe_get_target_info failed. Status code: {:?}",
            qe3_ret
        );
        return Err(Error::enclave_error("sgx_qe_get_target_info failed"));
    }

    Ok(qe_target_info)
}

/// Returns size of buffer to allocate for the quote
pub fn get_quote_size() -> Result<u32, Error> {
    let mut quote_size = 0u32;
    let qe3_ret = unsafe { sgx_qe_get_quote_size(&mut quote_size) };
    if qe3_ret != sgx_quote3_error_t::SGX_QL_SUCCESS {
        println!(
            "[Enclave Wrapper] sgx_qe_get_quote_size failed. Status code: {:?}",
            qe3_ret
        );
        return Err(Error::enclave_error("sgx_qe_get_quote_size failed"));
    }

    Ok(quote_size)
}

/// Returns DCAP quote from QE
pub fn get_qe_quote(report: sgx_report_t, quote_size: u32, p_quote: *mut u8) -> SgxResult<()> {
    println!("[Enclave Wrapper]: get_qe_quote");
    match unsafe { sgx_qe_get_quote(&report, quote_size, p_quote) } {
        sgx_quote3_error_t::SGX_QL_SUCCESS => Ok(()),
        err => {
            println!("Cannot get quote from QE. Status code: {:?}", err);
            SgxResult::Err(sgx_status_t::SGX_ERROR_UNEXPECTED)
        }
    }
}

/// Generates quote inside the enclave and writes it to the file
/// Since this function will be used only for test and dev purposes, 
/// we can ignore usages of `unwrap` or `expect`.
pub fn dump_dcap_quote(eid: sgx_enclave_id_t, filepath: &str) -> Result<(), Error> {
    let qe_target_info = get_qe_target_info()?;
    let quote_size = get_quote_size()?;
    let mut retval = std::mem::MaybeUninit::<types::AllocationWithResult>::uninit();

    let res = unsafe {
        enclave::ecall_dump_dcap_quote(
            eid,
            retval.as_mut_ptr(),
            &qe_target_info,
            quote_size,
        )
    };

    if res != sgx_status_t::SGX_SUCCESS {
        panic!("Call to `ecall_dump_dcap_quote` failed. Reason: {:?}", res);
    }

    let quote_res = unsafe { retval.assume_init() };
    if quote_res.status != sgx_status_t::SGX_SUCCESS {
        panic!("`ecall_dump_dcap_quote` returned error code: {:?}", quote_res.status);
    }

    let quote_vec = unsafe {
        Vec::from_raw_parts(quote_res.result_ptr, quote_res.result_size, quote_res.result_size)
    };

    let mut quote_file = std::fs::File::create(filepath)
        .expect("Cannot create file to write quote");

    quote_file.write_all(&quote_vec).expect("Cannot write quote to file");

    Ok(())
}
