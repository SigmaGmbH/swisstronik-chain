use alloc::string::ToString;
use alloc::vec::Vec;
use ethereum::Log;
use evm::backend::{OverlayedChangeSet, RuntimeBaseBackend, RuntimeEnvironment};
use evm::interpreter::error::{ExitError, ExitException, ExitResult, ExitSucceed};
use primitive_types::{H160, H256, U256};
use crate::{coder, querier};
use crate::protobuf_generated::ffi;
use crate::storage::FFIStorage;
use crate::types::{Storage};

pub struct TxEnvironment {
    pub chain_id: U256,
    pub gas_price: U256,
    pub block_number: U256,
    pub timestamp: U256,
    pub block_gas_limit: U256,
    pub block_base_fee_per_gas: U256,
    pub block_coinbase: H160,
}

impl From<ffi::TransactionContext> for TxEnvironment {
    fn from(context: ffi::TransactionContext) -> Self {
        Self {
            chain_id: U256::from(context.chain_id),
            gas_price: U256::from_big_endian(&context.gas_price),
            block_number: U256::from(context.block_number),
            timestamp: U256::from(context.timestamp),
            block_gas_limit: U256::from(context.block_gas_limit),
            block_base_fee_per_gas: U256::from_big_endian(&context.block_base_fee_per_gas),
            block_coinbase: H160::from_slice(&context.block_coinbase),
        }
    }
}

pub struct UpdatedBackend<'state> {
    // We keep GoQuerier to make it accessible for `OCALL` handlers
    pub querier: *mut querier::GoQuerier,
    // Data storage
    pub storage: &'state FFIStorage,
    // Emitted events
    pub logs: Vec<Log>,
    // Transaction context
    pub environment: TxEnvironment,
}

impl<'state> UpdatedBackend<'state> {
    pub fn new(
        querier: *mut querier::GoQuerier,
        storage: &'state FFIStorage,
        environment: TxEnvironment,
    ) -> Self {
        Self {
            querier,
            storage,
            logs: vec![],
            environment,
        }
    }

    pub fn apply_changeset(&self, changeset: &OverlayedChangeSet) -> ExitResult {
        for (address, balance) in changeset.balances.clone() {
            // self.storage.entry(address).or_default().balance = balance;
            // TODO: Add handler for update of account balance
        }

        for (address, nonce) in changeset.nonces.clone() {
            // self.storage.insert(address).or_default().nonce = nonce;
            // TODO: Add handler for update of account nonce
        }

        for (address, code) in changeset.codes.clone() {
            self.storage.insert_account_code(address, code).unwrap()
        }

        for ((address, key), value) in changeset.storages.clone() {
            self.storage.insert_storage_cell(address, key, value).map_err(|err| ExitException::Other(err.to_string().into()))?;
        }

        for address in changeset.deletes.clone() {
            self.storage.remove(&address).map_err(|err| ExitException::Other(err.to_string().into()))?;
        }

        Ok(ExitSucceed::Returned)
    }
}

impl<'state> RuntimeEnvironment for UpdatedBackend<'state> {
    fn block_hash(&self, number: U256) -> H256 {
        let encoded_request = coder::encode_query_block_hash(&number);
        match querier::make_request(self.querier, encoded_request) {
            Some(result) => {
                // Decode protobuf
                let decoded_result = match protobuf::parse_from_bytes::<ffi::QueryBlockHashResponse>(
                    result.as_slice(),
                ) {
                    Ok(res) => res,
                    Err(err) => {
                        println!("Cannot decode protobuf response: {:?}", err);
                        return H256::default();
                    }
                };
                H256::from_slice(decoded_result.hash.as_slice())
            }
            None => {
                println!("Get block hash failed. Empty response");
                H256::default()
            }
        }
    }

    fn block_number(&self) -> U256 {
        self.environment.block_number
    }

    fn block_coinbase(&self) -> H160 {
        self.environment.block_coinbase
    }

    fn block_timestamp(&self) -> U256 {
        self.environment.timestamp
    }

    fn block_difficulty(&self) -> U256 {
        U256::zero()
    }

    fn block_randomness(&self) -> Option<H256> {
        None
    }

    fn block_gas_limit(&self) -> U256 {
        self.environment.block_gas_limit
    }

    fn block_base_fee_per_gas(&self) -> U256 {
        self.environment.block_base_fee_per_gas
    }

    fn chain_id(&self) -> U256 {
        self.environment.chain_id
    }
}

impl<'state> RuntimeBaseBackend for UpdatedBackend<'state> {
    fn balance(&self, address: H160) -> U256 {
        self.storage.get_account(&address).0
    }

    fn code_size(&self, address: H160) -> U256 {
        // TODO: todo!()
        println!("DEBUG: code size req");
        U256::zero()
    }

    fn code_hash(&self, address: H160) -> H256 {
        // TODO: todo!()
        println!("DEBUG: code size req");
        H256::zero()
    }

    fn code(&self, address: H160) -> Vec<u8> {
        self.storage.get_account_code(&address).unwrap_or(Vec::new())
    }

    fn storage(&self, address: H160, index: H256) -> H256 {
        self.storage.get_account_storage_cell(&address, &index).unwrap_or(H256::default())
    }

    fn transient_storage(&self, address: H160, index: H256) -> H256 {
        // Should be implemented by overlayed backend
        H256::zero()
    }

    fn exists(&self, address: H160) -> bool {
        self.storage.contains_key(&address)
    }

    fn nonce(&self, address: H160) -> U256 {
        self.storage.get_account(&address).1
    }
}