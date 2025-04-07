#[cfg(feature = "std")]
use std::vec::Vec;

#[cfg(not(feature = "std"))]
use sgx_tstd::vec::Vec;

use evm::interpreter::error::{ExitResult, ExitSucceed};
use crate::LinearCostPrecompile;

/// The DataCopy precompile.
pub struct DataCopy;

impl LinearCostPrecompile for DataCopy {
	const BASE: u64 = 15;
	const WORD: u64 = 3;

	fn raw_execute(input: &[u8], _: u64) -> (ExitResult, Vec<u8>) {
		(ExitSucceed::Returned.into(), input.to_vec())
	}
}