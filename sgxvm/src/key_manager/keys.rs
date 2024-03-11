use crate::error::Error;
use crate::key_manager::{PUBLIC_KEY_SIZE, SEED_SIZE};
use sgx_types::{sgx_read_rand, sgx_status_t, SgxResult};
use std::vec::Vec;

pub const REGISTRATION_KEY_SIZE: usize = 32;
pub const PRIVATE_KEY_SIZE: usize = 32;

/// RegistrationKey handles all operations with registration key such as derivation of public key,
/// derivation of encryption key, etc.
pub struct RegistrationKey {
    inner: [u8; REGISTRATION_KEY_SIZE],
}

impl RegistrationKey {
    /// Generates public key for seed sharing
    pub fn public_key(&self) -> x25519_dalek::PublicKey {
        let secret = x25519_dalek::StaticSecret::from(self.inner);
        x25519_dalek::PublicKey::from(&secret)
    }

    /// Generates random registration key
    pub fn random() -> SgxResult<Self> {
        // Generate random seed
        let mut buffer = [0u8; REGISTRATION_KEY_SIZE];
        let res = unsafe { sgx_read_rand(&mut buffer as *mut u8, REGISTRATION_KEY_SIZE) };

        if res != sgx_status_t::SGX_SUCCESS {
            println!(
                "[Enclave] Cannot generate random reg key. Reason: {:?}",
                res
            );
            return Err(res);
        }

        Ok(Self { inner: buffer })
    }

    /// Performes Diffie-Hellman derivation of encryption key for master key encryption
    /// * public_key - User public key
    pub fn diffie_hellman(
        &self,
        public_key: x25519_dalek::PublicKey,
    ) -> x25519_dalek::SharedSecret {
        let secret = x25519_dalek::StaticSecret::from(self.inner);
        secret.diffie_hellman(&public_key)
    }
}

/// TransactionEncryptionKey is used to decrypt incoming transaction data and to encrypt enclave output
pub struct TransactionEncryptionKey {
    inner: [u8; PRIVATE_KEY_SIZE],
}

impl TransactionEncryptionKey {
    fn encrypt(
        &self,
        user_public_key: Vec<u8>,
        plaintext: Vec<u8>,
        salt: Vec<u8>,
    ) -> Result<Vec<u8>, Error> {
        // // Check if user_public_key has correct length
        // if user_public_key.len() != PUBLIC_KEY_SIZE {
        //     return Err(Error::encryption_err(format!(
        //         "[Encryption] Got public key with incorrect length. Expected: {:?}, Got: {:?}",
        //         user_public_key.len(),
        //         PUBLIC_KEY_SIZE
        //     )));
        // }

        // let public_key: [u8; PUBLIC_KEY_SIZE] = user_public_key.as_slice().try_into().map_err(|err| {
        //     Error::encryption_err("[Encryption] Wrong public key size");
        // })?;

        // let public_key = x25519_dalek::PublicKey::from(public_key);
        // // Convert master key to x25519 private key
        // let secret_key = x25519_dalek::StaticSecret::from(self.inner);
        // // Derive shared key
        // let shared_key = secret_key.diffie_hellman(&public_key);
        Ok(Vec::default())
    }
}

impl From<[u8; SEED_SIZE]> for TransactionEncryptionKey {
    fn from(input: [u8; SEED_SIZE]) -> Self {
        Self { inner: input }
    }
}
