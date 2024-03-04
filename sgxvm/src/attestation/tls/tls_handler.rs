use rustls::{self, ClientConfig, ClientSession};
use sgx_tcrypto::*;
use sgx_types::*;
use std::io;
use std::io::{Read, Write};
use std::sync::Arc;
use std::vec::Vec;
use std::{net::TcpStream, string::String};

use crate::attestation::consts::{ENCRYPTED_KEY_SIZE, PUBLIC_KEY_SIZE};
use crate::attestation::{
    cert::gen_ecc_cert,
    consts::QUOTE_SIGNATURE_TYPE,
    dcap::get_qe_quote,
    utils::{create_attestation_report, ServerAuth},
};
use crate::key_manager::{KeyManager, RegistrationKey};

/// Initializes new TLS client with report of Remote Attestation
pub fn perform_master_key_request(
    hostname: String,
    socket_fd: c_int,
    qe_target_info: Option<&sgx_target_info_t>,
    quote_size: Option<u32>,
) -> SgxResult<()> {
    // Construct client config for TLS
    let client_config = match (qe_target_info, quote_size) {
        // Construct DCAP Report
        (Some(qe_target_info), Some(quote_size)) => {
            get_client_config_dcap(qe_target_info, quote_size)?
        }
        // Construct EPID Report
        _ => get_client_config_epid()?,
    };

    // Prepare TLS connection
    let (mut sess, mut conn) = create_client_session_stream(hostname, socket_fd, client_config)?;
    let mut tls = rustls::Stream::new(&mut sess, &mut conn);

    // Generate temporary registration key used for master key encryption during transfer
    let reg_key = RegistrationKey::random()?;

    // Send public key, derived from Registration key, to Attestation server
    tls.write(reg_key.public_key().as_bytes()).map_err(|err| {
        println!(
            "[Enclave] Cannot send public key to Attestation server. Reason: {:?}",
            err
        );
        sgx_status_t::SGX_ERROR_UNEXPECTED
    })?;

    // Wait for Attestation server response
    let mut response = Vec::new();
    match tls.read_to_end(&mut response) {
        Err(ref err) if err.kind() == io::ErrorKind::ConnectionAborted => {
            println!("[Enclave] Attestation Client: connection aborted");
            return Err(sgx_status_t::SGX_ERROR_UNEXPECTED);
        }
        Err(e) => {
            println!(
                "[Enclave] Attestation Client: error in read_to_end: {:?}",
                e
            );
            return Err(sgx_status_t::SGX_ERROR_UNEXPECTED);
        }
        _ => {}
    };

    // Decrypt and seal master key
    decrypt_and_seal_master_key(&reg_key, &response)?;

    Ok(())
}

/// Creates TLS client config with EPID Report
fn get_client_config_epid() -> SgxResult<rustls::ClientConfig> {
    // Generate Keypair
    let ecc_handle = SgxEccHandle::new();
    ecc_handle.open().unwrap();

    let (prv_k, pub_k) = ecc_handle.create_key_pair().unwrap();

    let signed_report = create_attestation_report(&pub_k, QUOTE_SIGNATURE_TYPE)?;

    let payload: String = serde_json::to_string(&signed_report).map_err(|err| {
        println!(
            "Error serializing report. May be malformed, or badly encoded: {:?}",
            err
        );
        sgx_status_t::SGX_ERROR_UNEXPECTED
    })?;

    let (key_der, cert_der) = gen_ecc_cert(payload, &prv_k, &pub_k, &ecc_handle)?;
    ecc_handle.close().unwrap();

    let cfg = construct_client_config(key_der, cert_der);
    Ok(cfg)
}

/// Creates TLS client config with DCAP Report
fn get_client_config_dcap(
    qe_target_info: &sgx_target_info_t,
    quote_size: u32,
) -> SgxResult<rustls::ClientConfig> {
    // Generate Keypair
    let ecc_handle = SgxEccHandle::new();
    let _ = ecc_handle.open();

    let (prv_k, pub_k) = ecc_handle.create_key_pair()?;
    let qe_quote = get_qe_quote(&pub_k, qe_target_info, quote_size)?;
    let qe_quote_base_64 = base64::encode(&qe_quote[..]);

    let (key_der, cert_der) = gen_ecc_cert(qe_quote_base_64, &prv_k, &pub_k, &ecc_handle)?;
    let _ = ecc_handle.close();

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

/// Creates TLS session stream for client
fn create_client_session_stream(
    hostname: String,
    socket_fd: c_int,
    cfg: ClientConfig,
) -> SgxResult<(ClientSession, TcpStream)> {
    let dns_name = webpki::DNSNameRef::try_from_ascii_str(&hostname).map_err(|err| {
        println!(
            "[Enclave] Cannot construct correct DNS name. Reason: {:?}",
            err
        );
        sgx_status_t::SGX_ERROR_INVALID_PARAMETER
    })?;

    let mut sess = rustls::ClientSession::new(&Arc::new(cfg), dns_name);
    let mut conn = TcpStream::new(socket_fd).map_err(|err| {
        println!("[Enclave] Cannot start TcpStream. Reason: {:?}", err);
        sgx_status_t::SGX_ERROR_UNEXPECTED
    })?;

    Ok((sess, conn))
}

/// Decrypts and seals received master key
fn decrypt_and_seal_master_key(
    reg_key: &RegistrationKey,
    attn_server_response: &Vec<u8>,
) -> SgxResult<()> {
    // Validate response size. It should be equal or more 90 bytes
    // 32 public key | 16 nonce | ciphertext
    if attn_server_response.len() < ENCRYPTED_KEY_SIZE {
        println!("[Enclave] Wrong response size from Attestation Server");
        return Err(sgx_status_t::SGX_ERROR_UNEXPECTED);
    }

    // Extract public key and nonce + ciphertext
    let public_key = &attn_server_response[..PUBLIC_KEY_SIZE];
    let encrypted_master_key = &attn_server_response[PUBLIC_KEY_SIZE..];

    // Construct key manager
    let km = KeyManager::from_encrypted_master_key(
        reg_key,
        public_key.to_vec(),
        encrypted_master_key.to_vec(),
    )
    .map_err(|err| {
        println!(
            "[Enclave] Cannot construct from encrypted master key. Reason: {:?}",
            err
        );
        sgx_status_t::SGX_ERROR_UNEXPECTED
    })?;

    // Seal decrypted master key
    km.seal()?;
    println!("[Enclave] Master key successfully sealed");

    Ok(())
}
