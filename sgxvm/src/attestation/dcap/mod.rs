use sgx_tcrypto::SgxEccHandle;
use sgx_tse::rsgx_create_report;
use sgx_types::*;
use std::vec::Vec;

use crate::{attestation::cert, ocall};

// Intel's PCS signing root certificate.
const PCS_TRUST_ROOT_CERT: &str = r#"-----BEGIN CERTIFICATE-----
MIICjzCCAjSgAwIBAgIUImUM1lqdNInzg7SVUr9QGzknBqwwCgYIKoZIzj0EAwIw
aDEaMBgGA1UEAwwRSW50ZWwgU0dYIFJvb3QgQ0ExGjAYBgNVBAoMEUludGVsIENv
cnBvcmF0aW9uMRQwEgYDVQQHDAtTYW50YSBDbGFyYTELMAkGA1UECAwCQ0ExCzAJ
BgNVBAYTAlVTMB4XDTE4MDUyMTEwNDUxMFoXDTQ5MTIzMTIzNTk1OVowaDEaMBgG
A1UEAwwRSW50ZWwgU0dYIFJvb3QgQ0ExGjAYBgNVBAoMEUludGVsIENvcnBvcmF0
aW9uMRQwEgYDVQQHDAtTYW50YSBDbGFyYTELMAkGA1UECAwCQ0ExCzAJBgNVBAYT
AlVTMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEC6nEwMDIYZOj/iPWsCzaEKi7
1OiOSLRFhWGjbnBVJfVnkY4u3IjkDYYL0MxO4mqsyYjlBalTVYxFP2sJBK5zlKOB
uzCBuDAfBgNVHSMEGDAWgBQiZQzWWp00ifODtJVSv1AbOScGrDBSBgNVHR8ESzBJ
MEegRaBDhkFodHRwczovL2NlcnRpZmljYXRlcy50cnVzdGVkc2VydmljZXMuaW50
ZWwuY29tL0ludGVsU0dYUm9vdENBLmRlcjAdBgNVHQ4EFgQUImUM1lqdNInzg7SV
Ur9QGzknBqwwDgYDVR0PAQH/BAQDAgEGMBIGA1UdEwEB/wQIMAYBAf8CAQEwCgYI
KoZIzj0EAwIDSQAwRgIhAOW/5QkR+S9CiSDcNoowLuPRLsWGf/Yi7GSX94BgwTwg
AiEA4J0lrHoMs+Xo5o/sX6O9QWxHRAvZUGOdRQ7cvqRXaqI=
-----END CERTIFICATE-----"#;

#[no_mangle]
pub unsafe extern "C" fn ecall_dcap_attestation(
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

    println!("[Enclave] Creating ECC certificate");
    let qe_quote_base64 = base64::encode(qe_quote.as_slice());
    let (key_der, cert_der) = match cert::gen_ecc_cert(
        qe_quote_base64, &prv_k, &pub_k, &ecc_handle
    ) {
        Ok((key_der, cert_der)) => (key_der, cert_der),
        Err(err) => {
            println!("[Enclave] Cannot generate ECC cert. Status code: {:?}", err);
            return sgx_status_t::SGX_ERROR_UNEXPECTED;
        }
    };


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