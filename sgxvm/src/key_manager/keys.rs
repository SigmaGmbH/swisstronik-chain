use sgx_types::{sgx_read_rand, sgx_status_t, SgxResult};

pub const REGISTRATION_KEY_SIZE: usize = 32;

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
            println!("[Enclave] Cannot generate random reg key. Reason: {:?}", res);
            return Err(res);
        }

        Ok(Self { inner: buffer})
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