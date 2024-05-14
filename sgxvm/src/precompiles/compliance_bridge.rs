extern crate sgx_tstd as std;

use ethabi::{encode, Address, ParamType, Token as AbiToken, Token};
use evm::executor::stack::{PrecompileHandle, PrecompileOutput};
use evm::{ExitError, ExitRevert};
use primitive_types::H160;
use std::prelude::v1::*;
use std::vec::Vec;

use crate::precompiles::{
    ExitSucceed, LinearCostPrecompileWithQuerier, PrecompileFailure, PrecompileResult,
};
use crate::protobuf_generated::ffi;
use crate::{coder, querier, GoQuerier};

// Selector of addVerificationDetails function
const ADD_VERIFICATION_FN_SELECTOR: &str = "8812b27d";
// Selector of hasVerification function
const HAS_VERIFICATION_FN_SELECTOR: &str = "4887fcd8";
// Selector of getVerificationData function
const GET_VERIFICATION_DATA_FN_SELECTOR: &str = "cc8995ec";

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
        let (exit_status, output) = route(querier, context.caller, handle.input())?;
        Ok(PrecompileOutput {
            exit_status,
            output,
        })
    }
}

fn route(
    querier: *mut GoQuerier,
    caller: H160,
    data: &[u8],
) -> Result<(ExitSucceed, Vec<u8>), PrecompileFailure> {
    if data.len() <= 4 {
        return Err(PrecompileFailure::Revert {
            exit_status: ExitRevert::Reverted,
            output: encode(&vec![AbiToken::String("cannot decode input".into())]),
        });
    }

    let input_signature = hex::encode(data[..4].to_vec());
    match input_signature.as_str() {
        HAS_VERIFICATION_FN_SELECTOR => {
            let has_verification_params = vec![
                ParamType::Address,
                ParamType::Uint(32),
                ParamType::Uint(32),
                ParamType::Array(Box::new(ParamType::Address)),
            ];
            let decoded_params = decode_input(has_verification_params, &data[4..])?;

            let user_address = &decoded_params[0];
            let verification_type = &decoded_params[1];
            let expiration_timestamp = &decoded_params[2];
            let allowed_issuers = &decoded_params[3];

            // Decode allowed issuers
            let allowed_issuers = allowed_issuers.clone().into_array().unwrap();
            let allowed_issuers: Vec<Address> = allowed_issuers
                .into_iter()
                .map(|issuer| issuer.into_address().unwrap())
                .collect();

            let encoded_request = coder::encode_has_verification_request(
                user_address.clone().into_address().unwrap(),
                verification_type.clone().into_uint().unwrap().as_u32(),
                expiration_timestamp.clone().into_uint().unwrap().as_u32(),
                allowed_issuers,
            );

            match querier::make_request(querier, encoded_request) {
                Some(result) => {
                    let has_verification = protobuf::parse_from_bytes::<
                        ffi::QueryHasVerificationResponse,
                    >(result.as_slice())
                    .map_err(|_| PrecompileFailure::Revert {
                        exit_status: ExitRevert::Reverted,
                        output: encode(&vec![AbiToken::String(
                            "cannot decode protobuf response".into(),
                        )]),
                    })?;

                    let tokens = vec![AbiToken::Bool(has_verification.hasVerification)];

                    let encoded_response = encode(&tokens);
                    return Ok((ExitSucceed::Returned, encoded_response.to_vec()));
                }
                None => {
                    return Err(PrecompileFailure::Revert {
                        exit_status: ExitRevert::Reverted,
                        output: encode(&vec![AbiToken::String(
                            "call to x/compliance failed".into(),
                        )]),
                    })
                }
            }
        }
        ADD_VERIFICATION_FN_SELECTOR => {
            let verification_params = vec![
                ParamType::Address,
                ParamType::Uint(32),
                ParamType::Uint(32),
                ParamType::Uint(32),
                ParamType::Bytes,
                ParamType::String,
                ParamType::String,
                ParamType::Uint(32),
            ];
            let decoded_params = decode_input(verification_params, &data[4..])?;

            let user_address = &decoded_params[0];
            let verification_type = &decoded_params[1];
            let issuance_timestamp = &decoded_params[2];
            let expiration_timestamp = &decoded_params[3];
            let proof_data = &decoded_params[4];
            let schema = &decoded_params[5];
            let issuer_verification_id = &decoded_params[6];
            let version = &decoded_params[7];

            let encoded_request = coder::encode_add_verification_details_request(
                user_address.clone().into_address().unwrap(),
                caller,
                verification_type.clone().into_uint().unwrap().as_u32(),
                issuance_timestamp.clone().into_uint().unwrap().as_u32(),
                expiration_timestamp.clone().into_uint().unwrap().as_u32(),
                proof_data.clone().into_bytes().unwrap(),
                schema.clone().into_string().unwrap(),
                issuer_verification_id.clone().into_string().unwrap(),
                version.clone().into_uint().unwrap().as_u32(),
            );

            match querier::make_request(querier, encoded_request) {
                Some(result) => {
                    let _ = protobuf::parse_from_bytes::<ffi::QueryAddVerificationDetailsResponse>(
                        result.as_slice(),
                    )
                    .map_err(|_| PrecompileFailure::Revert {
                        exit_status: ExitRevert::Reverted,
                        output: encode(&vec![AbiToken::String(
                            "cannot parse protobuf response".into(),
                        )]),
                    })?;

                    Ok((ExitSucceed::Returned, Vec::default()))
                }
                None => {
                    return Err(PrecompileFailure::Revert {
                        exit_status: ExitRevert::Reverted,
                        output: encode(&vec![AbiToken::String(
                            "call to x/compliance failed".into(),
                        )]),
                    })
                }
            }
        }
        GET_VERIFICATION_DATA_FN_SELECTOR => {
            let get_verification_data_params = vec![ParamType::Address];
            let decoded_params = decode_input(get_verification_data_params, &data[4..])?;

            let user_address = &decoded_params[0];

            let encoded_request = coder::encode_get_verification_data(
                user_address.clone().into_address().unwrap(),
                caller,
            );

            match querier::make_request(querier, encoded_request) {
                Some(result) => {
                    let get_verification_data = protobuf::parse_from_bytes::<
                        ffi::QueryGetVerificationDataResponse,
                    >(result.as_slice())
                    .map_err(|_| PrecompileFailure::Revert {
                        exit_status: ExitRevert::Reverted,
                        output: encode(&vec![AbiToken::String(
                            "cannot decode protobuf response".into(),
                        )]),
                    })?;

                    let data = get_verification_data
                        .data
                        .into_iter()
                        .map(|log| {
                            let issuer_address = Address::from_slice(&log.issuerAddress);
                            let origin_chain =
                                String::from_utf8_lossy(&log.originChain).to_string(); // Convert bytes to string if required

                            let tokens = vec![
                                AbiToken::Address(issuer_address),
                                AbiToken::String(origin_chain),
                                AbiToken::Uint(log.issuanceTimestamp.into()),
                                AbiToken::Uint(log.expirationTimestamp.into()),
                                AbiToken::Bytes(log.originalData),
                                AbiToken::Bytes(log.schema),
                                AbiToken::Bytes(log.issuerVerificationId),
                                AbiToken::Uint(log.version.into()),
                            ];

                            tokens
                        })
                        .flatten()
                        .collect::<Vec<Token>>(); // Flatten the nested vectors and collect them into a single vector

                    let encoded_response = encode(&data);
                    return Ok((ExitSucceed::Returned, encoded_response.to_vec()));
                }
                None => {
                    return Err(PrecompileFailure::Revert {
                        exit_status: ExitRevert::Reverted,
                        output: encode(&vec![AbiToken::String(
                            "call to x/compliance failed".into(),
                        )]),
                    })
                }
            }
        }
        _ => Err(PrecompileFailure::Revert {
            exit_status: ExitRevert::Reverted,
            output: encode(&vec![AbiToken::String("incorrect request".into())]),
        }),
    }
}

fn decode_input(
    param_types: Vec<ParamType>,
    input: &[u8],
) -> Result<Vec<Token>, PrecompileFailure> {
    let decoded_params =
        ethabi::decode(&param_types, input).map_err(|err| PrecompileFailure::Revert {
            exit_status: ExitRevert::Reverted,
            output: encode(&vec![AbiToken::String(
                format!("cannot decode params: {:?}", err).into(),
            )]),
        })?;

    if decoded_params.len() != param_types.len() {
        return Err(PrecompileFailure::Revert {
            exit_status: ExitRevert::Reverted,
            output: encode(&vec![AbiToken::String(
                "incorrect decoded params len".into(),
            )]),
        });
    }

    Ok(decoded_params)
}
