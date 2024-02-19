use sgx_types::*;

struct Attesteer;

impl Attesteer {
    pub unsafe fn get_quote_size(&self) -> SgxResult<u32> {
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

    pub unsafe fn get_target_info(&self) -> SgxResult<sgx_target_info_t> {
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

    pub unsafe fn get_ecdsa_quote(&self, report: sgx_report_t, quote_size: u32) -> SgxResult<Vec<u8>> {
        println!("[Enclave Wrapper]: get_ecdsa_quote");
        let mut quote: Vec<u8> = vec![0; quote_size as usize];
        let qe3_result = sgx_qe_get_quote(&report,  quote_size, quote.as_mut_ptr() as _);
        match qe3_result {
            sgx_quote3_error_t::SGX_QL_SUCCESS => Ok(quote),
            _ => {
                println!("Cannot get ecdsa quote. Status code: {:?}", qe3_result);
                SgxResult::Err(sgx_status_t::SGX_ERROR_UNEXPECTED)
            }
        }
    }
}

pub(crate) unsafe fn try_dcap() -> SgxError {
    println!("TRYING DCAP 2");
    let attesteer = Attesteer{};

    let _target_info = attesteer.get_target_info()?;
    println!("target info generated");
    let quote_size = attesteer.get_quote_size()?;

    println!("QUOTE SIZE: {:?}", quote_size);
    Ok(())
}