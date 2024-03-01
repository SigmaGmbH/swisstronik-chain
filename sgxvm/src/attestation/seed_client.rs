use sgx_types::*;

use rustls;
use std::io;
use std::io::{Read, Write};
use std::net::TcpStream;
use std::prelude::v1::*;
use std::sync::Arc;
use std::vec::Vec;
use rustls::ClientConfig;

use crate::attestation::consts::{ENCRYPTED_KEY_SIZE, PUBLIC_KEY_SIZE};
use crate::key_manager::{KeyManager, RegistrationKey};

#[cfg(feature = "hardware_mode")]
pub fn request_master_key(cfg: ClientConfig, hostname: String, socket_fd: c_int) -> sgx_status_t {
    let dns_name = match webpki::DNSNameRef::try_from_ascii_str(hostname.as_str()) {
        Ok(dns_name) => dns_name,
        Err(err) => {
            println!("[Enclave] Attestation Client: wrong host. Reason: {:?}", err);
            return sgx_status_t::SGX_ERROR_UNEXPECTED;
        }
    };
    let mut sess = rustls::ClientSession::new(&Arc::new(cfg), dns_name);
    let mut conn = match TcpStream::new(socket_fd) {
        Ok(conn) => conn,
        Err(err) => {
            println!(
                "[Enclave] Attestation Client: cannot establish tcp connection. Reason: {:?}",
                err
            );
            return sgx_status_t::SGX_ERROR_UNEXPECTED;
        }
    };

    let mut tls = rustls::Stream::new(&mut sess, &mut conn);

    // Generate temporary registration key used for master key encryption during transfer
    let registration_key = match RegistrationKey::random() {
        Ok(key) => key,
        Err(err) => return err,
    };

    // Send client public key to the master key exchange server
    if let Err(err) = tls.write(registration_key.public_key().as_bytes()) {
        println!(
            "[Enclave] Attestation Client: cannot send public key to server. Reason: {:?}",
            err
        );
        return sgx_status_t::SGX_ERROR_UNEXPECTED;
    }

    let mut plaintext = Vec::new();
    match tls.read_to_end(&mut plaintext) {
        Err(ref err) if err.kind() == io::ErrorKind::ConnectionAborted => {
            println!("[Enclave] Attestation Client: connection aborted");
            return sgx_status_t::SGX_ERROR_UNEXPECTED;
        }
        Err(e) => {
            println!("[Enclave] Attestation Client: error in read_to_end: {:?}", e);
            return sgx_status_t::SGX_ERROR_UNEXPECTED;
        }
        _ => {}
    };

    // Check size of response. It should be equal or more 90 bytes
    // 32 public key | 16 nonce | ciphertext
    if plaintext.len() < ENCRYPTED_KEY_SIZE {
        println!("[Enclave] Attestation Client: wrong response size");
        return sgx_status_t::SGX_ERROR_UNEXPECTED;
    }

    // Extract public key and nonce + ciphertext
    let public_key = &plaintext[..PUBLIC_KEY_SIZE];
    let encrypted_seed = &plaintext[PUBLIC_KEY_SIZE..];

    // Construct key manager
    let key_manager: Result<_, _> = KeyManager::from_encrypted_master_key(
        &registration_key,
        public_key.to_vec(),
        encrypted_seed.to_vec(),
    );
    let key_manager = match key_manager {
        Ok(key_manager) => key_manager,
        Err(err) => {
            println!(
                "[Enclave] Attestation Client: cannot construct key manager. Reason: {:?}",
                err
            );
            return sgx_status_t::SGX_ERROR_UNEXPECTED;
        }
    };

    // Seal master key
    if let Err(error_status) = key_manager.seal() {
        println!(
            "[Enclave] Attestation Client: cannot seal master key. Reason: {:?}",
            error_status.as_str()
        );
        return error_status;
    }


    println!("[Enclave] Attestation successfully sealed");
    
    sgx_status_t::SGX_SUCCESS
}

#[cfg(not(feature = "hardware_mode"))]
pub fn request_master_key(_cfg: ClientConfig, _hostname: String, socket_fd: c_int) -> sgx_status_t {
    println!("[Enclave] Cannot perform Remote Attestation in Software Mode");
    sgx_status_t::SGX_ERROR_UNEXPECTED
}
