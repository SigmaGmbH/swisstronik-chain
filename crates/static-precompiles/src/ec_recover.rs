#[cfg(feature = "std")]
use std::vec::Vec;

#[cfg(not(feature = "std"))]
use sgx_tstd::vec::Vec;

use core::cmp::min;
use evm::interpreter::error::{ExitResult, ExitSucceed};
use k256::sha2::Digest;
use sha3::Keccak256;
use k256::{
    ecdsa::recoverable,
    elliptic_curve::{sec1::ToEncodedPoint},
};
use crate::LinearCostPrecompile;

// The ecrecover precompile.
pub struct ECRecover;

impl LinearCostPrecompile for ECRecover {
    const BASE: u64 = 3000;
    const WORD: u64 = 0;

    fn raw_execute(i: &[u8], _: u64) -> (ExitResult, Vec<u8>) {
        let mut input = [0u8; 128];
        input[..min(i.len(), 128)].copy_from_slice(&i[..min(i.len(), 128)]);

        let mut msg = [0u8; 32];
        let mut sig = [0u8; 65];

        msg.copy_from_slice(&input[0..32]);
        sig[0..64].copy_from_slice(&input[64..]);

        // EIP-155
        sig[64] = if input[63] > 26 {
            input[63] - 27
        } else {
            input[63]
        };

        if input[32..63] != [0u8; 31] {
            return (ExitSucceed::Returned.into(), [0u8; 0].to_vec());
        }

        let signature = match recoverable::Signature::try_from(&sig[..]) {
            Ok(signature) => signature,
            Err(_) => {
                return (ExitSucceed::Returned.into(), [0u8; 0].to_vec());
            }
        };

        let result = match signature.recover_verifying_key_from_digest_bytes(&msg.into()) {
            Ok(recovered_key) => {
                // Convert Ethereum style address
                let p = recovered_key.to_encoded_point(false);
                let mut hasher = Keccak256::new();
                hasher.update(&p.as_bytes()[1..]);
                let mut address = hasher.finalize();
                address[0..12].copy_from_slice(&[0u8; 12]);
                address.to_vec()
            }
            Err(_) => Vec::default(),
        };

        (ExitSucceed::Returned.into(), result)
    }
}