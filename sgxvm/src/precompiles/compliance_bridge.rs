extern crate sgx_tstd as std;

use ethabi::{encode, Address, ParamType, Token as AbiToken, Token};
use evm::GasMutState;
use evm::interpreter::error::{ExitError, ExitResult, ExitSucceed};
use evm::interpreter::runtime::RuntimeState;
use primitive_types::{H160, U256};
use std::prelude::v1::*;
use std::vec::Vec;

use crate::precompiles::LinearCostPrecompileWithQuerier;
use crate::{coder, querier, GoQuerier};
use crate::protobuf_generated::ffi::{QueryAddVerificationDetailsResponse, QueryAddVerificationDetailsV2, QueryAddVerificationDetailsV2Response, QueryGetVerificationDataResponse, QueryHasVerificationResponse, QueryIssuanceTreeRoot, QueryIssuanceTreeRootResponse, QueryRevocationTreeRootResponse, QueryRevokeVerification, QueryRevokeVerificationResponse};

// Selector of `addVerificationDetails` function
const ADD_VERIFICATION_FN_SELECTOR: &str = "e62364ab";
// Selector of `addVerificationDetailsV2` function
const ADD_VERIFICATION_V2_FN_SELECTOR: &str = "c2206580";
// Selector of `hasVerification` function
const HAS_VERIFICATION_FN_SELECTOR: &str = "4887fcd8";
// Selector of `getVerificationData` function
const GET_VERIFICATION_DATA_FN_SELECTOR: &str = "cc8995ec";
// Selector of `getRevocationTreeRoot` function
const GET_REVOCATION_TREE_ROOT_FN_SELECTOR: &str = "3db94a04";
// Selector of `getIssuanceTreeRoot` function
const GET_ISSUANCE_TREE_ROOT_FN_SELECTOR: &str = "d0376bd2";
const REVOKE_VERIFICATION_FN_SELECTOR: &str = "e711d86d";
const CONVERT_CREDENTIAL_FN_SELECTOR: &str = "0x460c4841";

/// Precompile for interactions with x/compliance module.
pub struct ComplianceBridge;

impl<G: AsRef<RuntimeState> + GasMutState> LinearCostPrecompileWithQuerier<G> for ComplianceBridge {
    const BASE: u64 = 60;
    const WORD: u64 = 150;

    fn execute(querier: *mut GoQuerier, input: &[u8], gasometer: &mut G) -> (ExitResult, Vec<u8>) {
        // For some reason, rust compiler cannot infer type for BASE and WORD consts,
        // therefore their values provided directly
        let cost = match static_precompiles::linear_cost(input.len() as u64, 60, 150) {
            Ok(cost) => cost,
            Err(e) => return (e.into(), Vec::new()),
        };

        if let Err(e) = gasometer.record_gas(cost.into()) {
            return (e.into(), Vec::new());
        }

        let d = gasometer.as_ref();
        route(querier, d.context.caller, input)
    }
}

fn route(
    querier: *mut GoQuerier,
    caller: H160,
    data: &[u8],
) -> (ExitResult, Vec<u8>) {
    if data.len() < 4 {
        return (ExitError::Reverted.into(), encode(&[AbiToken::String("cannot decode input".into())]));
    }

    let input_signature = hex::encode(data[..4].to_vec());
    match input_signature.as_str() {
        REVOKE_VERIFICATION_FN_SELECTOR => {
            let revoke_verification_params = vec![
                ParamType::Bytes,
            ];

            let decoded_params = match decode_input(revoke_verification_params, &data[4..]) {
                Ok(params) => params,
                Err(_) => return (ExitError::Reverted.into(), encode(&vec![AbiToken::String("failed to decode input parameters".into())]))
            };

            let verification_id = match decoded_params[0].clone().into_bytes() {
                Some(id) => id,
                None => return (ExitError::Reverted.into(), encode(&vec![AbiToken::String("cannot parse verification id".into())]))
            };

            let encoded_request = coder::encode_revoke_verification(verification_id, &caller);
            match querier::make_request(querier, encoded_request) {
                Some(result) => {
                    let _: QueryRevokeVerificationResponse = match protobuf::parse_from_bytes(&result) {
                        Ok(response) => response,
                        Err(_) => return (ExitError::Reverted.into(), encode(&[AbiToken::String("cannot decode protobuf response".into())]))
                    };

                    (ExitSucceed::Returned.into(), Vec::new())
                }
                None => (ExitError::Reverted.into(), encode(&[AbiToken::String("call to revokeVerification function to x/compliance failed".into())]))
            }
        }
        GET_REVOCATION_TREE_ROOT_FN_SELECTOR => {
            let encoded_request = coder::encode_get_revocation_tree_root_request();
            match querier::make_request(querier, encoded_request) {
                Some(result) => {
                    let res: QueryRevocationTreeRootResponse = match protobuf::parse_from_bytes(&result) {
                        Ok(response) => response,
                        Err(_) => return (ExitError::Reverted.into(), encode(&[AbiToken::String("cannot decode protobuf response".into())]))
                    };

                    let value = U256::from_big_endian(&res.root);
                    let tokens = vec![AbiToken::Uint(value)];

                    let encoded_response = encode(&tokens);
                    (ExitSucceed::Returned.into(), encoded_response.to_vec())
                }
                None => (ExitError::Reverted.into(), encode(&[AbiToken::String("call to getRevocationTreeRoot function to x/compliance failed".into())]))
            }
        }
        GET_ISSUANCE_TREE_ROOT_FN_SELECTOR => {
            let encoded_request = coder::encode_get_issuance_tree_root_request();
            match querier::make_request(querier, encoded_request) {
                Some(result) => {
                    let res: QueryIssuanceTreeRootResponse = match protobuf::parse_from_bytes(&result) {
                        Ok(response) => response,
                        Err(_) => return (ExitError::Reverted.into(), encode(&[AbiToken::String("cannot decode protobuf response".into())]))
                    };

                    let value = U256::from_big_endian(&res.root);
                    let tokens = vec![AbiToken::Uint(value)];

                    let encoded_response = encode(&tokens);
                    (ExitSucceed::Returned.into(), encoded_response.to_vec())
                }
                None => (ExitError::Reverted.into(), encode(&[AbiToken::String("call to getIssuanceTreeRoot function to x/compliance failed".into())]))
            }
        }
        HAS_VERIFICATION_FN_SELECTOR => {
            let has_verification_params = vec![
                ParamType::Address,
                ParamType::Uint(32),
                ParamType::Uint(32),
                ParamType::Array(Box::new(ParamType::Address)),
            ];

            let decoded_params = match decode_input(has_verification_params, &data[4..]) {
                Ok(params) => params,
                Err(_) => return (ExitError::Reverted.into(), encode(&vec![AbiToken::String("failed to decode input parameters".into())]))
            };

            let user_address = match decoded_params[0].clone().into_address() {
                Some(addr) => addr,
                None => return (ExitError::Reverted.into(), encode(&vec![AbiToken::String("invalid user address".into())]))
            };

            let verification_type = match decoded_params[1].clone().into_uint() {
                Some(vtype) => vtype.as_u32(),
                None => return (ExitError::Reverted.into(), encode(&[AbiToken::String("invalid verification type".into())]))
            };

            let expiration_timestamp = match decoded_params[2].clone().into_uint() {
                Some(timestamp) => timestamp.as_u32(),
                None => return (ExitError::Reverted.into(), encode(&[AbiToken::String("invalid expiration timestamp".into())]))
            };

            let allowed_issuers = match decoded_params[3].clone().into_array() {
                Some(array) => array,
                None => return (ExitError::Reverted.into(), encode(&[AbiToken::String("invalid allowed issuers array".into())]))
            };

            // Decode allowed issuers
            let allowed_issuers: Result<Vec<Address>, _> = allowed_issuers
                .into_iter()
                .map(|issuer| match issuer.into_address() {
                    Some(address) => Ok(address),
                    None => Err(()),
                })
                .collect();

            let allowed_issuers = match allowed_issuers {
                Ok(issuers) => issuers,
                Err(_) => {
                    return (ExitError::Reverted.into(), encode(&[AbiToken::String("one or more invalid issuer addresses".into())]))
                }
            };

            let encoded_request = coder::encode_has_verification_request(
                user_address,
                verification_type,
                expiration_timestamp,
                allowed_issuers,
            );

            match querier::make_request(querier, encoded_request) {
                Some(result) => {
                    let has_verification: QueryHasVerificationResponse = match protobuf::parse_from_bytes(&result) {
                        Ok(response) => response,
                        Err(_) => return (ExitError::Reverted.into(), encode(&[AbiToken::String("cannot decode protobuf response".into())]))
                    };

                    let tokens = vec![AbiToken::Bool(has_verification.hasVerification)];

                    let encoded_response = encode(&tokens);
                    (ExitSucceed::Returned.into(), encoded_response.to_vec())
                }
                None => (ExitError::Reverted.into(), encode(&[AbiToken::String("call to hasVerification function to x/compliance failed".into())]))
            }
        }
        ADD_VERIFICATION_FN_SELECTOR => {
            let verification_params = vec![
                ParamType::Address,
                ParamType::String,
                ParamType::Uint(32),
                ParamType::Uint(32),
                ParamType::Uint(32),
                ParamType::Bytes,
                ParamType::String,
                ParamType::String,
                ParamType::Uint(32),
            ];

            let decoded_params = match decode_input(verification_params, &data[4..]) {
                Ok(params) => params,
                Err(_) => return (ExitError::Reverted.into(), encode(&[AbiToken::String("failed to decode input parameters".into())]))
            };

            let user_address = match decoded_params[0].clone().into_address() {
                Some(addr) => addr,
                None => return (ExitError::Reverted.into(), encode(&[AbiToken::String("invalid user address".into())]))
            };

            let origin_chain = match decoded_params[1].clone().into_string() {
                Some(chain) => chain,
                None => return (ExitError::Reverted.into(), encode(&[AbiToken::String("invalid origin chain".into())]))
            };

            let verification_type = match decoded_params[2].clone().into_uint() {
                Some(vtype) => vtype.as_u32(),
                None => return (ExitError::Reverted.into(), encode(&[AbiToken::String("invalid verification type".into())]))
            };

            let issuance_timestamp = match decoded_params[3].clone().into_uint() {
                Some(timestamp) => timestamp.as_u32(),
                None => return (ExitError::Reverted.into(), encode(&[AbiToken::String("invalid issuance timestamp".into())]))
            };

            let expiration_timestamp = match decoded_params[4].clone().into_uint() {
                Some(timestamp) => timestamp.as_u32(),
                None => return (ExitError::Reverted.into(), encode(&[AbiToken::String("invalid expiration timestamp".into())]))
            };

            let proof_data = match decoded_params[5].clone().into_bytes() {
                Some(data) => data,
                None => return (ExitError::Reverted.into(), encode(&[AbiToken::String("invalid proof data".into())]))
            };

            let schema = match decoded_params[6].clone().into_string() {
                Some(schema) => schema,
                None => {
                    return (
                        ExitError::Reverted.into(), encode(&[AbiToken::String("invalid schema".into())]),
                    );
                }
            };

            let issuer_verification_id = match decoded_params[7].clone().into_string() {
                Some(id) => id,
                None => return (ExitError::Reverted.into(), encode(&[AbiToken::String("invalid issuer verification ID".into())]))
            };

            let version = match decoded_params[8].clone().into_uint() {
                Some(ver) => ver.as_u32(),
                None => return (
                    ExitError::Reverted.into(), encode(&[AbiToken::String(
                        "invalid version".into(),
                    )])
                )
            };

            let encoded_request = coder::encode_add_verification_details_request(
                user_address,
                caller,
                origin_chain,
                verification_type,
                issuance_timestamp,
                expiration_timestamp,
                proof_data,
                schema,
                issuer_verification_id,
                version,
            );

            match querier::make_request(querier, encoded_request) {
                Some(result) => {
                    let added_verification: QueryAddVerificationDetailsResponse = match protobuf::parse_from_bytes(result.as_slice()) {
                        Ok(response) => response,
                        Err(_) => return (ExitError::Reverted.into(), encode(&[AbiToken::String("cannot parse protobuf response".into())]))
                    };

                    let token = vec![AbiToken::Bytes(
                        added_verification.verificationId.into(),
                    )];
                    let encoded_response = encode(&token);

                    (ExitSucceed::Returned.into(), encoded_response.to_vec())
                },
                None => (ExitError::Reverted.into(), encode(&[AbiToken::String("call to addVerificationDetails to x/compliance failed".into())]))
            }
        },
        ADD_VERIFICATION_V2_FN_SELECTOR => {
            let verification_params = vec![
                ParamType::Address,
                ParamType::String,
                ParamType::Uint(32),
                ParamType::Uint(32),
                ParamType::Uint(32),
                ParamType::Bytes,
                ParamType::String,
                ParamType::String,
                ParamType::Uint(32),
                ParamType::FixedBytes(32),
            ];

            let decoded_params = match decode_input(verification_params, &data[4..]) {
                Ok(params) => params,
                Err(_) => return (ExitError::Reverted.into(), encode(&[AbiToken::String("failed to decode input parameters".into())]))
            };

            let user_address = match decoded_params[0].clone().into_address() {
                Some(addr) => addr,
                None => return (ExitError::Reverted.into(), encode(&[AbiToken::String("invalid user address".into())]))
            };

            let origin_chain = match decoded_params[1].clone().into_string() {
                Some(chain) => chain,
                None => return (ExitError::Reverted.into(), encode(&[AbiToken::String("invalid origin chain".into())]))
            };

            let verification_type = match decoded_params[2].clone().into_uint() {
                Some(vtype) => vtype.as_u32(),
                None => return (ExitError::Reverted.into(), encode(&[AbiToken::String("invalid verification type".into())]))
            };

            let issuance_timestamp = match decoded_params[3].clone().into_uint() {
                Some(timestamp) => timestamp.as_u32(),
                None => return (ExitError::Reverted.into(), encode(&[AbiToken::String("invalid issuance timestamp".into())]))
            };

            let expiration_timestamp = match decoded_params[4].clone().into_uint() {
                Some(timestamp) => timestamp.as_u32(),
                None => return (ExitError::Reverted.into(), encode(&[AbiToken::String("invalid expiration timestamp".into())]))
            };

            let proof_data = match decoded_params[5].clone().into_bytes() {
                Some(data) => data,
                None => return (ExitError::Reverted.into(), encode(&[AbiToken::String("invalid proof data".into())]))
            };

            let schema = match decoded_params[6].clone().into_string() {
                Some(schema) => schema,
                None => {
                    return (
                        ExitError::Reverted.into(), encode(&[AbiToken::String("invalid schema".into())]),
                    );
                }
            };

            let issuer_verification_id = match decoded_params[7].clone().into_string() {
                Some(id) => id,
                None => return (ExitError::Reverted.into(), encode(&[AbiToken::String("invalid issuer verification ID".into())]))
            };

            let version = match decoded_params[8].clone().into_uint() {
                Some(ver) => ver.as_u32(),
                None => return (
                    ExitError::Reverted.into(), encode(&[AbiToken::String(
                        "invalid version".into(),
                    )])
                )
            };

            let user_public_key = match decoded_params[9].clone().into_fixed_bytes() {
                Some(pk) => pk,
                None => {
                    return (
                        ExitError::Reverted.into(), encode(&[AbiToken::String("invalid user public key".into())]),
                    );
                }
            };

            let encoded_request = coder::encode_add_verification_details_v2_request(
                user_address,
                caller,
                origin_chain,
                verification_type,
                issuance_timestamp,
                expiration_timestamp,
                proof_data,
                schema,
                issuer_verification_id,
                version,
                user_public_key,
            );

            match querier::make_request(querier, encoded_request) {
                Some(result) => {
                    let added_verification: QueryAddVerificationDetailsV2Response = match protobuf::parse_from_bytes(result.as_slice()) {
                        Ok(response) => response,
                        Err(_) => return (ExitError::Reverted.into(), encode(&[AbiToken::String("cannot parse protobuf response".into())]))
                    };

                    let token = vec![AbiToken::Bytes(
                        added_verification.verificationId.into(),
                    )];
                    let encoded_response = encode(&token);

                    (ExitSucceed::Returned.into(), encoded_response.to_vec())
                },
                None => (ExitError::Reverted.into(), encode(&[AbiToken::String("call to addVerificationDetailsV2 to x/compliance failed".into())]))
            }
        },
        GET_VERIFICATION_DATA_FN_SELECTOR => {
            let get_verification_data_params = vec![ParamType::Address, ParamType::Address];
            let decoded_params = match decode_input(get_verification_data_params, &data[4..]) {
                Ok(params) => params,
                Err(_) => {
                    return (
                        ExitError::Reverted.into(),
                        encode(&[AbiToken::String(
                            "failed to decode input parameters".into(),
                        )]),
                    );
                }
            };

            let user_address = match decoded_params[0].clone().into_address() {
                Some(addr) => addr,
                None => {
                    return (
                        ExitError::Reverted.into(),
                        encode(&[AbiToken::String("invalid user address".into())]),
                    );
                }
            };

            let issuer_address = match decoded_params[1].clone().into_address() {
                Some(addr) => addr,
                None => return (ExitError::Reverted.into(), encode(&[AbiToken::String("invalid issuer address".into())]))
            };

            let encoded_request = coder::encode_get_verification_data(user_address, issuer_address);

            match querier::make_request(querier, encoded_request) {
                Some(result) => {
                    let get_verification_data: QueryGetVerificationDataResponse = match protobuf::parse_from_bytes(result.as_slice()) {
                        Ok(response) => response,
                        Err(_) => return (ExitError::Reverted.into(), encode(&[AbiToken::String("cannot decode protobuf response".into())]))
                    };

                    let data = get_verification_data
                        .data
                        .into_iter()
                        .flat_map(|log| {
                            let issuer_address = Address::from_slice(&log.issuerAddress);
                            let tokens = vec![AbiToken::Tuple(vec![
                                AbiToken::Uint(log.verificationType.into()),
                                AbiToken::Bytes(log.verificationID.clone().into()),
                                AbiToken::Address(issuer_address.clone()),
                                AbiToken::String(log.originChain.clone()),
                                AbiToken::Uint(log.issuanceTimestamp.into()),
                                AbiToken::Uint(log.expirationTimestamp.into()),
                                AbiToken::Bytes(log.originalData.clone().into()),
                                AbiToken::String(log.schema.clone()),
                                AbiToken::String(log.issuerVerificationId.clone()),
                                AbiToken::Uint(log.version.into()),
                            ])];

                            tokens.into_iter()
                        })
                        .collect::<Vec<AbiToken>>();

                    let encoded_response = encode(&[AbiToken::Array(data)]);
                    (ExitSucceed::Returned.into(), encoded_response.to_vec())
                },
                None => (ExitError::Reverted.into(), encode(&[AbiToken::String("call to getVerificationData failed to x/compliance failed".into())]))
            }
        },
        _ => (ExitError::Reverted.into(), encode(&vec![AbiToken::String("incorrect request".into())]))
    }
}

fn decode_input(
    param_types: Vec<ParamType>,
    input: &[u8],
) -> Result<Vec<Token>, (ExitError, Vec<u8>)> {
    let decoded_params =
        ethabi::decode(&param_types, input).map_err(|err| (
            ExitError::Reverted.into(),
            encode(&[AbiToken::String(format!("cannot decode params: {:?}", err).into())])
        ))?;

    if decoded_params.len() != param_types.len() {
        return Err((ExitError::Reverted.into(), encode(&[AbiToken::String("incorrect decoded params len".into())])));
    }

    Ok(decoded_params)
}
