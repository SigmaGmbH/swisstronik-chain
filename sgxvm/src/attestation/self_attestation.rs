use sgx_tcrypto::*;
use sgx_types::*;
use std::string::String;

use crate::attestation::{
    consts::QUOTE_SIGNATURE_TYPE,
    utils::create_attestation_report,
    cert::gen_ecc_cert,
};

#[cfg(feature = "hardware_mode")]
pub fn self_attest() -> sgx_status_t {
    use super::report::AttestationReport;

    let ecc_handle = SgxEccHandle::new();
    let _result = ecc_handle.open();
    let (prv_k, pub_k) = ecc_handle.create_key_pair().unwrap();

    let signed_report = match create_attestation_report(&pub_k, QUOTE_SIGNATURE_TYPE)
    {
        Ok(r) => r,
        Err(e) => {
            println!("[Enclave] Cannot create attestation report. Reason: {:?}", e.as_str());
            return sgx_status_t::SGX_ERROR_UNEXPECTED;
        }
    };

    let payload: String = match serde_json::to_string(&signed_report) {
        Ok(payload) => payload,
        Err(err) => {
            println!("Cannot serialize report. Reason: {:?}", err);
            return sgx_status_t::SGX_ERROR_UNEXPECTED;
        }
    };

    let (key_der, cert_der) = match gen_ecc_cert(payload, &prv_k, &pub_k, &ecc_handle)
    {
        Ok(result) => result,
        Err(err) => {
            println!("Error in gen_ecc_cert. Reason {:?}", err);
            return sgx_status_t::SGX_ERROR_UNEXPECTED;
        }
    };

    let report = match AttestationReport::from_cert(&cert_der) {
        Ok(report) => report,
        Err(err) => {
            println!("Cannot parse attestation report. Reason: {:?}", err);
            return sgx_status_t::SGX_ERROR_UNEXPECTED;
        }
    };    

    sgx_status_t::SGX_SUCCESS
}

#[cfg(not(feature = "hardware_mode"))]
pub fn self_attest() -> sgx_status_t {
    println!("[Enclave] You're using swisstronikd built in SOFTWARE mode. It cannot be used to connect to actual testnet / mainnet");    
    sgx_status_t::SGX_SUCCESS
}