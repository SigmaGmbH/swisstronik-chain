extern crate sgx_tstd as std;

use alloc::borrow::ToOwned;
use alloc::string::ToString;
use evm::executor::stack::{PrecompileHandle, PrecompileOutput};
use evm::{ExitError, ExitRevert};
use primitive_types::H160;
use std::prelude::v1::*;
use std::vec::Vec;
use ethabi::{Function, Param, ParamType, Token as AbiToken};

use crate::GoQuerier;
use crate::precompiles::{
    ExitSucceed, LinearCostPrecompileWithQuerier, PrecompileFailure, PrecompileResult,
};

// Selector of addVerificationDetails function
const ADD_VERIFICATION_FN_SELECTOR: &str = "455d0d34";
// Selector of hasVerification function
const HAS_VERIFICATION_FN_SELECTOR: &str = "4887fcd8";

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




        let context = handle.context();
        let (exit_status, output) = route(querier, context.address, handle.input())?;
        Ok(PrecompileOutput {
            exit_status,
            output,
        })
    }
}

fn route(querier: *mut GoQuerier, caller: H160, data: &[u8]) -> Result<(ExitSucceed, Vec<u8>), PrecompileFailure> {
    if data.len() <= 4 {
        return Err(PrecompileFailure::Error {
            exit_status: ExitError::Other("cannot decode input".into()),
        })
    }

    let input_signature = hex::encode(data[..4].to_vec());
    match input_signature.as_str() {
        HAS_VERIFICATION_FN_SELECTOR => {
            let verification_params = vec![ParamType::Address, ParamType::Uint(32), ParamType::Uint(32), ParamType::Uint(32), ParamType::Bytes];
            let decoded = ethabi::decode_whole(&verification_params, &data[4..]).map_err(|err| {
                PrecompileFailure::Error {
                    exit_status: ExitError::Other(format!("cannot decode params: {:?}", err).into()),
                }
            })?;
            // TODO: Implement READ from x/compliance
            Ok((ExitSucceed::Returned, Vec::default()))
        },
        ADD_VERIFICATION_FN_SELECTOR => {
            // TODO: Decode params
            let has_verification_params = vec![ParamType::Address, ParamType::Uint(32), ParamType::Uint(32), ParamType::Array(Box::new(ParamType::Address))];
            let decoded = ethabi::decode_whole(&has_verification_params, &data[4..]).map_err(|err| {
                PrecompileFailure::Error {
                    exit_status: ExitError::Other(format!("cannot decode params: {:?}", err).into()),
                }
            })?;
            // TODO: Implement WRITE to x/compliance
            Ok((ExitSucceed::Returned, Vec::default()))
        },
        _ => {
            Err(PrecompileFailure::Error {
                exit_status: ExitError::Other("cannot decode input".into()),
            })
        }
    }
}