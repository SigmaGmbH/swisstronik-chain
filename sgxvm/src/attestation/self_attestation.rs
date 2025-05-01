use sgx_types::*;

use crate::attestation::{
    tls::helpers,
    cert::verify_dcap_cert
};

#[cfg(feature = "hardware_mode")]
pub fn self_attest(
    qe_target_info: &sgx_target_info_t,
    quote_size: u32,
) -> SgxResult<()> {
    let (_, cert_der) = helpers::create_tls_cert_and_keys(qe_target_info, quote_size)?;

    // Verify quote
    match verify_dcap_cert(&cert_der) {
        Ok(_) => {
            #[cfg(feature = "mainnet")]
            println!("Your node is ready to be connected to mainnet");

            #[cfg(not(feature = "mainnet"))]
            println!("Your node is ready to be connected to testnet");

            Ok(())
        }
        Err (error) => {
            println!("[Enclave] Cannot verify DCAP cert. Reason: {:?}", error);
            Err(sgx_status_t::SGX_ERROR_UNEXPECTED)
        }
    }
}

#[cfg(not(feature = "hardware_mode"))]
pub fn self_attest(
    _: &sgx_target_info_t,
    _: u32,
) -> SgxResult<()> {
    println!("self_attest disabled in Software Mode");
    Err(sgx_status_t::SGX_ERROR_UNEXPECTED)
}