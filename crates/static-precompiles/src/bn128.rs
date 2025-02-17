use evm::GasMutState;
#[cfg(feature = "std")]
use std::vec::Vec;

#[cfg(not(feature = "std"))]
use sgx_tstd::vec::Vec;

use evm::interpreter::error::{ExitException, ExitResult, ExitSucceed};
use evm::interpreter::runtime::RuntimeState;
use primitive_types::U256;
use substrate_bn as bn;

use crate::{LinearCostPrecompile, Precompile};

fn read_fr(input: &[u8], start_inx: usize) -> Result<bn::Fr, ExitException> {
	bn::Fr::from_slice(&input[start_inx..(start_inx + 32)]).map_err(|_| ExitException::Other("Invalid field element".into()))
}

fn read_point(input: &[u8], start_inx: usize) -> Result<bn::G1, ExitException> {
	use bn::{Fq, AffineG1, G1, Group};

	let px = Fq::from_slice(&input[start_inx..(start_inx + 32)]).map_err(|_| ExitException::Other("Invalid point x coordinate".into()))?;
	let py = Fq::from_slice(&input[(start_inx + 32)..(start_inx + 64)]).map_err(|_| ExitException::Other("Invalid point y coordinate".into()))?;
	Ok(
		if px == Fq::zero() && py == Fq::zero() {
			G1::zero()
		} else {
			AffineG1::new(px, py).map_err(|_| ExitException::Other("Invalid curve point".into()))?.into()
		}
	)
}

/// The Bn128Add builtin
pub struct Bn128Add;

impl LinearCostPrecompile for Bn128Add {
	const BASE: u64 = 150;
	const WORD: u64 = 0;

	fn raw_execute(
		input: &[u8],
		_: u64,
	) -> (ExitResult, Vec<u8>) {
		use bn::AffineG1;

		let p1 = match read_point(input, 0) {
			Ok(p) => p,
			Err(err) => return (err.into(), Vec::new())
		};
		let p2 = match read_point(input, 64) {
			Ok(p) => p,
			Err(err) => return (err.into(), Vec::new())
		};

		let mut buf = [0u8; 64];
		if let Some(sum) = AffineG1::from_jacobian(p1 + p2) {
			// point not at infinity
			match sum.x().to_big_endian(&mut buf[0..32]) {
				Ok(_) => {},
				Err(_) => return (ExitException::Other("Cannot fail since 0..32 is 32-byte length".into()).into(), Vec::new())
			}

			match sum.y().to_big_endian(&mut buf[32..64]) {
				Ok(_) => {},
				Err(_) => return (ExitException::Other("Cannot fail since 32..64 is 32-byte length".into()).into(), Vec::new())
			}
		}

		(ExitSucceed::Returned.into(), buf.to_vec())
	}
}

/// The Bn128Mul builtin
pub struct Bn128Mul;

impl LinearCostPrecompile for Bn128Mul {
	const BASE: u64 = 6000;
	const WORD: u64 = 0;

	fn raw_execute(
		input: &[u8],
		_: u64,
	) -> (ExitResult, Vec<u8>) {
		use bn::AffineG1;

		let p = match read_point(input, 0) {
			Ok(p) => p,
			Err(err) => return (err.into(), Vec::new())
		};
		let fr = match read_fr(input, 64) {
			Ok(fr) => fr,
			Err(err) => return (err.into(), Vec::new())
		};

		let mut buf = [0u8; 64];
		if let Some(sum) = AffineG1::from_jacobian(p * fr) {
			// point not at infinity
			match sum.x().to_big_endian(&mut buf[0..32]) {
				Ok(_) => {},
				Err(_) => return (ExitException::Other("Cannot fail since 0..32 is 32-byte length".into()).into(), Vec::new())
			}

			match sum.y().to_big_endian(&mut buf[32..64]) {
				Ok(_) => {},
				Err(_) => return (ExitException::Other("Cannot fail since 0..32 is 32-byte length".into()).into(), Vec::new())
			}
		}

		(ExitSucceed::Returned.into(), buf.to_vec())
	}
}

/// The Bn128Pairing builtin
pub struct Bn128Pairing;

const BN_128_PAIRING_GAS_COST_PER_ELEMENT: u64 = 34000;
const BN_128_PAIRING_MIN_GAS_COST: u64 = 45000;

impl<G: AsRef<RuntimeState> + GasMutState> Precompile<G> for Bn128Pairing {
	fn execute(
		input: &[u8],
		gasometer: &mut G,
	) -> (ExitResult, Vec<u8>) {
		use bn::{AffineG1, AffineG2, Fq, Fq2, pairing_batch, G1, G2, Gt, Group};

		let ret_val = if input.is_empty() {
			if let Err(e) = gasometer.record_gas(U256::from(BN_128_PAIRING_MIN_GAS_COST)) {
				return (e.into(), Vec::new());
			}

			U256::one()
		} else {
			// (a, b_a, b_b - each 64-byte affine coordinates)
			let elements = input.len() / 192;

			let gas_cost_per_element = U256::from(BN_128_PAIRING_GAS_COST_PER_ELEMENT);
			let gas_cost = U256::from(BN_128_PAIRING_GAS_COST_PER_ELEMENT)
				.saturating_add(gas_cost_per_element * elements);

			if let Err(e) = gasometer.record_gas(gas_cost) {
				return (e.into(), Vec::new());
			}

			let mut vals = Vec::new();
			for idx in 0..elements {
				let a_x = match Fq::from_slice(&input[idx*192..idx*192+32]) {
					Ok(a) => a,
					Err(_) => return (ExitException::Other("Invalid a argument x coordinate".into()).into(), Vec::new())
				};

				let a_y = match Fq::from_slice(&input[idx*192+32..idx*192+64]) {
					Ok(a) => a,
					Err(_) => return (ExitException::Other("Invalid a argument y coordinate".into()).into(), Vec::new())
				};

				let b_a_y = match Fq::from_slice(&input[idx*192+64..idx*192+96]) {
					Ok(b) => b,
					Err(_) => return (ExitException::Other("Invalid b argument imaginary coeff x coordinate".into()).into(), Vec::new())
				};

				let b_a_x = match Fq::from_slice(&input[idx*192+96..idx*192+128]) {
					Ok(b) => b,
					Err(_) => return (ExitException::Other("Invalid b argument imaginary coeff y coordinate".into()).into(), Vec::new())
				};

				let b_b_y = match Fq::from_slice(&input[idx*192+128..idx*192+160]) {
					Ok(b) => b,
					Err(_) => return (ExitException::Other("Invalid b argument real coeff x coordinate".into()).into(), Vec::new())
				};

				let b_b_x = match Fq::from_slice(&input[idx*192+160..idx*192+192]) {
					Ok(b) => b,
					Err(_) => return (ExitException::Other("Invalid b argument real coeff y coordinate".into()).into(), Vec::new())
				};

				let b_a = Fq2::new(b_a_x, b_a_y);
				let b_b = Fq2::new(b_b_x, b_b_y);
				let b = if b_a.is_zero() && b_b.is_zero() {
					G2::zero()
				} else {
					let a_g2 = match AffineG2::new(b_a, b_b) {
						Ok(a) => a,
						Err(_) => return (ExitException::Other("Invalid b argument - not on curve".into()).into(), Vec::new())
					};

					G2::from(a_g2)
				};
				let a = if a_x.is_zero() && a_y.is_zero() {
					G1::zero()
				} else {
					let a_g1 = match AffineG1::new(a_x, a_y) {
						Ok(a) => a,
						Err(_) => return (ExitException::Other("Invalid a argumant - not on curve".into()).into(), Vec::new())
					};
					G1::from(a_g1)
				};
				vals.push((a, b));
			};

			let mul = pairing_batch(&vals);

			if mul == Gt::one() {
				U256::one()
			} else {
				U256::zero()
			}
		};

		let mut buf = [0u8; 32];
		ret_val.to_big_endian(&mut buf);

		(ExitSucceed::Returned.into(), buf.to_vec())
	}
}
