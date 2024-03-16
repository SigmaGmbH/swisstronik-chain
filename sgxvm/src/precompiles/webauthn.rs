extern crate sgx_tstd as std;

use evm::executor::stack::{PrecompileHandle, PrecompileOutput};
use evm::{ExitError};
use primitive_types::H160;
use std::prelude::v1::*;
use std::vec::Vec;
use ethabi::{Address, encode, ParamType, Token as AbiToken};

use crate::{coder, GoQuerier, querier};
use crate::precompiles::{
    ExitSucceed, LinearCostPrecompileWithQuerier, PrecompileFailure, PrecompileResult,
};
use crate::protobuf_generated::ffi;

const BEGIN_REGISTRATION_FN_SELECTOR: &str = "455d0d34";
const FINISH_REGISTRATION_FN_SELECTOR: &str = "4887fcd8";
const BEGIN_LOGIN_FN_SELECTOR: &str = "455d0d34";
const FINISH_LOGIN_FN_SELECTOR: &str = "4887fcd8";

/// Precompile for WebauthN (Passkeys) authentication
pub struct WebAuthn;

impl LinearCostPrecompileWithQuerier for WebAuthn {
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

fn begin_registration_handler(querier: *mut GoQuerier, existing_credentials:Vec<u8>, user_id: H160){}

fn route(querier: *mut GoQuerier, caller: H160, data: &[u8]) -> Result<(ExitSucceed, Vec<u8>), PrecompileFailure> {
    if data.len() <= 4 {
        return Err(PrecompileFailure::Error {
            exit_status: ExitError::Other("cannot decode input".into()),
        })
    }

    let input_signature = hex::encode(data[..4].to_vec());
    match input_signature.as_str() {
        BEGIN_REGISTRATION_FN_SELECTOR => {
            return Err(PrecompileFailure::Error {
                exit_status: ExitError::Other("Cannot obtain verification material".into()),
            })
        },
        FINISH_REGISTRATION_FN_SELECTOR => {
            return Err(PrecompileFailure::Error {
                exit_status: ExitError::Other("Cannot obtain verification material".into()),
            })
        },
        BEGIN_LOGIN_FN_SELECTOR => {
            return Err(PrecompileFailure::Error {
                exit_status: ExitError::Other("Cannot obtain verification material".into()),
            })
        },
        FINISH_LOGIN_FN_SELECTOR => {
            return Err(PrecompileFailure::Error {
                exit_status: ExitError::Other("Cannot obtain verification material".into()),
            })
        },
        _ => {
            Err(PrecompileFailure::Error {
                exit_status: ExitError::Other("cannot decode input".into()),
            })
        }
    }
}