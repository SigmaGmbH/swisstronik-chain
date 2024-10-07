#[cfg(feature = "std")]
use std::vec::Vec;

#[cfg(not(feature = "std"))]
use sgx_tstd::vec::Vec;

use core::cmp::min;
use evm::interpreter::error::{ExitException, ExitResult, ExitSucceed};
use p256::ecdsa::{signature::hazmat::PrehashVerifier, Signature, VerifyingKey};

use crate::LinearCostPrecompile;

pub struct P256Verify;

impl LinearCostPrecompile for P256Verify {
    const BASE: u64 = 3450;
    const WORD: u64 = 0;

    fn raw_execute(i: &[u8], target_gas: u64) -> (ExitResult, Vec<u8>){
        if i.len() < 160 {
            return (ExitException::Other("input must contain 160 bytes".into()).into(), Vec::new());
        };
        const P256VERIFY_BASE: u64 = 3_450;

        if P256VERIFY_BASE > target_gas {
            return (ExitException::OutOfGas.into(), Vec::new());
        }
        let mut input = [0u8; 160];
        input[..min(i.len(), 160)].copy_from_slice(&i[..min(i.len(), 160)]);

        // msg signed (msg is already the hash of the original message)
        let msg: [u8; 32] = input[..32].try_into().unwrap();
        // r, s: signature
        let sig: [u8; 64] = input[32..96].try_into().unwrap();
        // x, y: public key
        let pk: [u8; 64] = input[96..160].try_into().unwrap();
        // append 0x04 to the public key: uncompressed form
        let mut uncompressed_pk = [0u8; 65];
        uncompressed_pk[0] = 0x04;
        uncompressed_pk[1..].copy_from_slice(&pk);

        let public_key = match VerifyingKey::from_sec1_bytes(&uncompressed_pk) {
            Ok(v) => v,
            Err(_) => return (ExitException::Other("Public key recover failed".into()).into(), Vec::new())
        };

        let signature = match Signature::from_slice(&sig) {
            Ok(sig) => sig,
            Err(_) => return (ExitException::Other("Signature recover failed".into()).into(), Vec::new())
        };

        let mut buf = [0u8; 32];

        // verify
        if public_key.verify_prehash(&msg, &signature).is_ok() {
            buf[31] = 1u8;
        } else {
            buf[31] = 0u8;
        }

        (ExitSucceed::Returned.into(), buf.to_vec())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_empty_input() {
        let input: [u8; 0] = [];
        let cost: u64 = 1;

        let (success, _) = P256Verify::raw_execute(&input, cost);
        assert_eq!(success, ExitException::Other("input must contain 160 bytes".into()).into());
    }

    #[test]
    fn proper_sig_verify() {
        let input = hex::decode("4cee90eb86eaa050036147a12d49004b6b9c72bd725d39d4785011fe190f0b4da73bd4903f0ce3b639bbbf6e8e80d16931ff4bcf5993d58468e8fb19086e8cac36dbcd03009df8c59286b162af3bd7fcc0450c9aa81be5d10d312af6c66b1d604aebd3099c618202fcfe16ae7770b0c49ab5eadf74b754204a3bb6060e44eff37618b065f9832de4ca6ca971a7a1adc826d0f7c00181a5fb2ddf79ae00b4e10e").unwrap();
        let target_gas = 3_500u64;
        let (success, res) = P256Verify::raw_execute(&input, target_gas);
        assert_eq!(success, ExitSucceed::Returned.into());
        assert_eq!(res.len(), 32);
        assert_eq!(res[0], 0u8);
        assert_eq!(res[1], 0u8);
        assert_eq!(res[2], 0u8);
        assert_eq!(res[31], 1u8);
    }
}