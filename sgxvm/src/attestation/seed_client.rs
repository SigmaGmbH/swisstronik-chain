use sgx_tcrypto::*;
use sgx_types::*;

use rustls;
use core::slice;
use std::io;
use std::io::{Read, Write};
use std::net::TcpStream;
use std::prelude::v1::*;
use std::sync::Arc;
use std::vec::Vec;

use crate::attestation::{
    consts::{ENCRYPTED_KEY_SIZE, PUBLIC_KEY_SIZE, QUOTE_SIGNATURE_TYPE},
    cert::gen_ecc_cert,
    utils::{ServerAuth, create_attestation_report},
};
use crate::key_manager::{KeyManager, RegistrationKey};

#[no_mangle]
pub unsafe extern "C" fn ecall_request_seed(
    hostname: *const u8,
    data_len: usize,
    socket_fd: c_int,
) -> sgx_status_t {
    let hostname = slice::from_raw_parts(hostname, data_len);
    let hostname = match String::from_utf8(hostname.to_vec()) {
        Ok(hostname) => hostname,
        Err(err) => {
            println!("[Enclave] Seed Client. Cannot decode hostname. Reason: {:?}", err);
            return sgx_status_t::SGX_ERROR_UNEXPECTED;
        }
    };

    request_seed_inner(hostname, socket_fd)
}

#[cfg(feature = "hardware_mode")]
fn request_seed_inner(hostname: String, socket_fd: c_int) -> sgx_status_t {
    let cfg = match get_client_configuration() {
        Ok(cfg) => cfg,
        Err(err) => {
            println!(
                "[Enclave] Seed Client. Cannot construct client config. Reason: {}",
                err
            );
            return sgx_status_t::SGX_ERROR_UNEXPECTED;
        }
    };

    let dns_name = match webpki::DNSNameRef::try_from_ascii_str(hostname.as_str()) {
        Ok(dns_name) => dns_name,
        Err(err) => {
            println!("[Enclave] Seed Client: wrong host. Reason: {:?}", err);
            return sgx_status_t::SGX_ERROR_UNEXPECTED;
        }
    };
    let mut sess = rustls::ClientSession::new(&Arc::new(cfg), dns_name);
    let mut conn = match TcpStream::new(socket_fd) {
        Ok(conn) => conn,
        Err(err) => {
            println!(
                "[Enclave] Seed Client: cannot establish tcp connection. Reason: {:?}",
                err
            );
            return sgx_status_t::SGX_ERROR_UNEXPECTED;
        }
    };

    let mut tls = rustls::Stream::new(&mut sess, &mut conn);

    // Generate temporary registration key used for seed encryption during transfer
    let registration_key = match RegistrationKey::random() {
        Ok(key) => key,
        Err(err) => return err,
    };

    // Send client public key to the seed exchange server
    if let Err(err) = tls.write(registration_key.public_key().as_bytes()) {
        println!(
            "[Enclave] Seed Client: cannot send public key to server. Reason: {:?}",
            err
        );
        return sgx_status_t::SGX_ERROR_UNEXPECTED;
    }

    let mut plaintext = Vec::new();
    match tls.read_to_end(&mut plaintext) {
        Err(ref err) if err.kind() == io::ErrorKind::ConnectionAborted => {
            println!("[Enclave] Seed Client: connection aborted");
            return sgx_status_t::SGX_ERROR_UNEXPECTED;
        }
        Err(e) => {
            println!("[Enclave] Seed Client: error in read_to_end: {:?}", e);
            return sgx_status_t::SGX_ERROR_UNEXPECTED;
        }
        _ => {}
    };

    // Check size of response. It should be equal or more 90 bytes
    // 32 public key | 16 nonce | ciphertext
    if plaintext.len() < ENCRYPTED_KEY_SIZE {
        println!("[Enclave] Seed Client: wrong response size");
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
                "[Enclave] Seed Client: cannot construct key manager. Reason: {:?}",
                err
            );
            return sgx_status_t::SGX_ERROR_UNEXPECTED;
        }
    };

    // Seal master key
    if let Err(error_status) = key_manager.seal() {
        println!(
            "[Enclave] Seed Client: cannot seal master key. Reason: {:?}",
            error_status.as_str()
        );
        return error_status;
    }


    println!("[Enclave] Seed successfully sealed");
    
    sgx_status_t::SGX_SUCCESS
}

#[cfg(not(feature = "hardware_mode"))]
fn request_seed_inner(_hostname: String, socket_fd: c_int) -> sgx_status_t {
    let mut conn = TcpStream::new(socket_fd).unwrap();

    // Generate temporary registration key used for seed encryption during transfer
    let registration_key = match RegistrationKey::random() {
        Ok(key) => key,
        Err(err) => return err,
    };

    // Send client public key to the seed exchange server
    if let Err(err) = conn.write(registration_key.public_key().as_bytes()) {
        println!(
            "[Enclave] Seed Client: cannot send public key to server. Reason: {:?}",
            err
        );
        return sgx_status_t::SGX_ERROR_UNEXPECTED;
    }

    let mut plaintext = Vec::new();
    match conn.read_to_end(&mut plaintext) {
        Err(ref err) if err.kind() == io::ErrorKind::ConnectionAborted => {
            println!("[Enclave] Seed Client: connection aborted");
            return sgx_status_t::SGX_ERROR_UNEXPECTED;
        }
        Err(e) => {
            println!("[Enclave] Seed Client: error in read_to_end: {:?}", e);
            return sgx_status_t::SGX_ERROR_UNEXPECTED;
        }
        _ => {}
    };

    // Check size of response. It should be equal or more 90 bytes
    // 32 public key | 16 nonce | ciphertext
    if plaintext.len() < ENCRYPTED_KEY_SIZE {
        println!("[Enclave] Seed Client: wrong response size. Expected >90, Got {:?}", plaintext.len());
        return sgx_status_t::SGX_ERROR_UNEXPECTED;
    }

    // Extract public key and nonce + ciphertext
    let public_key = &plaintext[..PUBLIC_KEY_SIZE];
    let encrypted_seed = &plaintext[PUBLIC_KEY_SIZE..];

    // Construct key manager
    let key_manager = KeyManager::from_encrypted_master_key(
        &registration_key,
        public_key.to_vec(),
        encrypted_seed.to_vec(),
    );
    let key_manager = match key_manager {
        Ok(key_manager) => key_manager,
        Err(err) => {
            println!(
                "[Enclave] Seed Client: cannot construct key manager. Reason: {:?}",
                err
            );
            return sgx_status_t::SGX_ERROR_UNEXPECTED;
        }
    };

    // Seal master key
    if let Err(error_status) = key_manager.seal() {
        println!(
            "[Enclave] Seed Client: cannot seal master key. Reason: {:?}",
            error_status.as_str()
        );
        return error_status;
    }

    println!("[Enclave] Seed successfully sealed");

    sgx_status_t::SGX_SUCCESS
}

#[cfg(feature = "hardware_mode")]
fn get_client_configuration() -> Result<rustls::ClientConfig, String> {
    // Generate Keypair
    let ecc_handle = SgxEccHandle::new();
    match ecc_handle.open() {
        Err(err) => {
            return Err(format!("Cannot open SgxEccHandle. Reason: {:?}", err));
        },
        _ => {},
    };
    let (prv_k, pub_k) = match ecc_handle.create_key_pair() {
        Ok((prv_k, pub_k)) => (prv_k, pub_k),
        Err(err) => {
            return Err(format!("Cannot generate ecc keypair. Reason: {:?}", err));
        }
    };

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
    match ecc_handle.close() {
        Err(err) => {
            return Err(format!("Cannot close SgxEccHandle. Reason: {:?}", err));
        },
        _ => {},
    };

    let mut cfg = rustls::ClientConfig::new();
    let mut certs = Vec::new();
    certs.push(rustls::Certificate(cert_der));
    let privkey = rustls::PrivateKey(key_der);

    match cfg.set_single_client_cert(certs, privkey) {
        Err(err) => {
            return Err(format!("Cannot set client cert. Reason: {:?}", err));
        },
        _ => {},
    };
    cfg.dangerous()
        .set_certificate_verifier(Arc::new(ServerAuth::new(true)));
    cfg.versions.clear();
    cfg.versions.push(rustls::ProtocolVersion::TLSv1_2);

    Ok(cfg)
}
