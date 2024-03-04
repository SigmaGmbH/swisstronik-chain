use sgx_tcrypto::*;
use sgx_types::*;
use rustls::{self, ClientConfig};
use std::prelude::v1::*;
use std::sync::Arc;
use std::vec::Vec;

use crate::attestation::{
    dcap::get_qe_quote,
    cert::gen_ecc_cert,
    consts::QUOTE_SIGNATURE_TYPE,
    utils::{create_attestation_report, ServerAuth},
};

#[cfg(feature = "hardware_mode")]
fn get_server_config_epid() -> Result<rustls::ServerConfig, String> {
    // Generate Keypair
    let ecc_handle = SgxEccHandle::new();
    let _result = ecc_handle.open();
    let (prv_k, pub_k) = ecc_handle.create_key_pair().unwrap();

    let signed_report = match create_attestation_report(&pub_k, QUOTE_SIGNATURE_TYPE)
    {
        Ok(r) => r,
        Err(e) => {
            return Err(format!("Error creating attestation report: {:?}", e.as_str()));
        }
    };

    let payload: String = match serde_json::to_string(&signed_report) {
        Ok(payload) => payload,
        Err(err) => {
            return Err(format!(
                "Error serializing report. May be malformed, or badly encoded: {:?}",
                err
            ));
        }
    };
    let (key_der, cert_der) = match gen_ecc_cert(payload, &prv_k, &pub_k, &ecc_handle)
    {
        Ok(r) => r,
        Err(e) => {
            return Err(format!("Error in gen_ecc_cert: {:?}", e));
        }
    };
    let _result = ecc_handle.close();

    let mut cfg = rustls::ServerConfig::new(Arc::new(ClientAuth::new(true)));
    let mut certs = Vec::new();
    certs.push(rustls::Certificate(cert_der));
    let privkey = rustls::PrivateKey(key_der);

    cfg.set_single_cert_with_ocsp_and_sct(certs, privkey, vec![], vec![])
        .unwrap();

    Ok(cfg)
}