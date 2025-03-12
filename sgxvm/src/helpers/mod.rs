use sha3::{Keccak256, Digest};
use k256::{
    ecdsa::recoverable,
    elliptic_curve::{sec1::ToEncodedPoint, IsHigh},
};

pub mod tx;

pub fn recover_sender(msg: &[u8; 32], sig: &[u8; 65]) -> Option<[u8; 20]> {
    let mut sig_buf = [0u8; 65];
    sig_buf.copy_from_slice(sig);

    if sig_buf[64] > 26 {
        sig_buf[64] = sig[64] - 27
    }

    let signature = match recoverable::Signature::try_from(&sig_buf) {
        Ok(signature) => signature,
        Err(_) => return None,
    };

    let recovered_key = match signature.recover_verifying_key_from_digest_bytes(&msg.into()) {
        Ok(key) => key,
        Err(_) => return None,
    };

    let public_key = recovered_key.to_encoded_point(false);
    let mut hasher = Keccak256::new();
    hasher.update(&public_key.as_bytes()[1..]); // Skip the compression byte
    let hash = hasher.finalize();

    let mut address = [0u8; 20];
    address.copy_from_slice(&hash[12..32]);
    Some(address)
}