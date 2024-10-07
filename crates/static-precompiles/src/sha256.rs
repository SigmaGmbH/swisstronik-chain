#[cfg(feature = "std")]
use std::vec::Vec;

#[cfg(not(feature = "std"))]
use sgx_tstd::vec::Vec;

use evm::interpreter::error::{ExitResult, ExitSucceed};
use k256::sha2::{
    Sha256 as kSha256, 
    Digest
};
use crate::LinearCostPrecompile;

/// The sha256 precompile.
pub struct Sha256;

impl LinearCostPrecompile for Sha256 {
    const BASE: u64 = 60;
    const WORD: u64 = 12;

    fn raw_execute(input: &[u8], _cost: u64) -> (ExitResult, Vec<u8>) {
        let mut hasher = kSha256::new();
        hasher.update(input);
        let result = hasher.finalize();
        (ExitSucceed::Returned.into(), result.to_vec())
    }
}