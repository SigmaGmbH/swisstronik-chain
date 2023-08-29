use crate::{error::Error, key_manager::{PUBLIC_KEY_SIZE, self}};
use std::vec::Vec;

use crate::key_manager::UNSEALED_KEY_MANAGER;

pub const FUNCTION_SELECTOR_LEN: usize = 4;
pub const ZERO_FUNCTION_SELECTOR: [u8; 4] = [0u8; 4];
pub const PUBLIC_KEY_ONLY_DATA_LEN: usize = 36;
pub const ENCRYPTED_DATA_LEN: usize = 79;
pub const DEFAULT_STORAGE_VALUE: [u8; 32] = [0u8; 32];

/// Encrypts given storage cell value using sealed master key
pub fn encrypt_storage_cell(contract_address: Vec<u8>, value: Vec<u8>) -> Result<Vec<u8>, Error> {
    let key_manager = match &*UNSEALED_KEY_MANAGER {
        Some(key_manager) => key_manager,
        None => {
            return Err(Error::encryption_err(format!("Cannot unseal master key")));
        }
    };

    key_manager.encrypt_state(contract_address, value)
}

/// Decrypts given storage cell value using sealed master key
pub fn decrypt_storage_cell(contract_address: Vec<u8>, encrypted_value: Vec<u8>) -> Result<Vec<u8>, Error> {
    // It there is 32-byte zeroed vector, it means that storage slot was not initialized
    // In this case we return default value
    if encrypted_value == DEFAULT_STORAGE_VALUE.to_vec() {
        return Ok(encrypted_value)
    }

    let key_manager = match &*UNSEALED_KEY_MANAGER {
        Some(key_manager) => key_manager,
        None => {
            return Err(Error::encryption_err(format!("Cannot unseal master key")));
        }
    };

    key_manager.decrypt_state(contract_address, encrypted_value)
}

/// Extracts user public and encrypted data from provided tx `data` field.
/// If data starts with 0x00000000 prefix and has 36 bytes length, it means that there is only public key and no ciphertext.
/// If data has length of 78 and more bytes, we handle it as encrypted data
/// * tx_data – `data` field of transaction
pub fn extract_public_key_and_data(tx_data: Vec<u8>) -> Result<(Vec<u8>, Vec<u8>), Error> {
    // Check if provided tx data starts with `ZERO_FUNCTION_SELECTOR`
    // and has length of 36 bytes (4 prefix | 32 public key)
    if tx_data.len() == PUBLIC_KEY_ONLY_DATA_LEN && &tx_data[..4] == ZERO_FUNCTION_SELECTOR {
        let public_key = &tx_data[FUNCTION_SELECTOR_LEN..PUBLIC_KEY_ONLY_DATA_LEN];
        // Return extracted public key and empty ciphertext
        return Ok((public_key.to_vec(), Vec::default()))
    }

    // Otherwise check if tx data has length of 79
    // or more bytes (32 public key | 15 nonce | 16 ad | 16+ ciphertext)
    // If it is not, throw an ECDH error
    if tx_data.len() < ENCRYPTED_DATA_LEN {
        return Err(Error::ecdh_err("Wrong public key size"));
    }

    // Extract public key & encrypted data
    let public_key = &tx_data[..PUBLIC_KEY_SIZE];
    let encrypted_data = &tx_data[PUBLIC_KEY_SIZE..];

    Ok((public_key.to_vec(), encrypted_data.to_vec()))
}

/// Decrypts transaction data using derived shared secret
pub fn decrypt_transaction_data(encrypted_data: Vec<u8>, public_key: Vec<u8>) -> Result<Vec<u8>, Error> {
    let key_manager = match &*UNSEALED_KEY_MANAGER {
        Some(key_manager) => key_manager,
        None => {
            return Err(Error::encryption_err(format!("Cannot unseal master key")));
        }
    };

    key_manager.decrypt_ecdh(public_key.to_vec(), encrypted_data)
}

pub fn encrypt_transaction_data(data: Vec<u8>, user_public_key: Vec<u8>) -> Result<Vec<u8>, Error> {
    if user_public_key.len() < PUBLIC_KEY_SIZE {
        return Err(Error::ecdh_err("Wrong public key size"));
    }

    let key_manager = match &*UNSEALED_KEY_MANAGER {
        Some(key_manager) => key_manager,
        None => {
            return Err(Error::encryption_err(format!("Cannot unseal master key")));
        }
    };

    key_manager.encrypt_ecdh(data, user_public_key)
}
