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
use thiserror_no_std::Error;
use multibase::{Base, decode};

#[derive(Debug, Deserialize)]
struct Header {
    pub alg: String,
    pub typ: String,
}

impl Header {
    fn validate(&self) -> Result<(), PrecompileFailure> {
        if self.typ != "JWT" {
            return Err(PrecompileFailure::Error {
                exit_status: ExitError::Other("Wrong header type. Expected JWT".into()),
            })
        }
    
        if self.alg != "EdDSA" {
            return Err(PrecompileFailure::Error {
                exit_status: ExitError::Other("Wrong algorithm. Expected EdDSA".into()),
            })
        }
    
        Ok(())
    } 
}

#[derive(Debug, Deserialize)]
struct VerifiableCredential {
    vc: VC,
    sub: String,
    nbf: i64,
    iss: String,
}

#[derive(Debug, Deserialize)]
struct VC {
    #[serde(rename = "@context")]
    context: Vec<String>,
    #[serde(rename = "type")]
    vc_type: Vec<String>,
    #[serde(rename = "credentialSubject")]
    credential_subject: CredentialSubject,
}

#[derive(Debug, Deserialize)]
struct CredentialSubject {
    address: String,
}

struct VerificationMethod {
    vm: String,
    vm_type: String,
}

#[derive(Error, Debug)]
pub enum VerificationError {
    #[error("Cannot split JWT: {}", msg)]
    CannotSplitJWT { msg: String },
    #[error("Header verification failed: {}", msg)]
    HeaderVerificationError { msg: String },
    #[error("Cannot parse JSON: {}", msg)]
    JSONParseError { msg: String },
    #[error("Signature verification failed: {}", msg)]
    SignatureVerificationError { msg: String },
    #[error("Cannot convert address: {}", msg)]
    ConvertAddressError { msg: String },
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
        let jwt: String = match rlp::decode(input) {
            Ok(res) => res,
            Err(_) => {
                return Err(PrecompileFailure::Error {
                    exit_status: ExitError::Other("cannot decode provided JWT proof".into()),
                })
            }
        };

        // Split JWT into parts
        let (header, payload, signature, data) = split_jwt(jwt.as_str())?;

        // Parse header
        let header: Header = match serde_json::from_str(header.as_str()) {
            Ok(header) => header,
            Err(err) => {
                return Err(PrecompileFailure::Error {
                    exit_status: ExitError::Other("Cannot parse JWT header".into()),
                })
            }
        };

        // Validate header
        header.validate()?;

        // Parse payload
        let parsed_payload: VerifiableCredential = match serde_json::from_str(payload.as_str()) {
            Ok(res) => res,
            Err(e) => {
                return Err(PrecompileFailure::Error {
                    exit_status: ExitError::Other("Cannot parse JWT payload".into()),
                })
            }
        };
        
        // Extract issuer from payload and obtain verification material
        let issuer = parsed_payload.iss;
        let verification_materials = get_verification_material(querier, issuer)?;
        println!("DEBUG: received VMs: {:?}", verification_materials);

        // Find appropriate verification material
        let vm = verification_materials.iter().find(|verification_method| verification_method.verificationMethodType == "Ed25519VerificationKey2020" || verification_method.verificationMethodType == "Ed25519VerificationKey2018");
        println!("DEBUG: found vms: {:?}", vm); // TODO: Remove debug log
        let vm = match vm {
            Some(method) => method.verificationMaterial.clone(),
            None => {
                return Err(PrecompileFailure::Error {
                    exit_status: ExitError::Other("Cannot find appropriate verification method".into()),
                })
            }
        };
        
        match verify_signature(data, signature, vm) {
            Err(_) => {
                return Err(PrecompileFailure::Error {
                    exit_status: ExitError::Other("Cannot verify signature".into()),
                })
            },
            _ => (),
        };

        let credential_subject = match convert_bech32_address(parsed_payload.vc.credential_subject.address) {
            Ok(addr) => addr,
            Err(_) => {
                return Err(PrecompileFailure::Error {
                    exit_status: ExitError::Other("Cannot convert bech32 address into ethereum".into()),
                })
            }
        };
        Ok((ExitSucceed::Returned, credential_subject))
    }
}

/// Splits provided JWT into header, payload, signature and data.
/// Data field contains concatenated header and payload and can be used for signature verification
fn split_jwt(jwt: &str) -> Result<(String, String, String, String), PrecompileFailure> {
    let parts: Vec<&str> = jwt.split('.').collect();

    if parts.len() == 3 {
        let header = String::from_utf8(base64_decode(parts[0])).unwrap(); // TODO: Remove unwrap
        let payload = String::from_utf8(base64_decode(parts[1])).unwrap(); // TODO: Remove unwrap
        let signature = parts[2].to_string();
        let data = format!("{}.{}", parts[0], parts[1]);

        return Ok((header, payload, signature, data));
    }

    return Err(PrecompileFailure::Error {
        exit_status: ExitError::Other("Wrong amount of parts in JWT".into()),
    })
}

/// Validates JSON-encoded JWT header
fn validate_header(header_json: String) -> Result<(), VerificationError> {
    // Parse and validate header
    let header: Header = match serde_json::from_str(header_json.as_str()) {
        Ok(header) => header,
        Err(err) => {
            return Err(VerificationError::HeaderVerificationError {
                msg: format!("Cannot parse JSON header. Reason: {:?}", err),
            })
        }
    };

    if header.typ != "JWT" {
        return Err(VerificationError::HeaderVerificationError {
            msg: format!("Wrong header type. Expected JWT, Got: {:?}", header.typ),
        });
    }

    if header.alg != "EdDSA" {
        return Err(VerificationError::HeaderVerificationError {
            msg: format!("Wrong alg. Expected EdDSA, Got: {:?}", header.alg),
        });
    }

    Ok(())
}

/// Verifies provided ed25519 signature
fn verify_signature(data: String, signature: String, vm: String) -> Result<(), VerificationError> { // TODO: Replace with PrecompileFailure
    // Construct signature
    let signature = base64_decode(signature.as_str());
    let signature = Signature::from_slice(&signature).map_err(|err| {
        VerificationError::SignatureVerificationError {
            msg: format!("Cannot construct signature. Reason: {}", err),
        }
    })?;

    println!("DEBUG: signature extracted");
    let public_key = multibase_to_hex(vm);
    println!("DEBUG: multibase public key decoded");
    // let public_key =
    //     multibase_to_hex(vm).map_err(|err| VerificationError::SignatureVerificationError {
    //         msg: format!("Cannot decode public key. Reason: {}", err),
    //     })?;

    let public_key: &[u8; 32] = public_key.as_slice().try_into().map_err(|err| {
        VerificationError::SignatureVerificationError {
            msg: format!(
                "Cannot convert public key to fixed byte array. Reason: {}",
                err
            ),
        }
    })?;
    println!("DEBUG: pk extracted");

    let verification_key = VerifyingKey::from_bytes(public_key).map_err(|err| {
        VerificationError::SignatureVerificationError {
            msg: format!("Cannot construct verification key. Reason: {}", err),
        }
    })?;
    println!("DEBUG: verification key constructed");


    // Verify signature
    let data = data.as_bytes();
    verification_key.verify(data, &signature).map_err(|err| {
        VerificationError::SignatureVerificationError {
            msg: format!("Cannot verify signature. Reason: {}", err),
        }
    })
}

fn convert_bech32_address(address: String) -> Result<Vec<u8>, VerificationError> {
    let (_, data, _) =
        bech32::decode(address.as_str()).map_err(|err| VerificationError::ConvertAddressError {
            msg: format!("Cannot decode bech32 address: {}", err),
        })?;
    let data =
        Vec::<u8>::from_base32(&data).map_err(|err| VerificationError::ConvertAddressError {
            msg: format!("Cannot convert base32 to bytes"),
        })?;
    Ok(data)
}

fn base64_decode(input: &str) -> Vec<u8> {
    base64::decode_config(&input, base64::URL_SAFE).unwrap_or_default()
}

fn get_verification_material(connector: *mut GoQuerier, did_url: String) -> Result<Vec<ffi::VerificationMethod>, PrecompileFailure> {
    let encoded_request = coder::encode_verification_methods_request(did_url);
    match ocall::make_request(connector, encoded_request) {
        Some(result) => {
            // Decode protobuf
            let decoded_result = match protobuf::parse_from_bytes::<ffi::QueryVerificationMethodsResponse>(result.as_slice()) {
                Ok(res) => res,
                Err(err) => {
                    return Err(PrecompileFailure::Error {
                        exit_status: ExitError::Other("Cannot decode protobuf response".into()),
                    })
                }
            };
            Ok(decoded_result.vm.to_vec())
        },
        None => {
            return Err(PrecompileFailure::Error {
                exit_status: ExitError::Other("Cannot obtain verification material".into()),
            })
        }
    }
}

fn multibase_to_hex(value: String) -> Vec<u8> {
    let (base, data) = multibase::decode(value).expect("Cannot decode multibase");
    println!("DEBUG: decoded multibase. Result len: {:?}", data.len());
    data[2..].to_vec()
}