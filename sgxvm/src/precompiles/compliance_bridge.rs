extern crate sgx_tstd as std;

use evm::executor::stack::{PrecompileHandle, PrecompileOutput};
use std::prelude::v1::*;
use std::vec::Vec;

use crate::GoQuerier;
use crate::precompiles::{
    ExitSucceed, LinearCostPrecompileWithQuerier, PrecompileFailure, PrecompileResult,
};

/// Precompile for interactions with x/compliance module.
pub struct ComplianceBridge;

impl LinearCostPrecompileWithQuerier for ComplianceBridge {
    const BASE: u64 = 60;
    const WORD: u64 = 150;

    fn execute(querier: *mut GoQuerier, handle: &mut impl PrecompileHandle) -> PrecompileResult {
        let target_gas = handle.gas_limit();
        let cost = crate::precompiles::ensure_linear_cost(
            target_gas,
            handle.input().len() as u64,
            Self::BASE,
            Self::WORD,
        )?;

        handle.record_cost(cost)?;
        let (exit_status, output) = execute_inner(querier, handle.input())?;
        Ok(PrecompileOutput {
            exit_status,
            output,
        })
    }
}

fn execute_inner(
    querier: *mut GoQuerier,
    input: &[u8],
) -> Result<(ExitSucceed, Vec<u8>), PrecompileFailure> {
    Ok((ExitSucceed::Returned, Vec::default()))
}
