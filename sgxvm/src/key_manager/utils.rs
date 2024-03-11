use hmac::{Hmac, Mac, NewMac as _};
use super::PRIVATE_KEY_SIZE;

pub fn derive_key(master_key: &[u8; PRIVATE_KEY_SIZE], info: &[u8]) -> [u8; PRIVATE_KEY_SIZE] {
    let mut kdf = Hmac::<sha2::Sha256>::new_from_slice(info).expect("Unable to create KDF");
    kdf.update(master_key);
    let mut derived_key = [0u8; PRIVATE_KEY_SIZE];
    let digest = kdf.finalize();
    derived_key.copy_from_slice(&digest.into_bytes()[..PRIVATE_KEY_SIZE]);

    derived_key
}