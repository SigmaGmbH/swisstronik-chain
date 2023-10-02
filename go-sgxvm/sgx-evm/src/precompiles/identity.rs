extern crate sgx_tstd as std;

use std::vec::Vec;
use rlp::Rlp;
use evm::executor::stack::{PrecompileHandle, PrecompileOutput};
use crate::precompiles::{
    ExitError, 
    ExitSucceed, 
    LinearCostPrecompileWithQuerier, 
    PrecompileFailure,
    PrecompileResult,
};
use crate::querier::GoQuerier;

/// The identity precompile.
pub struct Identity;

impl LinearCostPrecompileWithQuerier for Identity {
    const BASE: u64 = 60;
    const WORD: u64 = 150;

    fn execute(querier: *mut GoQuerier, handle: &mut impl PrecompileHandle) -> PrecompileResult {
        let target_gas = handle.gas_limit();
        let cost = crate::precompiles::ensure_linear_cost(target_gas, handle.input().len() as u64, Self::BASE, Self::WORD)?;

        handle.record_cost(cost)?;
        let (exit_status, output) = Self::raw_execute(querier, handle.input(), cost)?;
        Ok(PrecompileOutput {
            exit_status,
            output,
        })
    }

    fn raw_execute(querier: *mut GoQuerier, input: &[u8], _: u64) -> Result<(ExitSucceed, Vec<u8>), PrecompileFailure> {
        // Expects to receive RLP-encoded Verifiable Credential with JWT proof
        // should contain [issuer_id, credential_subject_id, credential_subject_address, proof]
        let rlp_raw = Rlp::new(input);
        let _decoded_values: Vec<std::string::String> = match rlp_raw.as_list() {
            Ok(res) => res,
            Err(_) => {
                return Err(PrecompileFailure::Error {
                    exit_status: ExitError::Other("cannot decode provided RLP bytes".into()),
                })
            }
        };

        Ok((ExitSucceed::Returned, input.to_vec()))
    }
}