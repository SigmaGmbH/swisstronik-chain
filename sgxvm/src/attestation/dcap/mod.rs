use sgx_tcrypto::SgxEccHandle;
use sgx_tse::rsgx_create_report;
use sgx_types::*;
use std::untrusted::time::SystemTimeEx;
use std::{time::SystemTime, vec::Vec};

use std::str::FromStr;
use yasna::models::ObjectIdentifier;

use crate::{attestation::cert, ocall};

pub fn perform_dcap_attestation(
    hostname: *const u8,
    data_len: usize,
    socket_fd: c_int,
    qe_target_info: &sgx_target_info_t,
	quote_size: u32,
) -> sgx_status_t {
    println!("[Enclave] Getting QE quote");
    let ecc_handle = SgxEccHandle::new();
    let _result = ecc_handle.open();
    let (prv_k, pub_k) = match ecc_handle.create_key_pair() {
        Ok((prv_k, pub_k)) => (prv_k, pub_k),
        Err(err) => {
            println!("[Enclave] Cannot create keypair for DCAP Cert. Status code: {:?}", err);
            return sgx_status_t::SGX_ERROR_UNEXPECTED;
        }
    };
    let qe_quote = match get_qe_quote(&pub_k, qe_target_info, quote_size) {
        Ok(qe_quote) => qe_quote,
        Err(err) => {
            println!("[Enclave] Cannot obtain qe quote from PCCS. Status code: {:?}", err);
            return sgx_status_t::SGX_ERROR_UNEXPECTED;
        }
    };

    println!("[Enclave] Verify quote");
    match verify_dcap_quote(qe_quote) {
        Ok(_) => {
            println!("[Enclave] Quote verified");
        },
        Err(err) => {
            println!("[Enclave] Cannot verify quote. Status code: {:?}", err);
            return err
        }
    }

    sgx_status_t::SGX_SUCCESS
}

fn get_qe_quote(
    pub_k: &sgx_ec256_public_t,
    qe_target_info: &sgx_target_info_t,
    quote_size: u32,
) -> SgxResult<Vec<u8>> {
    let mut report_data: sgx_report_data_t = sgx_report_data_t::default();
    
    // Copy public key to report data
    let mut pub_k_gx = pub_k.gx.clone();
    pub_k_gx.reverse();
    let mut pub_k_gy = pub_k.gy.clone();
    pub_k_gy.reverse();
    report_data.d[..32].clone_from_slice(&pub_k_gx);
    report_data.d[32..].clone_from_slice(&pub_k_gy);

    // Prepare report
    let report = match rsgx_create_report(qe_target_info, &report_data) {
        Ok(report) => report,
        Err(err) => {
            println!("[Enclave] Call to rsgx_create_report failed. Status code: {:?}", err);
            return Err(err);
        }
    };

    // Get quote from PCCS
    let mut ret_val = sgx_status_t::SGX_SUCCESS;
    let mut quote_buf = vec![0u8; quote_size as usize]; 
    let res = unsafe {
        ocall::ocall_get_ecdsa_quote(
            &mut ret_val as *mut sgx_status_t, 
            &report as *const sgx_report_t, 
            quote_buf.as_mut_ptr(),
            quote_size
        )
    };

    let qe_quote: Vec<u8> = match (res, ret_val) {
        (sgx_status_t::SGX_SUCCESS, sgx_status_t::SGX_SUCCESS) => Vec::from(&quote_buf[..quote_size as usize]),
        (_, _) => {
            println!(
                "[Enclave] ocall_get_ecdsa_quote failed. Status codes: res: {:?}, ret_val: {:?}",
                res, ret_val
            );
            return Err(sgx_status_t::SGX_ERROR_UNEXPECTED);
        }
    };

    // Perform additional check if quote was tampered
    let p_quote3: *const sgx_quote3_t = qe_quote.as_ptr() as *const sgx_quote3_t;
    let quote3: sgx_quote3_t = unsafe { *p_quote3 };

    if quote3.report_body.mr_enclave.m != report.body.mr_enclave.m {
        println!("MRENCLAVE in quote and report are different. Quote was tampered!");
        return Err(sgx_status_t::SGX_ERROR_UNEXPECTED);
    }

    Ok(qe_quote)
}

fn verify_dcap_quote(quote_vec: Vec<u8>) -> SgxResult<()> {
    // Reconstruct quote
    let p_quote3: *const sgx_quote3_t = quote_vec.as_ptr() as *const sgx_quote3_t;
    let quote: sgx_quote3_t = unsafe { *p_quote3 };

    // Prepare data for enclave
    let mut self_target_info: sgx_target_info_t = unsafe { std::mem::zeroed() };
    let quote_collateral: sgx_ql_qve_collateral_t = unsafe { std::mem::zeroed() };
    let mut report_info: sgx_ql_qe_report_info_t = unsafe { std::mem::zeroed() };
    let supplemental_data_size = std::mem::size_of::<sgx_ql_qv_supplemental_t>() as u32;
    let mut supplemental_data = vec![0u8; supplemental_data_size as usize];

    // Generate target_info for enclave
    let ret_val = unsafe { sgx_self_target(&mut self_target_info as *mut sgx_target_info_t) };
    if ret_val != sgx_status_t::SGX_SUCCESS {
        println!("Call to sgx_self_target failed. Status code: {:?}", ret_val);
        return Err(sgx_status_t::SGX_ERROR_UNEXPECTED)
    }

    // Generate random nonce to ensure that quote was not tampered
    let mut nonce = vec![0u8; 16];
    let rev_val = unsafe { sgx_read_rand(nonce.as_mut_ptr(), nonce.len()) };
    if rev_val != sgx_status_t::SGX_SUCCESS {
        println!("Call to sgx_read_rand failed. Status code: {:?}", ret_val);
        return Err(sgx_status_t::SGX_ERROR_UNEXPECTED) 
    }

    // Prepare current timestamp
    let timestamp = match SystemTime::now().duration_since(SystemTime::UNIX_EPOCH) {
        Ok(timestamp) => timestamp,
        Err(err) => {
            println!("Cannot get current timestamp. Reason: {:?}", err);
            return Err(sgx_status_t::SGX_ERROR_UNEXPECTED)  
        }
    };
    let timestamp_secs: i64 = match timestamp.as_secs().try_into() {
        Ok(secs) => secs,
        Err(err) => {
            println!("Cannot convert current timestamp to i64. Reason: {:?}", err);
            return Err(sgx_status_t::SGX_ERROR_UNEXPECTED);
        }
    };

    // Fill report info
    report_info.nonce.rand.copy_from_slice(&nonce);
    report_info.app_enclave_target_info = self_target_info;

    // Send OCALL to QvE
    let mut ret_val = sgx_status_t::SGX_SUCCESS;
    let mut qve_report_info: sgx_ql_qe_report_info_t = report_info;
    let mut quote_verification_result = sgx_ql_qv_result_t::SGX_QL_QV_RESULT_UNSPECIFIED;
    let mut collateral_expiration_status = 1u32;

    let res = unsafe {
        ocall::ocall_get_qve_report(
            &mut ret_val as *mut sgx_status_t, 
            quote_vec.as_ptr(), 
            quote_vec.len() as u32, 
            timestamp_secs, 
            &quote_collateral as *const sgx_ql_qve_collateral_t, 
            &mut collateral_expiration_status as *mut u32, 
            &mut quote_verification_result as *mut sgx_ql_qv_result_t, 
            &mut qve_report_info as *mut sgx_ql_qe_report_info_t, 
            supplemental_data.as_mut_ptr(), 
            supplemental_data_size,
        )
    };
    match (res, ret_val) {
        (sgx_status_t::SGX_SUCCESS, sgx_status_t::SGX_SUCCESS) => (),
        (_, _) => {
            println!(
                "[Enclave] ocall_get_qve_report failed. Status codes: res: {:?}, ret_val: {:?}",
                res, ret_val
            );
            return Err(sgx_status_t::SGX_ERROR_UNEXPECTED);
        }
    };

    Ok(())
}