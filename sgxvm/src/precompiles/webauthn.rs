extern crate sgx_tstd as std;

use crate::precompiles::{
    ExitError, ExitSucceed, LinearCostPrecompileWithQuerier, PrecompileFailure, PrecompileResult,
};
use crate::{
    GoQuerier,
    coder,
    protobuf_generated::ffi,
    querier,
};
use ed25519_dalek::{Signature, Verifier, VerifyingKey};
use evm::executor::stack::{PrecompileHandle, PrecompileOutput};
use ethabi::{Token as AbiToken, encode as encodeAbi};
use primitive_types::H160;
use serde::Deserialize;
use std::{
    string::{String, ToString},
    vec::Vec,
};
use bech32::FromBase32;

use std::prelude::v1::*;
use std::time::*;
use std::untrusted::time::SystemTimeEx;
use chrono::TimeZone;
use chrono::Utc as TzUtc;

/// The webauthn precompile.
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
        let (exit_status, output) = Self::raw_execute(querier, handle.input(), cost)?;
        Ok(PrecompileOutput {
            exit_status,
            output,
        })
    }

    fn raw_execute(
        querier: *mut GoQuerier,
        input: &[u8],
        _: u64,
    ) -> Result<(ExitSucceed, Vec<u8>), PrecompileFailure> {
        // Expects to receive RLP-encoded JWT proof for Verifiable Credential
        let jwt: String = rlp::decode(input).map_err(|_| PrecompileFailure::Error {
            exit_status: ExitError::Other("cannot decode provided JWT proof".into()),
        })?;

        // Split JWT into parts
        let (header, payload, signature, data) = split_jwt(jwt.as_str())?;

        // Parse and validate header
        let header: Header = serde_json::from_str(header.as_str()).map_err(|_| PrecompileFailure::Error {
            exit_status: ExitError::Other("Cannot parse JWT header".into()),
        })?;

        // Validate header
        header.validate()?;

        // Parse payload
        let parsed_payload: VerifiableCredential = serde_json::from_str(payload.as_str()).map_err(|_| PrecompileFailure::Error {
            exit_status: ExitError::Other("Cannot parse JWT payload".into()),
        })?;

        // Since we issue VC without expiration date, verify nbf (not valid before) field of JWT, it should be less than current timestamp 
        validate_nbf(parsed_payload.nbf)?;

        // Extract issuer from payload and obtain verification material
        let verification_materials = get_verification_material(querier, parsed_payload.iss.clone())?;

        // Find appropriate verification material
        let vm = verification_materials
            .iter()
            .find(|verification_method| verification_method.verificationMethodType == "Ed25519VerificationKey2020" || verification_method.verificationMethodType == "Ed25519VerificationKey2018")
            .and_then(|method| Some(method.verificationMaterial.clone()))
            .ok_or(PrecompileFailure::Error {
                exit_status: ExitError::Other("Cannot find appropriate verification method".into()),
            })?;
        
        verify_signature(&data, &signature, &vm)?;

        let credential_subject = convert_bech32_address(parsed_payload.vc.credential_subject.user_address)?;
        let output = encode_output(credential_subject, parsed_payload.iss);

        Ok((ExitSucceed::Returned, output))
    }
}
