use core::marker::PhantomData;

use evm_precompile_blake2f::Blake2F;
use evm_precompile_bn128::{Bn128Add, Bn128Mul, Bn128Pairing};
use evm_precompile_modexp::Modexp;
// use evm_precompile_curve25519::{Curve25519Add, Curve25519ScalarMul};
use evm_precompile_simple::{ECRecover, Identity, Ripemd160, Sha256};
use precompile_std::{Precompile, PrecompileHandle, PrecompileResult, PrecompileSet, IsPrecompileResult};
use primitive_types::H160;

// use evm_precompile_sha3fips::{Sha3FIPS256, Sha3FIPS512};

pub struct EVMPrecompiles<R>(PhantomData<R>);

impl<R> EVMPrecompiles<R>
    where
        R: evm::backend::Backend,
{
    pub fn new() -> Self {
        Self(Default::default())
    }
    pub fn used_addresses() -> [H160; 9] {
        [
            hash(1),
            hash(2),
            hash(3),
            hash(4),
            hash(5),
            hash(6),
            hash(7),
            hash(8),
            hash(9),
            // hash(1024),
            // hash(1025),
        ]
    }
}
impl<R> PrecompileSet for EVMPrecompiles<R>
    where
        R: evm::backend::Backend,
{
    fn execute(&self, handle: &mut impl PrecompileHandle) -> Option<PrecompileResult> {
        match handle.code_address() {
            // Ethereum precompiles:
            a if a == hash(1) => Some(ECRecover::execute(handle)),
            a if a == hash(2) => Some(Sha256::execute(handle)),
            a if a == hash(3) => Some(Ripemd160::execute(handle)),
            a if a == hash(4) => Some(Identity::execute(handle)),
            a if a == hash(5) => Some(Modexp::execute(handle)),
            a if a == hash(6) => Some(Bn128Add::execute(handle)),
            a if a == hash(7) => Some(Bn128Mul::execute(handle)),
            a if a == hash(8) => Some(Bn128Pairing::execute(handle)),
            a if a == hash(9) => Some(Blake2F::execute(handle)),
            // Non-Frontier specific nor Ethereum precompiles :
            // a if a == hash(1024) => Some(Sha3FIPS256::execute(handle)),
            // a if a == hash(1025) => Some(Sha3FIPS512::execute(handle)),
            // a if a == hash(1026) => Some(ECRecoverPublicKey::execute(handle)),
            _ => None,
        }
    }

    fn is_precompile(&self, address: H160, _gas: u64) -> IsPrecompileResult {
		IsPrecompileResult::Answer {
			is_precompile: Self::used_addresses().contains(&address),
			extra_cost: 0,
		}
    }
}

#[inline]
fn hash(a: u64) -> H160 {
    H160::from_low_u64_be(a)
}
