#[cfg(feature = "std")]
use std::vec::Vec;

#[cfg(not(feature = "std"))]
use sgx_tstd::vec::Vec;

use curve25519_dalek::{
    ristretto::{CompressedRistretto, RistrettoPoint},
    scalar::Scalar,
    traits::Identity,
};
use ed25519_dalek::{Signature, Verifier, VerifyingKey};
use evm::interpreter::error::{ExitException, ExitResult, ExitSucceed};

use crate::LinearCostPrecompile;

pub struct Ed25519Verify;

impl LinearCostPrecompile for Ed25519Verify {
    const BASE: u64 = 2000;
    const WORD: u64 = 0;

    fn raw_execute(input: &[u8], _: u64) -> (ExitResult, Vec<u8>) {
        if input.len() < 128 {
            return (ExitException::Other("input must contain 128 bytes".into()).into(), Vec::new());
        };

        let mut i = [0u8; 128];
        i[..128].copy_from_slice(&input[..128]);

        let mut buf = [0u8; 32];

        let msg = &i[0..32];
        let pk = match VerifyingKey::try_from(&i[32..64]) {
            Ok(pk) => pk,
            Err(_) => return (ExitException::Other("Public key recover failed".into()).into(), Vec::new())
        };

        let sig = match Signature::try_from(&i[64..128]) {
            Ok(sig) => sig,
            Err(_) => return (ExitException::Other("Signature recover failed".into()).into(), Vec::new())
        };

        // https://docs.rs/rust-crypto/0.2.36/crypto/ed25519/fn.verify.html
        if pk.verify(msg, &sig).is_ok() {
            buf[31] = 0u8;
        } else {
            buf[31] = 1u8;
        };

        (ExitSucceed::Returned.into(), buf.to_vec())
    }
}

// Adds at most 10 curve25519 points and returns the CompressedRistretto bytes representation
pub struct Curve25519Add;

impl LinearCostPrecompile for Curve25519Add {
    const BASE: u64 = 150;
    const WORD: u64 = 0;

    fn raw_execute(input: &[u8], _: u64) -> (ExitResult, Vec<u8>) {
        if input.len() % 32 != 0 {
            return (ExitException::Other("input must contain multiple of 32 bytes".into()).into(), Vec::new());
        };

        if input.len() > 320 {
            return (ExitException::Other("input cannot be greater than 320 bytes (10 compressed points)".into()).into(), Vec::new());
        };

        let mut points = Vec::new();
        let mut temp_buf = <&[u8]>::clone(&input);
        while !temp_buf.is_empty() {
            let mut buf = [0; 32];
            buf.copy_from_slice(&temp_buf[0..32]);
            let point = CompressedRistretto(buf);
            points.push(point);
            temp_buf = &temp_buf[32..];
        }

        let sum = points
            .iter()
            .fold(RistrettoPoint::identity(), |acc, point| {
                let pt = point.decompress().unwrap_or_else(RistrettoPoint::identity);
                acc + pt
            });

        (ExitSucceed::Returned.into(), sum.compress().to_bytes().to_vec())
    }
}

// Multiplies a scalar field element with an elliptic curve point
pub struct Curve25519ScalarMul;

impl LinearCostPrecompile for Curve25519ScalarMul {
    const BASE: u64 = 6000;
    const WORD: u64 = 0;

    fn raw_execute(input: &[u8], _: u64) -> (ExitResult, Vec<u8>) {
        if input.len() != 64 {
            return (ExitException::Other("input must contain 64 bytes (scalar - 32 bytes, point - 32 bytes)".into()).into(), Vec::new());
        };

        // first 32 bytes is for the scalar value
        let mut scalar_buf = [0; 32];
        scalar_buf.copy_from_slice(&input[0..32]);
        let scalar = Scalar::from_bytes_mod_order(scalar_buf);

        // second 32 bytes is for the compressed ristretto point bytes
        let mut pt_buf = [0; 32];
        pt_buf.copy_from_slice(&input[32..64]);
        let point = CompressedRistretto(pt_buf)
            .decompress()
            .unwrap_or_else(RistrettoPoint::identity);

        let scalar_mul = scalar * point;
        (
            ExitSucceed::Returned.into(),
            scalar_mul.compress().to_bytes().to_vec(),
        )
    }
}

#[cfg(test)]
mod tests {
    use curve25519_dalek::constants;
    use ed25519_dalek::{Signer, SigningKey};

    use super::*;

    #[test]
    fn test_empty_input() {
        let input: [u8; 0] = [];
        let cost: u64 = 1;

        let (success, _) = Ed25519Verify::raw_execute(&input, cost);
        assert_eq!(success, ExitException::Other("input must contain 128 bytes".into()).into());
    }

    #[test]
    fn test_verify() {
        #[allow(clippy::zero_prefixed_literal)]
            let secret_key_bytes: [u8; ed25519_dalek::SECRET_KEY_LENGTH] = [
            157, 097, 177, 157, 239, 253, 090, 096, 186, 132, 074, 244, 146, 236, 044, 196, 068,
            073, 197, 105, 123, 050, 105, 025, 112, 059, 172, 003, 028, 174, 127, 096,
        ];

        let keypair = SigningKey::from_bytes(&secret_key_bytes);
        let public_key = keypair.verifying_key();

        let msg: &[u8] = b"abcdefghijklmnopqrstuvwxyz123456";
        assert_eq!(msg.len(), 32);
        let signature = keypair.sign(msg);

        // input is:
        // 1) message (32 bytes)
        // 2) pubkey (32 bytes)
        // 3) signature (64 bytes)
        let mut input: Vec<u8> = Vec::with_capacity(128);
        input.extend_from_slice(msg);
        input.extend_from_slice(&public_key.to_bytes());
        input.extend_from_slice(&signature.to_bytes());
        assert_eq!(input.len(), 128);

        let cost: u64 = 1;

        let (success, res) = Ed25519Verify::raw_execute(&input, cost);
        assert_eq!(res.len(), 32);
        assert_eq!(res[0], 0u8);
        assert_eq!(res[1], 0u8);
        assert_eq!(res[2], 0u8);
        assert_eq!(res[31], 0u8);
        assert_eq!(success, ExitSucceed::Returned.into());

        // try again with a different message
        let msg: &[u8] = b"BAD_MESSAGE_mnopqrstuvwxyz123456";

        let mut input: Vec<u8> = Vec::with_capacity(128);
        input.extend_from_slice(msg);
        input.extend_from_slice(&public_key.to_bytes());
        input.extend_from_slice(&signature.to_bytes());
        assert_eq!(input.len(), 128);

        let (success, output) = Ed25519Verify::raw_execute(&input, cost);
        assert_eq!(success, ExitSucceed::Returned.into());
        assert_eq!(output.len(), 32);
        assert_eq!(output[0], 0u8);
        assert_eq!(output[1], 0u8);
        assert_eq!(output[2], 0u8);
        assert_eq!(output[31], 1u8); // non-zero indicates error (in our case, 1)
    }

    #[test]
    fn test_sum() {
        let s1 = Scalar::from(999u64);
        let p1 = constants::RISTRETTO_BASEPOINT_POINT * s1;

        let s2 = Scalar::from(333u64);
        let p2 = constants::RISTRETTO_BASEPOINT_POINT * s2;

        let vec = vec![p1, p2];
        let mut input = vec![];
        input.extend_from_slice(&p1.compress().to_bytes());
        input.extend_from_slice(&p2.compress().to_bytes());

        let sum: RistrettoPoint = vec.iter().sum();
        let cost: u64 = 1;

        let (success, out) = Curve25519Add::raw_execute(&input, cost);
        assert_eq!(success, ExitSucceed::Returned.into());
        assert_eq!(out, sum.compress().to_bytes());
    }

    #[test]
    fn test_empty() {
        // Test that sum works for the empty iterator
        let input = vec![];

        let cost: u64 = 1;

        let (success, res) = Curve25519Add::raw_execute(&input, cost);
        assert_eq!(success, ExitSucceed::Returned.into());
        assert_eq!(res, RistrettoPoint::identity().compress().to_bytes());
    }

    #[test]
    fn test_scalar_mul() {
        let s1 = Scalar::from(999u64);
        let s2 = Scalar::from(333u64);
        let p1 = constants::RISTRETTO_BASEPOINT_POINT * s1;
        let p2 = constants::RISTRETTO_BASEPOINT_POINT * s2;

        let mut input = vec![];
        input.extend_from_slice(&s1.to_bytes());
        input.extend_from_slice(&constants::RISTRETTO_BASEPOINT_POINT.compress().to_bytes());

        let cost: u64 = 1;

        let (success, out) = Curve25519ScalarMul::raw_execute(&input, cost);
        assert_eq!(success, ExitSucceed::Returned.into());
        assert_eq!(out, p1.compress().to_bytes());
        assert_ne!(out, p2.compress().to_bytes());
    }

    #[test]
    fn test_scalar_mul_empty_error() {
        let input = vec![];

        let cost: u64 = 1;

        let (success, _) = Curve25519ScalarMul::raw_execute(&input, cost);
        assert_eq!(success, ExitException::Other("input must contain 64 bytes (scalar - 32 bytes, point - 32 bytes)".into()).into());
    }

    #[test]
    fn test_point_addition_bad_length() {
        let input: Vec<u8> = [0u8; 33].to_vec();

        let cost: u64 = 1;

        let (success, _) = Curve25519Add::raw_execute(&input, cost);
        assert_eq!(success, ExitException::Other("input must contain multiple of 32 bytes".into()).into());
    }

    #[test]
    fn test_point_addition_too_many_points() {
        let mut input = vec![];
        input.extend_from_slice(&constants::RISTRETTO_BASEPOINT_POINT.compress().to_bytes()); // 1
        input.extend_from_slice(&constants::RISTRETTO_BASEPOINT_POINT.compress().to_bytes()); // 2
        input.extend_from_slice(&constants::RISTRETTO_BASEPOINT_POINT.compress().to_bytes()); // 3
        input.extend_from_slice(&constants::RISTRETTO_BASEPOINT_POINT.compress().to_bytes()); // 4
        input.extend_from_slice(&constants::RISTRETTO_BASEPOINT_POINT.compress().to_bytes()); // 5
        input.extend_from_slice(&constants::RISTRETTO_BASEPOINT_POINT.compress().to_bytes()); // 6
        input.extend_from_slice(&constants::RISTRETTO_BASEPOINT_POINT.compress().to_bytes()); // 7
        input.extend_from_slice(&constants::RISTRETTO_BASEPOINT_POINT.compress().to_bytes()); // 8
        input.extend_from_slice(&constants::RISTRETTO_BASEPOINT_POINT.compress().to_bytes()); // 9
        input.extend_from_slice(&constants::RISTRETTO_BASEPOINT_POINT.compress().to_bytes()); // 10
        input.extend_from_slice(&constants::RISTRETTO_BASEPOINT_POINT.compress().to_bytes()); // 11

        let cost: u64 = 1;

        let (success, _) = Curve25519Add::raw_execute(&input, cost);
        assert_eq!(success, ExitException::Other("input cannot be greater than 320 bytes (10 compressed points)".into()).into());
    }
}