extern crate sgx_tstd as std;

use evm::GasMutState;
use evm::interpreter::error::ExitResult;
use evm::interpreter::runtime::RuntimeState;
use evm::standard::PrecompileSet;
use std::vec::Vec;
use primitive_types::H160;

use static_precompiles::{secp256r1, modexp, ripemd160, ec_recover, sha3fips, sha256, datacopy, bn128, blake2f, curve25519};
use static_precompiles::Precompile;

use crate::GoQuerier;

mod compliance_bridge;

/// Precompile with possibility to interact with Cosmos side using GoQuerier
pub trait LinearCostPrecompileWithQuerier<G> {
    const BASE: u64;
    const WORD: u64;

    fn execute(querier: *mut GoQuerier, input: &[u8], gasometer: &mut G) -> (ExitResult, Vec<u8>);
}

pub struct EVMPrecompiles {
    querier: *mut GoQuerier,
}

impl EVMPrecompiles {
    pub fn new(querier: *mut GoQuerier) -> Self {
        Self{ querier }
    }
}

impl<G: AsRef<RuntimeState> + GasMutState, H> PrecompileSet<G, H> for EVMPrecompiles {
    fn execute(&self, code_address: H160, input: &[u8], gasometer: &mut G, _handler: &mut H) -> Option<(ExitResult, Vec<u8>)> {
        match code_address {
            // Ethereum precompiles:
            a if a == hash(1) => Some(ec_recover::ECRecover::execute(input, gasometer)),
            a if a == hash(2) => Some(sha256::Sha256::execute(input, gasometer)),
            a if a == hash(3) => Some(ripemd160::Ripemd160::execute(input, gasometer)),
            a if a == hash(4) => Some(datacopy::DataCopy::execute(input, gasometer)),
            a if a == hash(5) => Some(modexp::Modexp::execute(input, gasometer)),
            a if a == hash(6) => Some(bn128::Bn128Add::execute(input, gasometer)),
            a if a == hash(7) => Some(bn128::Bn128Mul::execute(input, gasometer)),
            a if a == hash(8) => Some(bn128::Bn128Pairing::execute(input, gasometer)),
            a if a == hash(9) => Some(blake2f::Blake2F::execute(input, gasometer)),
            // RIP-7212
            a if a == hash(0x100) => Some(secp256r1::P256Verify::execute(input, gasometer)),
            // Non-Frontier specific nor Ethereum precompiles :
            a if a == hash(1024) => Some(sha3fips::Sha3FIPS256::execute(input, gasometer)),
            a if a == hash(1025) => Some(sha3fips::Sha3FIPS512::execute(input, gasometer)),
            a if a == hash(1028) => Some(compliance_bridge::ComplianceBridge::execute(self.querier, input, gasometer)),
            a if a == hash(1029) => Some(curve25519::Curve25519Add::execute(input, gasometer)),
            a if a == hash(1030) => Some(curve25519::Curve25519ScalarMul::execute(input, gasometer)),
            a if a == hash(1031) => Some(curve25519::Ed25519Verify::execute(input, gasometer)),
            _ => None,
        }
    }
}

#[inline]
fn hash(a: u64) -> H160 {
    H160::from_low_u64_be(a)
}
