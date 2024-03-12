use deoxysii::NONCE_SIZE;

use crate::{error::Error, key_manager::PUBLIC_KEY_SIZE};
use std::vec::Vec;

use crate::key_manager::UNSEALED_KEY_MANAGER;

pub const FUNCTION_SELECTOR_LEN: usize = 4;
pub const ZERO_FUNCTION_SELECTOR: [u8; 4] = [0u8; 4];
pub const PUBLIC_KEY_ONLY_DATA_LEN: usize = 36;
pub const ENCRYPTED_DATA_LEN: usize = 79;
pub const DEFAULT_STORAGE_VALUE: [u8; 32] = [0u8; 32];

/// Encrypts given storage cell value using specific storage key for provided contract address
/// * contract_address - Address of the contract. Used to derive unique storage encryption key for state of this smart contract
/// * value - Raw storage value to encrypt
pub fn encrypt_storage_cell(contract_address: Vec<u8>, encryption_salt: Vec<u8>, value: Vec<u8>) -> Result<Vec<u8>, Error> {
    if let Some(km) = &*UNSEALED_KEY_MANAGER {
        return km.encrypt_state(contract_address, encryption_salt, value)
    };

    Err(Error::encryption_err("Cannot unseal master key"))
}

/// Decrypts given storage cell value using specific storage key for provided contract address
/// * contract_address - Address of the contract. Used to derive unique storage encryption key for state of this smart contract
/// * value - Encrypted storage value
pub fn decrypt_storage_cell(contract_address: Vec<u8>, encrypted_value: Vec<u8>) -> Result<Vec<u8>, Error> {
    // It there is 32-byte zeroed vector, it means that storage slot was not initialized
    // In this case we return default value
    if encrypted_value == DEFAULT_STORAGE_VALUE.to_vec() {
        return Ok(encrypted_value)
    }

    if let Some(km) = &*UNSEALED_KEY_MANAGER {
        return km.decrypt_state(contract_address, encrypted_value);
    }

    return Err(Error::encryption_err(format!("Cannot unseal master key")));
}

/// Extracts user public and encrypted data from provided tx `data` field.
/// If data starts with 0x00000000 prefix and has 36 bytes length, it means that there is only public key and no ciphertext.
/// If data has length of 78 and more bytes, we handle it as encrypted data
/// * tx_data - `data` field of transaction
pub fn extract_public_key_and_data(tx_data: Vec<u8>) -> Result<(Vec<u8>, Vec<u8>, Vec<u8>), Error> {
    // Check if provided tx data starts with `ZERO_FUNCTION_SELECTOR`
    // and has length of 36 bytes (4 prefix | 32 public key)
    if tx_data.len() == PUBLIC_KEY_ONLY_DATA_LEN && &tx_data[..4] == ZERO_FUNCTION_SELECTOR {
        let public_key = &tx_data[FUNCTION_SELECTOR_LEN..PUBLIC_KEY_ONLY_DATA_LEN];
        // Return extracted public key and empty ciphertext
        return Ok((public_key.to_vec(), Vec::default(), Vec::default()))
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
    let nonce = &encrypted_data[..NONCE_SIZE];

    Ok((public_key.to_vec(), encrypted_data.to_vec(), nonce.to_vec()))
}

/// Decrypts transaction data using derived shared secret
/// * encrypted_data - Encrypted data 
/// * public_key - Public key provided by user
pub fn decrypt_transaction_data(encrypted_data: Vec<u8>, public_key: Vec<u8>) -> Result<Vec<u8>, Error> {
    // if let Some(km) = &*UNSEALED_KEY_MANAGER {
    //     // return km.decrypt_ecdh(public_key.to_vec(), encrypted_data);
    //     return km.tx_key.decrypt(public_key, encrypted_data);
    // }
    //
    // return Err(Error::encryption_err(format!("Cannot unseal master key")));
    match &*UNSEALED_KEY_MANAGER {
        Some(key_manager) => key_manager.tx_key.decrypt(public_key, encrypted_data),
        None => Err(Error::encryption_err("Cannot unseal master key"))
    }
}

/// Encrypts transaction data or response
/// * data - Raw transaction data or node response
/// * public_key - Public key provided by user
pub fn encrypt_transaction_data(data: Vec<u8>, user_public_key: Vec<u8>, nonce: Vec<u8>) -> Result<Vec<u8>, Error> {
    if user_public_key.len() != PUBLIC_KEY_SIZE {
        return Err(Error::ecdh_err("Wrong public key size"));
    }

    if nonce.len() != NONCE_SIZE {
        return Err(Error::ecdh_err("Wrong nonce size"));
    }

    match &*UNSEALED_KEY_MANAGER {
        Some(key_manager) => key_manager.tx_key.encrypt(user_public_key, data, nonce),
        None => Err(Error::encryption_err("Cannot unseal master key"))
    }

    // if let Some(km) = &*UNSEALED_KEY_MANAGER {
    //     return km.encrypt_ecdh(data, user_public_key, nonce);
    // }
    //
    // return Err(Error::encryption_err(format!("Cannot unseal master key")));
}
