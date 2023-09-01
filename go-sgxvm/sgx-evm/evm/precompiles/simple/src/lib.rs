#![no_std]

extern crate alloc;

use alloc::vec::Vec;
use core::cmp::min;

use precompile_std::{ExitSucceed, LinearCostPrecompile, PrecompileFailure};
use k256::sha2::{Sha256 as kSha256, Digest};
use sha3::{Keccak256};
use k256::{
    ecdsa::recoverable,
    elliptic_curve::{sec1::ToEncodedPoint, IsHigh},
};

/// The identity precompile.
pub struct Identity;

impl LinearCostPrecompile for Identity {
    const BASE: u64 = 15;
    const WORD: u64 = 3;

    fn execute(input: &[u8], _: u64) -> Result<(ExitSucceed, Vec<u8>), PrecompileFailure> {
        Ok((ExitSucceed::Returned, input.to_vec()))
    }
}

// The ecrecover precompile.
pub struct ECRecover;

impl LinearCostPrecompile for ECRecover {
    const BASE: u64 = 3000;
    const WORD: u64 = 0;

    fn execute(i: &[u8], _: u64) -> Result<(ExitSucceed, Vec<u8>), PrecompileFailure> {
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
            return Ok((ExitSucceed::Returned, [0u8; 0].to_vec()));
        }

        let signature = match recoverable::Signature::try_from(&sig[..]) {
            Ok(signature) => signature,
            Err(_) => {
                return Ok((ExitSucceed::Returned, [0u8; 0].to_vec()));
            }
        };

        if signature.s().is_high().into() {
            return Ok((ExitSucceed::Returned, [0u8; 0].to_vec()));
        }

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

        Ok((ExitSucceed::Returned, result))
    }
}

/// The ripemd precompile.
pub struct Ripemd160;

impl LinearCostPrecompile for Ripemd160 {
    const BASE: u64 = 600;
    const WORD: u64 = 120;

    fn execute(input: &[u8], _cost: u64) -> Result<(ExitSucceed, Vec<u8>), PrecompileFailure> {
        let mut ret = [0u8; 32];
        ret[12..32].copy_from_slice(&ripemd::Ripemd160::digest(input));
        Ok((ExitSucceed::Returned, ret.to_vec()))
    }
}

/// The sha256 precompile.
pub struct Sha256;

impl LinearCostPrecompile for Sha256 {
    const BASE: u64 = 60;
    const WORD: u64 = 12;

    fn execute(input: &[u8], _cost: u64) -> Result<(ExitSucceed, Vec<u8>), PrecompileFailure> {
        let mut hasher = kSha256::new();
        hasher.update(input);
        let result = hasher.finalize();
        Ok((ExitSucceed::Returned, result.to_vec()))
    }
}

/// The ECRecoverPublicKey precompile.
/// Similar to ECRecover, but returns the pubkey (not the corresponding Ethereum address)
// pub struct ECRecoverPublicKey;
//
// impl LinearCostPrecompile for ECRecoverPublicKey {
//     const BASE: u64 = 3000;
//     const WORD: u64 = 0;
//
//     fn execute(i: &[u8], _: u64) -> Result<(ExitSucceed, Vec<u8>), PrecompileFailure> {
//         let mut input = [0u8; 128];
//         input[..min(i.len(), 128)].copy_from_slice(&i[..min(i.len(), 128)]);
//
//         let mut msg = [0u8; 32];
//         let mut sig = [0u8; 65];
//
//         msg[0..32].copy_from_slice(&input[0..32]);
//         sig[0..32].copy_from_slice(&input[64..96]);
//         sig[32..64].copy_from_slice(&input[96..128]);
//         sig[64] = input[63];
//
//         let pubkey = sp_io::crypto::secp256k1_ecdsa_recover(&sig, &msg).map_err(|_| {
//             PrecompileFailure::Error {
//                 exit_status: ExitError::Other("Public key recover failed".into()),
//             }
//         })?;
//
//         Ok((ExitSucceed::Returned, pubkey.to_vec()))
//     }
// }

#[cfg(test)]
mod tests {
    use super::*;
    use pallet_evm_test_vector_support::test_precompile_test_vectors;

    // TODO: this fails on the test "InvalidHighV-bits-1" where it is expected to return ""
    #[test]
    fn process_consensus_tests_for_ecrecover() -> Result<(), String> {
        test_precompile_test_vectors::<ECRecover>("../testdata/ecRecover.json")?;
        Ok(())
    }

    #[test]
    fn process_consensus_tests_for_sha256() -> Result<(), String> {
        test_precompile_test_vectors::<Sha256>("../testdata/common_sha256.json")?;
        Ok(())
    }

    #[test]
    fn process_consensus_tests_for_ripemd160() -> Result<(), String> {
        test_precompile_test_vectors::<Ripemd160>("../testdata/common_ripemd.json")?;
        Ok(())
    }
}
