use sgx_tcrypto::*;
use sgx_types::*;
use rustls::{self, ClientConfig};
use std::prelude::v1::*;
use std::sync::Arc;
use std::vec::Vec;

use crate::attestation::{
    cert::gen_ecc_cert,
    consts::QUOTE_SIGNATURE_TYPE,
    utils::{create_attestation_report, ServerAuth},
};

#[cfg(feature = "hardware_mode")]
pub fn get_client_config_epid() -> SgxResult<rustls::ClientConfig> {
    // Generate Keypair
    let ecc_handle = SgxEccHandle::new();
    ecc_handle.open().unwrap();

    let (prv_k, pub_k) = ecc_handle.create_key_pair().unwrap();

    let signed_report = create_attestation_report(&pub_k, QUOTE_SIGNATURE_TYPE)?;
    
    let payload: String = serde_json::to_string(&signed_report).map_err(|err| {
        println!("Error serializing report. May be malformed, or badly encoded: {:?}", err);
        sgx_status_t::SGX_ERROR_UNEXPECTED
    })?;

    let (key_der, cert_der) = gen_ecc_cert(payload, &prv_k, &pub_k, &ecc_handle)?;
    ecc_handle.close().unwrap();

    let cfg = construct_client_config(key_der, cert_der);
    Ok(cfg)
}

#[cfg(feature = "hardware_mode")]
fn get_client_config_dcap(
    qe_target_info: &sgx_target_info_t,
    quote_size: u32,
) -> SgxResult<rustls::ClientConfig> {
    // Generate Keypair
    let ecc_handle = SgxEccHandle::new();
    ecc_handle.open().unwrap();

    let (prv_k, pub_k) = ecc_handle.create_key_pair().unwrap();

    let signed_report = create_attestation_report(&pub_k, QUOTE_SIGNATURE_TYPE)?;

    let payload: String = serde_json::to_string(&signed_report).map_err(|err| {
        println!("Error serializing report. May be malformed, or badly encoded: {:?}", err);
        sgx_status_t::SGX_ERROR_UNEXPECTED
    })?;

    let (key_der, cert_der) = gen_ecc_cert(payload, &prv_k, &pub_k, &ecc_handle)?;
    ecc_handle.close().unwrap();

    let cfg = construct_client_config(key_der, cert_der);
    Ok(cfg)
}

/// Prepares config for client side of TLS connection
fn construct_client_config(key_der: Vec<u8>, cert_der: Vec<u8>) -> ClientConfig {
    let mut cfg = rustls::ClientConfig::new();
    let mut certs = Vec::new();
    certs.push(rustls::Certificate(cert_der));
    let privkey = rustls::PrivateKey(key_der);

    cfg.set_single_client_cert(certs, privkey).unwrap();
    cfg.dangerous()
        .set_certificate_verifier(Arc::new(ServerAuth::new(true)));
    cfg.versions.clear();
    cfg.versions.push(rustls::ProtocolVersion::TLSv1_2);
    cfg
}
