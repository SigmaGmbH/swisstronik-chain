use evm::backend::{
    Apply,
    ApplyBackend as EvmApplyBackend,
    Backend as EvmBackend,
    Basic,
    Log
};
use evm::Config;
use primitive_types::{H160, H256, U256};

use std::vec::Vec;

use crate::storage::Storage;

pub static GASOMETER_CONFIG: Config = Config::london();

/// Information required by the evm
#[derive(Clone, Default, PartialEq, Eq)]
pub struct Vicinity {
    pub origin: H160,
    pub nonce: U256,
}

/// Supertrait for our version of EVM Backend
pub trait ExtendedBackend: EvmBackend + EvmApplyBackend {
    fn get_logs(&self) -> Vec<Log>;
}

/// Backend for EVM that allows access to storage
pub struct Backend<'state> {
	// Contains gas price and original sender
    pub vicinity: Vicinity,
	// Accounts state
	pub state: &'state mut dyn Storage,
    // Emitted events
    pub logs: Vec<Log>,
}

impl<'state> ExtendedBackend for Backend<'state> {
    fn get_logs(&self) -> Vec<Log> {
        self.logs.clone()
    }
}

/// Implementation of trait `Backend` provided by evm crate
/// This trait declares readonly functions for the backend
impl<'state> EvmBackend for Backend<'state> {
    fn gas_price(&self) -> U256 {
        U256::zero()
    }

    fn origin(&self) -> H160 {
        self.vicinity.origin
    }

    fn block_hash(&self, _number: U256) -> H256 {
        H256::default()
    }

    fn block_number(&self) -> U256 {
        U256::zero()
    }

    fn block_coinbase(&self) -> H160 {
        H160::default()
    }

    fn block_timestamp(&self) -> U256 {
        U256::zero()
    }

    fn block_difficulty(&self) -> U256 {
        U256::zero()
    }

    fn block_gas_limit(&self) -> U256 {
        U256::max_value()
    }

    fn block_base_fee_per_gas(&self) -> U256 {
        U256::zero()
    }

    fn chain_id(&self) -> U256 {
        U256::one()
    }

    fn original_storage(&self, _address: H160, _index: H256) -> Option<H256> {
        None
    }

    fn block_randomness(&self) -> Option<H256> {
        None
    }

	fn basic(&self, address: H160) -> Basic {
		self.state.get_account(&address)
    }

    fn code(&self, address: H160) -> Vec<u8> {
		self.state
			.get_account_code(&address)
			.unwrap_or_default()
    }

    fn storage(
        &self,
        address: H160,
        index: H256,
    ) -> H256 {
        self.state
            .get_account_storage_cell(&address, &index)
            .unwrap_or_default()
    }

	fn exists(&self, address: H160) -> bool {
        self.state.contains_key(&address)
    }
}

/// Implementation of trait `Apply` provided by evm crate
/// This trait declares write operations for the backend
impl<'state> EvmApplyBackend for Backend<'state> {
	fn apply<A, I, L>(&mut self, values: A, logs: L, _delete_empty: bool)
	where
		A: IntoIterator<Item = Apply<I>>,
		I: IntoIterator<Item = (H256, H256)>,
		L: IntoIterator<Item = Log>,
	{
        let mut total_supply_add = U256::zero();
        let mut total_supply_sub = U256::zero();

		for apply in values {
			match apply {
				Apply::Modify {
					address,
					basic,
					code,
					storage,
                    ..
				} => {
                    // Reset storage is ignored since storage cannot be efficiently reset as this
                    // would require iterating over all of the storage keys

                    // Update account balance and nonce
                    let previous_account_data = self.state.get_account(&address);

                    if basic.balance > previous_account_data.balance {
                        total_supply_add =
                            total_supply_add.checked_add(basic.balance - previous_account_data.balance).unwrap();
                    } else {
                        total_supply_sub =
                            total_supply_sub.checked_add(previous_account_data.balance - basic.balance).unwrap();
                    }
                    self.state.insert_account(address, basic);

                    // Handle contract updates
                    if let Some(code) = code {
                        self.state.insert_account_code(address, code);
                    }

                    // Handle storage updates
                    for (index, value) in storage {
                        if value == H256::default() {
                            self.state.remove_storage_cell(&address, &index);
                        } else {
                            self.state.insert_storage_cell(address, index, value);
                        }
                    }
				},
                // Used by SELFDESTRUCT opcode
				Apply::Delete { address } => {
					self.state.remove(&address);
				}
			}
		}

        // Used to avoid corrupting state via invariant violation
        assert!(
            total_supply_add == total_supply_sub,
            "evm execution would lead to invariant violation ({} != {})",
            total_supply_add,
            total_supply_sub
        );

		for log in logs {
			self.logs.push(log);
		}
	}
}
