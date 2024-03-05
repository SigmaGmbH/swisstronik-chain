use sgx_types::*;
use std::io;
use std::io::{Read, Write};
use std::vec::Vec;
use std::string::String;

use crate::key_manager::{RegistrationKey, UNSEALED_KEY_MANAGER};

pub mod helpers;
pub mod auth;

/// Initializes new TLS client with report of Remote Attestation
pub fn perform_master_key_request(
    hostname: String,
    socket_fd: c_int,
    qe_target_info: Option<&sgx_target_info_t>,
    quote_size: Option<u32>,
) -> SgxResult<()> {
    let (key_der, cert_der) = helpers::create_tls_cert_and_keys(qe_target_info, quote_size)?;
    let client_config = helpers::construct_client_config(key_der, cert_der);

    // Prepare TLS connection
    let (mut sess, mut conn) =
        helpers::create_client_session_stream(hostname, socket_fd, client_config)?;
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
    helpers::decrypt_and_seal_master_key(&reg_key, &response)?;

    Ok(())
}

/// Initializes new TLS server to share master key
pub fn perform_master_key_provisioning(
    socket_fd: c_int,
    qe_target_info: Option<&sgx_target_info_t>,
    quote_size: Option<u32>,
) -> SgxResult<()> {
    let (key_der, cert_der) = helpers::create_tls_cert_and_keys(qe_target_info, quote_size)?;
    let server_config = helpers::construct_server_config(key_der, cert_der);

    // Prepare TLS connection
    let (mut sess, mut conn) = helpers::create_server_session_stream(socket_fd, server_config)?;
    let mut tls = rustls::Stream::new(&mut sess, &mut conn);

    // Read client registration public key
    let mut client_public_key = [0u8; 32];
    tls.read(&mut client_public_key).map_err(|err| {
        println!(
            "[Enclave] Attestation Server: error in read_to_end: {:?}",
            err
        );
        sgx_status_t::SGX_ERROR_UNEXPECTED
    })?;

    // Generate registration key for ECDH
    let registration_key = RegistrationKey::random()?;

    // Unseal key manager to get access to master key
    let key_manager = match &*UNSEALED_KEY_MANAGER {
        Some(key_manager) => key_manager,
        None => {
            println!("[Enclave] Cannot unseal master key");
            return Err(sgx_status_t::SGX_ERROR_UNEXPECTED);
        }
    };

    // Encrypt master key and send it to the client
    let encrypted_master_key = key_manager
        .to_encrypted_master_key(&registration_key, client_public_key.to_vec())
        .map_err(|err| {
            println!(
                "[Enclave] Cannot encrypt master key to share it. Reason: {:?}",
                err
            );
            sgx_status_t::SGX_ERROR_UNEXPECTED
        })?;

    // Send encrypted master key back to client
    tls.write(&encrypted_master_key).map_err(|err| {
        println!(
            "[Enclave] Cannot send encrypted master key to client. Reason: {:?}",
            err
        );
        sgx_status_t::SGX_ERROR_UNEXPECTED
    })?;

    Ok(())
}
