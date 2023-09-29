#![no_std]

extern crate alloc;
extern crate sgx_tstd as std;

use alloc::vec::Vec;
use precompile_std::{
    ExitError, ExitSucceed, LinearCostPrecompile, PrecompileFailure,
};
use rlp::Rlp;

/// The identity precompile.
pub struct Identity;

impl LinearCostPrecompile for Identity {
    const BASE: u64 = 60;
    const WORD: u64 = 150;

    fn execute(input: &[u8], _: u64) -> Result<(ExitSucceed, Vec<u8>), PrecompileFailure> {
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