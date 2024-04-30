use sgx_types::{sgx_status_t, SgxResult};
use serde::{Deserialize, Serialize};
use std::vec::Vec;
use std::string::String;

use crate::key_manager::{utils, keys, consts};
use crate::encryption;

#[derive(Serialize, Deserialize)]
pub struct Epoch {
    epoch_number: u16,
    epoch_key: [u8; 32],
    starting_block: u64
}

impl Epoch {
    pub fn get_tx_key(&self) -> keys::TransactionEncryptionKey {
        let tx_key_bytes = utils::derive_key(&self.epoch_key, consts::TX_KEY_PREFIX);
        keys::TransactionEncryptionKey::from(tx_key_bytes)
    }

    pub fn get_state_key(&self) -> keys::StateEncryptionKey {
        let state_key_bytes = utils::derive_key(&self.epoch_key, consts::STATE_KEY_PREFIX);
        keys::StateEncryptionKey::from(state_key_bytes)
    }
}

#[derive(Serialize, Deserialize)]
pub struct EpochManager {
    epochs: Vec<Epoch>
}

impl EpochManager {
    pub fn get_latest_epoch(&self) -> SgxResult<Epoch> {
        match self.epochs.into_iter().max_by_key(|epoch| epoch.epoch_key) {
            Some(epoch) => Ok(epoch),
            None => {
                println!("[EpochManager] No epoch data found");
                Err(sgx_status_t::SGX_ERROR_UNEXPECTED)
            }
        }
    }

    pub fn serialize(&self) -> SgxResult<String> {
        let encoded = serde_json::to_string(&self).map_err(|err| {
            println!("[EpochManager] Cannot serialize. Reason: {:?}", err);
            sgx_status_t::SGX_ERROR_UNEXPECTED
        })?;

        Ok(encoded)
    }

    pub fn deserialize(input: &str) -> SgxResult<Self> {
        let epoch_manager: EpochManager = serde_json::from_str(input).map_err(|err| {
            println!("[EpochManager] Cannot deserialize. Reason: {:?}", err);
            sgx_status_t::SGX_ERROR_UNEXPECTED
        })?;

        Ok(epoch_manager)
    }

    pub fn deserialize_from_slice(input: &[u8]) -> SgxResult<Self> {
        let epoch_manager: EpochManager = serde_json::from_slice(input).map_err(|err| {
            println!("[EpochManager] Cannot deserialize from slice. Reason: {:?}", err);
            sgx_status_t::SGX_ERROR_UNEXPECTED
        })?;

        Ok(epoch_manager)
    }

    pub fn random_with_single_epoch() -> SgxResult<Self> {
        let epoch_key = utils::random_bytes32().map_err(|err| {
            println!("[KeyManager] Cannot create random epoch key. Reason: {:?}", err);
            err
        })?;
        let epoch_number = 0u16;
        let starting_block = 0u64;

        let epoch = Epoch {epoch_number, epoch_key, starting_block};
        Ok(Self {
            epochs: vec![epoch],
        })
    }

    #[cfg(feature = "attestation_server")]
    pub fn encrypt(
        &self,
        reg_key: &keys::RegistrationKey,
        public_key: Vec<u8>,
    ) -> SgxResult<Vec<u8>> {
        // Convert public key to appropriate format
        let public_key: [u8; 32] = public_key.try_into().map_err(|err| {
            println!("[EpochManager] Cannot convert public key during encryption. Reason: {:?}", err);
            sgx_status_t::SGX_ERROR_UNEXPECTED
        })?;
        let public_key = x25519_dalek::PublicKey::from(public_key);

        let shared_secret = reg_key.diffie_hellman(public_key);
        let encoded_epoch_manager = self.serialize()?;
        let encrypted_value = encryption::encrypt_deoxys(shared_secret.as_bytes(), encoded_epoch_manager.as_bytes().to_vec(), None).map_err(|err| {
            println!("[EpochManager] Cannot encrypt serialized epoch manager. Reason: {:?}", err);
            sgx_status_t::SGX_ERROR_UNEXPECTED
        })?;

        // Add public key as prefix
        let reg_public_key = reg_key.public_key();
        Ok([reg_public_key.as_bytes(), encrypted_value.as_slice()].concat())
    }

    pub fn decrypt(
        reg_key: &keys::RegistrationKey,
        public_key: Vec<u8>,
        encrypted_epoch_data: Vec<u8>,
    ) -> SgxResult<Self> {
        // Convert public key to appropriate format
        let public_key: [u8; 32] = public_key.try_into().map_err(|err| {
            println!("[EpochManager] Cannot convert public key during decryption. Reason: {:?}", err);
            sgx_status_t::SGX_ERROR_UNEXPECTED
        })?;
        let public_key = x25519_dalek::PublicKey::from(public_key);

        // Derive shared secret
        let shared_secret = reg_key.diffie_hellman(public_key);

        // Decrypt epoch data
        let epoch_data = encryption::decrypt_deoxys(shared_secret.as_bytes(), encrypted_epoch_data).map_err(|err| {
            println!("[EpochManager] Cannot decrypt serialized epoch manager. Reason: {:?}", err);
            sgx_status_t::SGX_ERROR_UNEXPECTED
        })?;
        let epoch_manager = EpochManager::deserialize_from_slice(&epoch_data)?;

        Ok(epoch_manager)
    }
}

