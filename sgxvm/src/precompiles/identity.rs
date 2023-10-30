extern crate sgx_tstd as std;

use crate::precompiles::{
    ExitError, ExitSucceed, LinearCostPrecompileWithQuerier, PrecompileFailure, PrecompileResult,
};
use crate::{
    GoQuerier,
    coder,
    ocall,
    protobuf_generated::ffi,
};
use ed25519_dalek::{Signature, Verifier, VerifyingKey};
use evm::executor::stack::{PrecompileHandle, PrecompileOutput};
use serde::Deserialize;
use std::{
    string::{String, ToString},
    vec::Vec,
};
use bech32::FromBase32;

#[derive(Deserialize)]
struct Header {
    pub alg: String,
    pub typ: String,
}

impl Header {
    fn validate(&self) -> Result<(), PrecompileFailure> {
        match (self.typ.as_str(), self.alg.as_str()) {
            ("JWT", "EdDSA") => Ok(()),
            _ => Err(PrecompileFailure::Error {
                exit_status: ExitError::Other("Invalid JWT header".into()),
            }),
        }
    } 
}

#[derive(Deserialize)]
struct VerifiableCredential {
    vc: VC,
    sub: String,
    nbf: i64,
    iss: String,
}

#[derive(Deserialize)]
struct VC {
    #[serde(rename = "@context")]
    context: Vec<String>,
    #[serde(rename = "type")]
    vc_type: Vec<String>,
    #[serde(rename = "credentialSubject")]
    credential_subject: CredentialSubject,
}

#[derive(Deserialize)]
struct CredentialSubject {
    #[serde(alias = "address")]
    user_address: String,
}

/// The identity precompile.
pub struct Identity;

impl LinearCostPrecompileWithQuerier for Identity {
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
        
        // Extract issuer from payload and obtain verification material
        let verification_materials = get_verification_material(querier, parsed_payload.iss)?;

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
        Ok((ExitSucceed::Returned, credential_subject))
    }
}

/// Splits provided JWT into header, payload, signature and data.
/// Data field contains concatenated header and payload and can be used for signature verification
fn split_jwt(jwt: &str) -> Result<(String, String, String, String), PrecompileFailure> {
    let parts: Vec<&str> = jwt.split('.').collect();

    if parts.len() != 3 {
        return Err(PrecompileFailure::Error {
            exit_status: ExitError::Other("Wrong amount of parts in JWT".into()),
        });
    }

    let header = String::from_utf8(base64_decode(parts[0])).map_err(|_| PrecompileFailure::Error {
        exit_status: ExitError::Other("Cannot decode JWT header to utf-8".into()),
    })?;

    let payload = String::from_utf8(base64_decode(parts[1])).map_err(|_| PrecompileFailure::Error {
        exit_status: ExitError::Other("Cannot decode JWT payload to utf-8".into()),
    })?;

    let signature = parts[2].to_string();
    let data = format!("{}.{}", parts[0], parts[1]);

    Ok((header, payload, signature, data))
}

/// Verifies provided ed25519 signature
fn verify_signature(data: &str, signature: &str, vm: &str) -> Result<(), PrecompileFailure> {
    // Construct signature
    let signature = base64_decode(signature);
    let signature = Signature::from_slice(&signature).map_err(|err| {
        PrecompileFailure::Error {
            exit_status: ExitError::Other("Cannot construct signature".into()),
        }
    })?;

    let public_key = multibase_to_vec(vm)?;

    let public_key: &[u8; 32] = public_key.as_slice().try_into().map_err(|err| {
        PrecompileFailure::Error {
            exit_status: ExitError::Other("Cannot convert public key to fixed bytes array".into()),
        }
    })?;

    let verification_key = VerifyingKey::from_bytes(public_key).map_err(|err| {
        PrecompileFailure::Error {
            exit_status: ExitError::Other("Cannot construct verification key".into()),
        }
    })?;

    // Verify signature
    verification_key
        .verify(data.as_bytes(), &signature)
        .map_err(|_| PrecompileFailure::Error {
            exit_status: ExitError::Other("Signature verification failed".into()),
        })?;

    Ok(())    
}

fn convert_bech32_address(address: String) -> Result<Vec<u8>, PrecompileFailure> {
    // If address is 0x-prefixed we treat it as ethereum-like address
    if address.starts_with("0x") {
        return hex::decode(&address[2..]).map_err(|_| PrecompileFailure::Error {
            exit_status: ExitError::Other("Cannot decode address".into()),
        });
    }

    let (_, data, _) =
        bech32::decode(address.as_str()).map_err(|_| PrecompileFailure::Error {
            exit_status: ExitError::Other("Cannot decode bech32 address".into()),
        })?;
    let data =
        Vec::<u8>::from_base32(&data).map_err(|_| PrecompileFailure::Error {
            exit_status: ExitError::Other("Cannot convert base32 to bytes".into()),
        })?;
    Ok(data)
}

fn base64_decode(input: &str) -> Vec<u8> {
    base64::decode_config(&input, base64::URL_SAFE).unwrap_or_default()
}

/// Makes `OCALL` to obtain verification methods from DID registry
/// * connector – pointer to GoQuerier, which will be used to make queries to `x/did` module
/// * did_url – url to DID document
fn get_verification_material(connector: *mut GoQuerier, did_url: String) -> Result<Vec<ffi::VerificationMethod>, PrecompileFailure> {
    let encoded_request = coder::encode_verification_methods_request(did_url);
    match ocall::make_request(connector, encoded_request) {
        Some(result) => {
            // Decode protobuf and extract verification methods
            protobuf::parse_from_bytes::<ffi::QueryVerificationMethodsResponse>(result.as_slice())
                .map_err(|_| PrecompileFailure::Error { exit_status: ExitError::Other("Cannot decode protobuf response".into()) })
                .and_then(|decoded_result| Ok(decoded_result.vm.to_vec()))
        },
        None => {
            return Err(PrecompileFailure::Error {
                exit_status: ExitError::Other("Cannot obtain verification material".into()),
            })
        }
    }
}

/// Decodes multibase encoded verification material to bytes
fn multibase_to_vec(value: &str) -> Result<Vec<u8>, PrecompileFailure> {
    let (_, decoded_data) = multibase::decode(value)
        .map_err(|_| PrecompileFailure::Error { 
            exit_status: ExitError::Other("Cannot decode multibase".into()) 
    })?;

    Ok(decoded_data[2..].to_vec())
}