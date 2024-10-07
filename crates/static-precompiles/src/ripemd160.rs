#[cfg(feature = "std")]
use std::vec::Vec;

#[cfg(not(feature = "std"))]
use sgx_tstd::vec::Vec;

use evm::interpreter::error::{ExitResult, ExitSucceed};
use k256::sha2::Digest;
use crate::LinearCostPrecompile;

/// The ripemd precompile.
pub struct Ripemd160;

impl LinearCostPrecompile for Ripemd160 {
    const BASE: u64 = 600;
    const WORD: u64 = 120;

    fn raw_execute(input: &[u8], _cost: u64) -> (ExitResult, Vec<u8>) {
        let mut ret = [0u8; 32];
        ret[12..32].copy_from_slice(&ripemd::Ripemd160::digest(input));
        (ExitSucceed::Returned.into(), ret.to_vec())
    }
}