#![cfg_attr(not(feature = "std"), no_std)]

use evm::GasMutState;
use evm::interpreter::error::{ExitError, ExitException, ExitResult};
use evm::interpreter::runtime::RuntimeState;

#[cfg(feature = "std")]
use std::vec::Vec;

#[cfg(not(feature = "std"))]
use sgx_tstd::vec::Vec;

pub mod blake2f;
pub mod bn128;
pub mod curve25519;
pub mod modexp;
pub mod sha3fips;
pub mod ec_recover;
pub mod sha256;
pub mod ripemd160;
pub mod datacopy;
pub mod secp256r1;

pub trait Precompile<G> {
    fn execute(input: &[u8], gasometer: &mut G) -> (ExitResult, Vec<u8>);
}

pub trait LinearCostPrecompile {
    const BASE: u64;
    const WORD: u64;

    fn raw_execute(
        input: &[u8],
        cost: u64,
    ) -> (ExitResult, Vec<u8>);
}

impl<T: LinearCostPrecompile, G: AsRef<RuntimeState> + GasMutState> Precompile<G> for T {
    fn execute(input: &[u8], gasometer: &mut G) -> (ExitResult, Vec<u8>) {
        let cost = match linear_cost(input.len() as u64, T::BASE, T::WORD) {
            Ok(cost) => cost,
            Err(e) => return (Err(e), Vec::new()),
        };
        if let Err(err) = gasometer.record_gas(cost.into()) {
            return (err.into(), Vec::new());
        };

        T::raw_execute(input, cost)
    }
}

pub fn linear_cost(len: u64, base: u64, word: u64) -> Result<u64, ExitError> {
    let cost = base
        .checked_add(
            word.checked_mul(len.saturating_add(31) / 32)
                .ok_or(ExitException::OutOfGas)?,
        )
        .ok_or(ExitException::OutOfGas)?;

    Ok(cost)
}