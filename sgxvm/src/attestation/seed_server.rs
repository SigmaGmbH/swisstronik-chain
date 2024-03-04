use sgx_tcrypto::*;
use sgx_types::*;

use rustls;
use std::io::{Read, Write};
use std::net::TcpStream;
use std::prelude::v1::*;
use std::sync::Arc;
use std::vec::Vec;

use crate::attestation::{
    consts::QUOTE_SIGNATURE_TYPE,
    utils::{create_attestation_report, ClientAuth},
    cert::gen_ecc_cert,
};
use crate::key_manager::{RegistrationKey, UNSEALED_KEY_MANAGER};

#[cfg(feature = "hardware_mode")]
pub fn share_seed_inner(socket_fd: c_int) -> sgx_status_t {
    let cfg = match get_server_configuration() {
        Ok(cfg) => cfg,
        Err(err) => {
            println!("{}", err);
            return sgx_status_t::SGX_ERROR_UNEXPECTED;
        }
    };

    let mut sess = rustls::ServerSession::new(&Arc::new(cfg));
    let mut conn = match TcpStream::new(socket_fd) {
        Ok(conn) => conn,
        Err(err) => {
            println!(
                "[Enclave] Seed Server: cannot establish connection with client: {:?}",
                err
            );
            return sgx_status_t::SGX_ERROR_UNEXPECTED;
        }
    };

    let mut tls = rustls::Stream::new(&mut sess, &mut conn);
    let mut client_public_key = [0u8; 32];
    if let Err(err) = tls.read(&mut client_public_key) {
        println!("[Enclave] Seed Server: error in read_to_end: {:?}", err);
        return sgx_status_t::SGX_ERROR_UNEXPECTED;
    };

    // Generate registration key for ECDH
    let registration_key = match RegistrationKey::random() {
        Ok(key) => key,
        Err(err) => {
            return err;
        }
    };

    // Unseal key manager to get access to master key
    let key_manager = match &*UNSEALED_KEY_MANAGER {
        Some(key_manager) => key_manager,
        None => {
            println!("Cannot unseal master key");
            return sgx_status_t::SGX_ERROR_UNEXPECTED;
        }
    };

    // Encrypt master key and send it to the client
    let encrypted_master_key = match key_manager.to_encrypted_master_key(&registration_key, client_public_key.to_vec()) {
        Ok(ciphertext) => ciphertext,
        Err(err) => {
            println!("[Enclave] Cannot encrypt master key. Reason: {:?}", err);
            return sgx_status_t::SGX_ERROR_UNEXPECTED;
        }
    };

    // Send encrypted master key back to client
    match tls.write(encrypted_master_key.as_slice()) {
        Ok(_) => sgx_status_t::SGX_SUCCESS,
        Err(err) => {
            println!("[Enclave] Cannot send encrypted master key to client. Reason: {:?}", err);
            sgx_status_t::SGX_ERROR_UNEXPECTED
        }
    }
}

#[cfg(not(feature = "hardware_mode"))]
pub fn share_seed_inner(socket_fd: c_int) -> sgx_status_t {
    println!("[Enclave] Cannot attest peer in software mode");
    sgx_status_t::SGX_ERROR_UNEXPECTED
}

#[cfg(feature = "hardware_mode")]
fn get_server_configuration() -> Result<rustls::ServerConfig, String> {
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
